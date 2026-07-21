package imbridge

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// iLink API constants aligned with OpenClaw @tencent-weixin/openclaw-weixin and doagent wechat.rs.
const (
	defaultIlinkBaseURL      = "https://ilinkai.weixin.qq.com"
	ilinkAppID               = "bot"
	ilinkAppClientVersion    = (2 << 16) | (2 << 8)
	ilinkAuthType            = "ilink_bot_token"
	epGetBotQR               = "ilink/bot/get_bot_qrcode"
	epGetQRStatus            = "ilink/bot/get_qrcode_status"
	epGetUpdates             = "getupdates"
	epSendMessage            = "sendmessage"
	ilinkPollTimeout         = 35 * time.Second
)

type weixinBridgeConfig struct {
	AccountID      string `json:"account_id"`
	BotToken       string `json:"bot_token"`
	BaseURL        string `json:"base_url,omitempty"`
	UserID         string `json:"user_id,omitempty"`
	BotAgent       string `json:"bot_agent,omitempty"`
	GetUpdatesBuf  string `json:"get_updates_buf,omitempty"`
}

func parseWeixinConfig(raw json.RawMessage) (weixinBridgeConfig, error) {
	var cfg weixinBridgeConfig
	if len(raw) == 0 {
		return cfg, nil
	}
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return cfg, err
	}
	if strings.TrimSpace(cfg.BaseURL) == "" {
		cfg.BaseURL = defaultIlinkBaseURL
	}
	return cfg, nil
}

func (cfg weixinBridgeConfig) baseURL() string {
	if strings.TrimSpace(cfg.BaseURL) == "" {
		return defaultIlinkBaseURL
	}
	return strings.TrimRight(cfg.BaseURL, "/")
}

type ilinkClient struct {
	http *http.Client
}

func newIlinkClient(httpClient *http.Client) *ilinkClient {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &ilinkClient{http: httpClient}
}

func randomWechatUIN() string {
	var buf [4]byte
	_, _ = rand.Read(buf[:])
	n := uint32(buf[0])<<24 | uint32(buf[1])<<16 | uint32(buf[2])<<8 | uint32(buf[3])
	return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", n)))
}

func (c *ilinkClient) headers(bodyLen *int, token string) http.Header {
	h := make(http.Header)
	h.Set("iLink-App-Id", ilinkAppID)
	h.Set("iLink-App-ClientVersion", fmt.Sprintf("%d", ilinkAppClientVersion))
	h.Set("X-WECHAT-UIN", randomWechatUIN())
	if bodyLen != nil {
		h.Set("Content-Type", "application/json")
	}
	if token != "" {
		h.Set("AuthorizationType", ilinkAuthType)
		h.Set("Authorization", "Bearer "+token)
	}
	if agent := strings.TrimSpace("AgentCloud/1.0"); agent != "" {
		h.Set("bot_agent", agent)
	}
	return h
}

func (c *ilinkClient) getJSON(ctx context.Context, baseURL, endpoint string) (map[string]any, error) {
	url := baseURL + "/" + strings.TrimPrefix(endpoint, "/")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header = c.headers(nil, "")
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("ilink GET %s: HTTP %d: %s", endpoint, resp.StatusCode, string(body))
	}
	var out map[string]any
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ilinkClient) postJSON(ctx context.Context, baseURL, endpoint, token string, payload any) (map[string]any, error) {
	url := baseURL + "/" + strings.TrimPrefix(endpoint, "/")
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	n := len(b)
	req.Header = c.headers(&n, token)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("ilink POST %s: HTTP %d: %s", endpoint, resp.StatusCode, string(body))
	}
	var out map[string]any
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	return out, nil
}

type weixinQRResult struct {
	QRCodeValue string
	QRCodeURL   string
}

func (c *ilinkClient) fetchBotQR(ctx context.Context, botType string) (*weixinQRResult, error) {
	endpoint := fmt.Sprintf("%s?bot_type=%s", epGetBotQR, botType)
	payload, err := c.getJSON(ctx, defaultIlinkBaseURL, endpoint)
	if err != nil {
		return nil, err
	}
	value, _ := payload["qrcode"].(string)
	url, _ := payload["qrcode_img_content"].(string)
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, fmt.Errorf("ilink QR response missing qrcode")
	}
	return &weixinQRResult{QRCodeValue: value, QRCodeURL: normalizeQRImageSrc(url)}, nil
}

func (c *ilinkClient) pollQRStatus(ctx context.Context, baseURL, qrcodeValue string) (map[string]any, error) {
	endpoint := fmt.Sprintf("%s?qrcode=%s", epGetQRStatus, qrcodeValue)
	return c.getJSON(ctx, baseURL, endpoint)
}

type weixinUpdatesResult struct {
	Messages       []map[string]any
	GetUpdatesBuf  string
	LongPollMS     int
	Ret            int
	ErrCode        int
	ErrMsg         string
}

func (c *ilinkClient) getUpdates(ctx context.Context, cfg weixinBridgeConfig) (*weixinUpdatesResult, error) {
	ctx, cancel := context.WithTimeout(ctx, ilinkPollTimeout+5*time.Second)
	defer cancel()
	payload, err := c.postJSON(ctx, cfg.baseURL(), epGetUpdates, cfg.BotToken, map[string]string{
		"get_updates_buf": cfg.GetUpdatesBuf,
	})
	if err != nil {
		return nil, err
	}
	out := &weixinUpdatesResult{}
	if v, ok := payload["ret"].(float64); ok {
		out.Ret = int(v)
	}
	if v, ok := payload["errcode"].(float64); ok {
		out.ErrCode = int(v)
	}
	if v, ok := payload["errmsg"].(string); ok {
		out.ErrMsg = v
	}
	if v, ok := payload["get_updates_buf"].(string); ok {
		out.GetUpdatesBuf = v
	}
	if v, ok := payload["longpolling_timeout_ms"].(float64); ok {
		out.LongPollMS = int(v)
	}
	if raw, ok := payload["msgs"].([]any); ok {
		for _, item := range raw {
			if m, ok := item.(map[string]any); ok {
				out.Messages = append(out.Messages, m)
			}
		}
	}
	return out, nil
}

func (c *ilinkClient) sendText(ctx context.Context, cfg weixinBridgeConfig, toUserID, contextToken, text string) error {
	if strings.TrimSpace(cfg.BotAgent) != "" {
		// bot_agent is sent as header in OpenClaw; optional per-connection override stored in config.
	}
	_, err := c.postJSON(ctx, cfg.baseURL(), epSendMessage, cfg.BotToken, map[string]any{
		"msg": map[string]any{
			"to_user_id":     toUserID,
			"context_token":  contextToken,
			"item_list": []map[string]any{
				{
					"type":       1,
					"text_item": map[string]string{"text": text},
				},
			},
		},
	})
	return err
}

func parseWeixinInbound(msg map[string]any) *InboundEvent {
	msgType, _ := msg["message_type"].(float64)
	if int(msgType) == 2 {
		return nil // skip bot-originated
	}
	fromUser, _ := msg["from_user_id"].(string)
	sessionID, _ := msg["session_id"].(string)
	contextToken, _ := msg["context_token"].(string)
	thread := strings.TrimSpace(sessionID)
	if thread == "" {
		thread = strings.TrimSpace(fromUser)
	}
	text := extractWeixinText(msg)
	if strings.TrimSpace(text) == "" {
		return nil
	}
	return &InboundEvent{
		ExternalThreadID: thread,
		SenderName:       fromUser,
		Text:             strings.TrimSpace(text),
		ContextToken:     contextToken,
	}
}

func extractWeixinText(msg map[string]any) string {
	items, ok := msg["item_list"].([]any)
	if !ok {
		return ""
	}
	var parts []string
	for _, raw := range items {
		item, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		t, _ := item["type"].(float64)
		if int(t) != 1 {
			continue
		}
		if textItem, ok := item["text_item"].(map[string]any); ok {
			if text, ok := textItem["text"].(string); ok && strings.TrimSpace(text) != "" {
				parts = append(parts, text)
			}
		}
	}
	return strings.Join(parts, "\n")
}

func normalizeQRImageSrc(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if strings.HasPrefix(raw, "data:image/") || strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		return raw
	}
	return "data:image/png;base64," + raw
}

func mergeWeixinConfig(raw json.RawMessage, patch weixinBridgeConfig) (json.RawMessage, error) {
	cfg, err := parseWeixinConfig(raw)
	if err != nil {
		return nil, err
	}
	if patch.AccountID != "" {
		cfg.AccountID = patch.AccountID
	}
	if patch.BotToken != "" {
		cfg.BotToken = patch.BotToken
	}
	if patch.BaseURL != "" {
		cfg.BaseURL = patch.BaseURL
	}
	if patch.UserID != "" {
		cfg.UserID = patch.UserID
	}
	if patch.BotAgent != "" {
		cfg.BotAgent = patch.BotAgent
	}
	cfg.GetUpdatesBuf = patch.GetUpdatesBuf
	return json.Marshal(cfg)
}
