package server

import (
	"context"
	"time"

	"github.com/l8ai-cn/agentcloud/relay/internal/auth"
	relaybackend "github.com/l8ai-cn/agentcloud/relay/internal/backend"
	"github.com/l8ai-cn/agentcloud/relay/internal/config"
)

const previewCookieName = "gw_preview"

type PreviewConfig struct {
	ReconnectGrace    time.Duration
	StreamTimeout     time.Duration
	StreamWindowBytes int
	ReauthorizeEvery  time.Duration
	CookieSecure      bool
	CookieMode        config.PreviewCookieMode
	PublicOrigin      string
	PublicHost        string
}

type previewSessionIssuer interface {
	Issue(bootstrap *auth.RelayClaims, expiry time.Duration) (*auth.IssuedPreviewSession, error)
}

type previewSessionBackend interface {
	RedeemPreviewBootstrap(ctx context.Context, bootstrapID string, session relaybackend.PreviewSessionRegistration) error
	AuthorizePreviewSession(ctx context.Context, identity relaybackend.PreviewSessionIdentity) error
}
