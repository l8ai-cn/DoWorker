package imbridge

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	domain "github.com/anthropics/agentsmesh/backend/internal/domain/imbridge"
)

type httpJSON struct {
	HTTP *http.Client
}

func (h httpJSON) client() *http.Client {
	if h.HTTP != nil {
		return h.HTTP
	}
	return http.DefaultClient
}

func doJSONRequest(ctx context.Context, c *http.Client, method, rawURL string, headers map[string]string, body any, out any) error {
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		rdr = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, rawURL, rdr)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("http %d: %s", resp.StatusCode, string(data))
	}
	if out == nil {
		return nil
	}
	return json.Unmarshal(data, out)
}

// --- Feishu ---

type FeishuProvider struct{ HTTP *http.Client }

func (p *FeishuProvider) Type() string        { return domain.ProviderFeishu }
func (p *FeishuProvider) DisplayName() string { return "飞书" }

type feishuBridgeConfig struct {
	AppID             string `json:"app_id"`
	AppSecret         string `json:"app_secret"`
	VerificationToken string `json:"verification_token"`
	EncryptKey        string `json:"encrypt_key,omitempty"`
	DefaultChatID     string `json:"default_chat_id,omitempty"`
}

func (p *FeishuProvider) ValidateConfig(raw json.RawMessage) error {
	var cfg feishuBridgeConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return err
	}
	if cfg.AppID == "" || cfg.AppSecret == "" || cfg.VerificationToken == "" {
		return errors.New("feishu requires app_id, app_secret, verification_token")
	}
	return nil
}

func (p *FeishuProvider) VerifyWebhook(_ context.Context, raw json.RawMessage, _ http.Header, body []byte) error {
	var cfg feishuBridgeConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return err
	}
	var envelope struct {
		Token string `json:"token"`
		Type  string `json:"type"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return err
	}
	if envelope.Token != "" && envelope.Token != cfg.VerificationToken {
		return errors.New("feishu verification token mismatch")
	}
	return nil
}

func (p *FeishuProvider) ParseInbound(_ context.Context, raw json.RawMessage, _ http.Header, body []byte) (*InboundEvent, error) {
	var envelope struct {
		Challenge string `json:"challenge"`
		Type      string `json:"type"`
		Header    struct {
			EventType string `json:"event_type"`
		} `json:"header"`
		Event struct {
			Message struct {
				ChatID  string `json:"chat_id"`
				Content string `json:"content"`
			} `json:"message"`
			Sender struct {
				SenderID struct {
					OpenID string `json:"open_id"`
				} `json:"sender_id"`
			} `json:"sender"`
		} `json:"event"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, err
	}
	if envelope.Challenge != "" {
		return &InboundEvent{Challenge: envelope.Challenge}, nil
	}
	text := envelope.Event.Message.Content
	var content struct {
		Text string `json:"text"`
	}
	_ = json.Unmarshal([]byte(text), &content)
	if content.Text != "" {
		text = content.Text
	}
	return &InboundEvent{
		ExternalThreadID: envelope.Event.Message.ChatID,
		SenderName:       envelope.Event.Sender.SenderID.OpenID,
		Text:             strings.TrimSpace(text),
	}, nil
}

func (p *FeishuProvider) SendOutbound(ctx context.Context, raw json.RawMessage, msg OutboundMessage) error {
	var cfg feishuBridgeConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return err
	}
	token, err := p.tenantToken(ctx, cfg)
	if err != nil {
		return err
	}
	chatID := msg.ExternalThreadID
	if chatID == "" {
		chatID = cfg.DefaultChatID
	}
	payload := map[string]any{
		"receive_id": chatID,
		"msg_type":   "text",
		"content":    map[string]string{"text": msg.Text},
	}
	return doJSONRequest(ctx, p.client(), http.MethodPost,
		"https://open.feishu.cn/open-apis/im/v1/messages?receive_id_type=chat_id",
		map[string]string{"Authorization": "Bearer " + token}, payload, nil)
}

func (p *FeishuProvider) tenantToken(ctx context.Context, cfg feishuBridgeConfig) (string, error) {
	var out struct {
		Code              int    `json:"code"`
		TenantAccessToken string `json:"tenant_access_token"`
	}
	err := doJSONRequest(ctx, p.client(), http.MethodPost,
		"https://open.feishu.cn/open-apis/auth/v3/tenant_access_token/internal",
		nil, map[string]string{"app_id": cfg.AppID, "app_secret": cfg.AppSecret}, &out)
	if err != nil {
		return "", err
	}
	if out.Code != 0 || out.TenantAccessToken == "" {
		return "", fmt.Errorf("feishu auth failed code=%d", out.Code)
	}
	return out.TenantAccessToken, nil
}

func (p *FeishuProvider) client() *http.Client { return (&httpJSON{HTTP: p.HTTP}).client() }

// --- DingTalk ---

type DingTalkProvider struct{ HTTP *http.Client }

func (p *DingTalkProvider) Type() string        { return domain.ProviderDingTalk }
func (p *DingTalkProvider) DisplayName() string { return "钉钉" }

type dingTalkBridgeConfig struct {
	AppKey      string `json:"app_key"`
	AppSecret   string `json:"app_secret"`
	SigningSecret string `json:"signing_secret"`
	WebhookURL  string `json:"webhook_url,omitempty"`
}

func (p *DingTalkProvider) ValidateConfig(raw json.RawMessage) error {
	var cfg dingTalkBridgeConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return err
	}
	if cfg.AppKey == "" || cfg.AppSecret == "" {
		return errors.New("dingtalk requires app_key and app_secret")
	}
	return nil
}

func (p *DingTalkProvider) VerifyWebhook(_ context.Context, raw json.RawMessage, headers http.Header, body []byte) error {
	var cfg dingTalkBridgeConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return err
	}
	if cfg.SigningSecret == "" {
		return nil
	}
	ts := headers.Get("timestamp")
	sign := headers.Get("sign")
	if ts == "" || sign == "" {
		return errors.New("dingtalk missing timestamp/sign headers")
	}
	mac := hmac.New(sha256.New, []byte(cfg.SigningSecret))
	mac.Write([]byte(ts + "\n" + cfg.SigningSecret))
	expected := hex.EncodeToString(mac.Sum(nil))
	if sign != expected {
		return errors.New("dingtalk signature mismatch")
	}
	return nil
}

func (p *DingTalkProvider) ParseInbound(_ context.Context, _ json.RawMessage, _ http.Header, body []byte) (*InboundEvent, error) {
	var payload struct {
		ConversationID string `json:"conversationId"`
		ConversationType string `json:"conversationType"`
		Text           struct {
			Content string `json:"content"`
		} `json:"text"`
		SenderNick string `json:"senderNick"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	thread := payload.ConversationID
	if thread == "" {
		thread = payload.ConversationType
	}
	return &InboundEvent{
		ExternalThreadID: thread,
		SenderName:       payload.SenderNick,
		Text:             strings.TrimSpace(payload.Text.Content),
	}, nil
}

func (p *DingTalkProvider) SendOutbound(ctx context.Context, raw json.RawMessage, msg OutboundMessage) error {
	var cfg dingTalkBridgeConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return err
	}
	if cfg.WebhookURL != "" {
		return doJSONRequest(ctx, p.client(), http.MethodPost, cfg.WebhookURL, nil, map[string]any{
			"msgtype": "text",
			"text":    map[string]string{"content": msg.Text},
		}, nil)
	}
	token, err := p.accessToken(ctx, cfg)
	if err != nil {
		return err
	}
	return doJSONRequest(ctx, p.client(), http.MethodPost,
		"https://api.dingtalk.com/v1.0/robot/oToMessages/batchSend",
		map[string]string{"x-acs-dingtalk-access-token": token},
		map[string]any{
			"robotCode": cfg.AppKey,
			"userIds":   []string{msg.ExternalThreadID},
			"msgKey":    "sampleText",
			"msgParam":  map[string]string{"content": msg.Text},
		}, nil)
}

func (p *DingTalkProvider) accessToken(ctx context.Context, cfg dingTalkBridgeConfig) (string, error) {
	var out struct {
		AccessToken string `json:"accessToken"`
	}
	const u = "https://api.dingtalk.com/v1.0/oauth2/accessToken"
	err := doJSONRequest(ctx, p.client(), http.MethodPost, u, nil, map[string]string{
		"appKey": cfg.AppKey, "appSecret": cfg.AppSecret,
	}, &out)
	if err != nil {
		return "", err
	}
	return out.AccessToken, nil
}

func (p *DingTalkProvider) client() *http.Client { return (&httpJSON{HTTP: p.HTTP}).client() }

// --- WeCom ---

type WeComProvider struct{ HTTP *http.Client }

func (p *WeComProvider) Type() string        { return domain.ProviderWeCom }
func (p *WeComProvider) DisplayName() string { return "企业微信" }

type weComBridgeConfig struct {
	CorpID         string `json:"corp_id"`
	CorpSecret     string `json:"corp_secret"`
	Token          string `json:"token"`
	EncodingAESKey string `json:"encoding_aes_key,omitempty"`
	AgentID        int64  `json:"agent_id"`
}

func (p *WeComProvider) ValidateConfig(raw json.RawMessage) error {
	var cfg weComBridgeConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return err
	}
	if cfg.CorpID == "" || cfg.CorpSecret == "" || cfg.Token == "" || cfg.AgentID == 0 {
		return errors.New("wecom requires corp_id, corp_secret, token, agent_id")
	}
	return nil
}

func (p *WeComProvider) VerifyWebhook(_ context.Context, raw json.RawMessage, _ http.Header, _ []byte) error {
	return nil
}

func (p *WeComProvider) ParseInbound(_ context.Context, _ json.RawMessage, _ http.Header, body []byte) (*InboundEvent, error) {
	var payload struct {
		MsgType string `json:"MsgType"`
		Content string `json:"Content"`
		FromUserName string `json:"FromUserName"`
		ChatID  string `json:"ChatId"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	thread := payload.ChatID
	if thread == "" {
		thread = payload.FromUserName
	}
	return &InboundEvent{
		ExternalThreadID: thread,
		SenderName:       payload.FromUserName,
		Text:             strings.TrimSpace(payload.Content),
	}, nil
}

func (p *WeComProvider) SendOutbound(ctx context.Context, raw json.RawMessage, msg OutboundMessage) error {
	var cfg weComBridgeConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return err
	}
	token, err := p.accessToken(ctx, cfg)
	if err != nil {
		return err
	}
	u := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=%s", url.QueryEscape(token))
	return doJSONRequest(ctx, p.client(), http.MethodPost, u, nil, map[string]any{
		"touser":  msg.ExternalThreadID,
		"msgtype": "text",
		"agentid": cfg.AgentID,
		"text":    map[string]string{"content": msg.Text},
	}, nil)
}

func (p *WeComProvider) accessToken(ctx context.Context, cfg weComBridgeConfig) (string, error) {
	var out struct {
		AccessToken string `json:"access_token"`
		ErrCode     int    `json:"errcode"`
	}
	u := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=%s&corpsecret=%s",
		url.QueryEscape(cfg.CorpID), url.QueryEscape(cfg.CorpSecret))
	if err := doJSONRequest(ctx, p.client(), http.MethodGet, u, nil, nil, &out); err != nil {
		return "", err
	}
	if out.ErrCode != 0 || out.AccessToken == "" {
		return "", fmt.Errorf("wecom token errcode=%d", out.ErrCode)
	}
	return out.AccessToken, nil
}

func (p *WeComProvider) client() *http.Client { return (&httpJSON{HTTP: p.HTTP}).client() }

// --- Slack ---

type SlackProvider struct{ HTTP *http.Client }

func (p *SlackProvider) Type() string        { return domain.ProviderSlack }
func (p *SlackProvider) DisplayName() string { return "Slack" }

type slackBridgeConfig struct {
	SigningSecret string `json:"signing_secret"`
	BotToken      string `json:"bot_token"`
	DefaultChannel string `json:"default_channel,omitempty"`
}

func (p *SlackProvider) ValidateConfig(raw json.RawMessage) error {
	var cfg slackBridgeConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return err
	}
	if cfg.SigningSecret == "" || cfg.BotToken == "" {
		return errors.New("slack requires signing_secret and bot_token")
	}
	return nil
}

func (p *SlackProvider) VerifyWebhook(_ context.Context, raw json.RawMessage, headers http.Header, body []byte) error {
	var cfg slackBridgeConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return err
	}
	ts := headers.Get("X-Slack-Request-Timestamp")
	sig := headers.Get("X-Slack-Signature")
	if ts == "" || sig == "" {
		return nil
	}
	if age := time.Since(time.Unix(parseInt64(ts), 0)); age > 5*time.Minute || age < -5*time.Minute {
		return errors.New("slack timestamp out of range")
	}
	base := "v0:" + ts + ":" + string(body)
	mac := hmac.New(sha256.New, []byte(cfg.SigningSecret))
	mac.Write([]byte(base))
	expected := "v0=" + hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(sig)) {
		return errors.New("slack signature mismatch")
	}
	return nil
}

func (p *SlackProvider) ParseInbound(_ context.Context, raw json.RawMessage, _ http.Header, body []byte) (*InboundEvent, error) {
	var payload struct {
		Type      string `json:"type"`
		Challenge string `json:"challenge"`
		Event     struct {
			Type    string `json:"type"`
			User    string `json:"user"`
			Text    string `json:"text"`
			Channel string `json:"channel"`
		} `json:"event"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	if payload.Challenge != "" {
		return &InboundEvent{Challenge: payload.Challenge}, nil
	}
	return &InboundEvent{
		ExternalThreadID: payload.Event.Channel,
		SenderName:       payload.Event.User,
		Text:             strings.TrimSpace(payload.Event.Text),
	}, nil
}

func (p *SlackProvider) SendOutbound(ctx context.Context, raw json.RawMessage, msg OutboundMessage) error {
	var cfg slackBridgeConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return err
	}
	channel := msg.ExternalThreadID
	if channel == "" {
		channel = cfg.DefaultChannel
	}
	return doJSONRequest(ctx, p.client(), http.MethodPost, "https://slack.com/api/chat.postMessage",
		map[string]string{"Authorization": "Bearer " + cfg.BotToken},
		map[string]any{"channel": channel, "text": msg.Text}, nil)
}

func (p *SlackProvider) client() *http.Client { return (&httpJSON{HTTP: p.HTTP}).client() }

func parseInt64(s string) int64 {
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}
