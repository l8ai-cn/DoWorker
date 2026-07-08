package imbridge

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	domain "github.com/anthropics/agentsmesh/backend/internal/domain/imbridge"
)

// WeixinProvider implements personal WeChat via Tencent iLink (OpenClaw openclaw-weixin pattern).
type WeixinProvider struct {
	HTTP   *http.Client
	client *ilinkClient
}

func NewWeixinProvider(httpClient *http.Client) *WeixinProvider {
	return &WeixinProvider{
		HTTP:   httpClient,
		client: newIlinkClient(httpClient),
	}
}

func (p *WeixinProvider) Type() string        { return domain.ProviderWeixin }
func (p *WeixinProvider) DisplayName() string { return "微信" }

func (p *WeixinProvider) ValidateConfig(raw json.RawMessage) error {
	cfg, err := parseWeixinConfig(raw)
	if err != nil {
		return err
	}
	if cfg.BotToken == "" {
		return nil // pre-QR-login placeholder config
	}
	if cfg.AccountID == "" {
		return errors.New("weixin requires account_id after login")
	}
	return nil
}

func (p *WeixinProvider) VerifyWebhook(_ context.Context, _ json.RawMessage, _ http.Header, _ []byte) error {
	return nil
}

func (p *WeixinProvider) ParseInbound(_ context.Context, _ json.RawMessage, _ http.Header, body []byte) (*InboundEvent, error) {
	var msg map[string]any
	if err := json.Unmarshal(body, &msg); err != nil {
		return nil, err
	}
	if event := parseWeixinInbound(msg); event != nil {
		return event, nil
	}
	return nil, nil
}

func (p *WeixinProvider) SendOutbound(ctx context.Context, raw json.RawMessage, msg OutboundMessage) error {
	cfg, err := parseWeixinConfig(raw)
	if err != nil {
		return err
	}
	if cfg.BotToken == "" {
		return errors.New("weixin connection is not logged in")
	}
	toUser := msg.ExternalThreadID
	if strings.HasPrefix(toUser, "channel:") {
		toUser = ""
	}
	if toUser == "" {
		return errors.New("weixin outbound requires peer user id")
	}
	return p.client.sendText(ctx, cfg, toUser, msg.ContextToken, msg.Text)
}

func (p *WeixinProvider) ilink() *ilinkClient {
	if p.client == nil {
		p.client = newIlinkClient(p.HTTP)
	}
	return p.client
}
