package imbridge

import (
	"context"
	"encoding/json"
	"net/http"
)

// InboundEvent is the normalized message shape after a provider parses inbound traffic.
type InboundEvent struct {
	ExternalThreadID string
	SenderName       string
	Text             string
	Challenge        string // URL verification handshake (Feishu/Slack)
	ContextToken     string // Weixin iLink reply token
}

// OutboundMessage is the normalized outbound payload for collaboration channels.
type OutboundMessage struct {
	ExternalThreadID string
	Text             string
	SenderLabel      string
	ContextToken     string // Weixin iLink reply token
}

// Provider implements one IM platform. Registry pattern mirrors OpenClaw
// claw-connect identities and knowledgebase connectors — add providers
// without touching the core bridge service.
type Provider interface {
	Type() string
	DisplayName() string
	ValidateConfig(raw json.RawMessage) error
	VerifyWebhook(ctx context.Context, cfg json.RawMessage, headers http.Header, body []byte) error
	ParseInbound(ctx context.Context, cfg json.RawMessage, headers http.Header, body []byte) (*InboundEvent, error)
	SendOutbound(ctx context.Context, cfg json.RawMessage, msg OutboundMessage) error
}
