package server

import (
	"context"
	"log/slog"
	"net/http"
	"path"
	"strings"

	"github.com/l8ai-cn/agentcloud/relay/internal/auth"
	"github.com/l8ai-cn/agentcloud/relay/internal/proxy"
	"github.com/l8ai-cn/agentcloud/relay/internal/tunnel"
)

// PreviewHandler routes authenticated preview requests through the runner
// tunnel using only the target and canonical path bound into the JWT.
type PreviewHandler struct {
	validator      *auth.TokenValidator
	sessionIssuer  previewSessionIssuer
	sessionBackend previewSessionBackend
	registry       *tunnel.Registry
	limiter        *tunnel.PodLimiter
	cfg            PreviewConfig
	logger         *slog.Logger
}

// NewPreviewHandler constructs a PreviewHandler.
func NewPreviewHandler(
	v *auth.TokenValidator,
	issuer previewSessionIssuer,
	backend previewSessionBackend,
	registry *tunnel.Registry,
	limiter *tunnel.PodLimiter,
	cfg PreviewConfig,
) *PreviewHandler {
	return &PreviewHandler{
		validator:      v,
		sessionIssuer:  issuer,
		sessionBackend: backend,
		registry:       registry,
		limiter:        limiter,
		cfg:            cfg,
		logger:         slog.With("component", "preview_handler"),
	}
}

func (h *PreviewHandler) extractToken(r *http.Request) string {
	if c, err := r.Cookie(previewCookieName); err == nil && c.Value != "" {
		return c.Value
	}
	return ""
}

func (h *PreviewHandler) HandlePreview(w http.ResponseWriter, r *http.Request) {
	if !h.requirePublicHost(w, r) {
		return
	}
	h.handlePreview(w, r)
}

func (h *PreviewHandler) handlePreview(w http.ResponseWriter, r *http.Request) {
	podKey, _, ok := parsePreviewPath(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}
	escapedPrefix := "/preview/" + podKey
	escapedPath := r.URL.EscapedPath()
	escapedRest := ""
	if escapedPath != escapedPrefix {
		if !strings.HasPrefix(escapedPath, escapedPrefix+"/") {
			http.NotFound(w, r)
			return
		}
		escapedRest = strings.TrimPrefix(escapedPath, escapedPrefix+"/")
	}

	claims, ok := h.authenticate(w, r, podKey)
	if !ok {
		return
	}
	rawQuery, err := previewRawQuery(r.URL.RawQuery)
	if err != nil {
		writePreviewError(w, "invalid_query", http.StatusBadRequest)
		return
	}
	upstreamPath, err := joinPreviewUpstreamPath(claims.PreviewPath, escapedRest)
	if err != nil {
		writePreviewError(w, "invalid_path", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	tun := h.registry.WaitForTunnel(ctx, claims.RunnerID, h.cfg.ReconnectGrace)
	if tun == nil {
		writePreviewError(w, "target_offline", http.StatusBadGateway)
		return
	}

	release, err := h.limiter.Acquire(ctx, podKey)
	if err != nil {
		if err == tunnel.ErrTargetBusy {
			writePreviewError(w, "target_busy", http.StatusTooManyRequests)
		} else {
			writePreviewError(w, "request_cancelled", http.StatusServiceUnavailable)
		}
		return
	}
	defer release()

	params := proxy.ProxyParams{
		PodKey:           podKey,
		Target:           claims.PreviewTarget,
		Path:             upstreamPath,
		RawQuery:         rawQuery,
		HiddenCookieName: previewCookieName,
		ExpectedOrigin:   claims.PreviewOrigin,
		Reauthorize: func(ctx context.Context) error {
			return h.sessionBackend.AuthorizePreviewSession(
				ctx,
				previewSessionIdentity(claims),
			)
		},
		ReauthorizeEvery: h.cfg.ReauthorizeEvery,
		WindowBytes:      h.cfg.StreamWindowBytes,
		Timeout:          h.cfg.StreamTimeout,
	}

	if isWebSocketUpgrade(r) {
		if err := proxy.ProxyWebSocket(ctx, tun, w, r, params); err != nil {
			h.logger.Debug("preview websocket proxy error", "pod_key", podKey, "error", err)
		}
		return
	}

	if err := proxy.ProxyHTTP(ctx, tun, w, r, params); err != nil {
		h.logger.Debug("preview proxy error", "pod_key", podKey, "error", err)
	}
}

func joinPreviewUpstreamPath(base, rest string) (string, error) {
	if rest == "" {
		return base, nil
	}
	requestPath := "/" + rest
	normalizedRest, err := auth.NormalizePreviewPath(requestPath)
	if err != nil {
		return "", err
	}
	hasTrailingSlash := strings.HasSuffix(requestPath, "/")
	if requestPath != normalizedRest && (!hasTrailingSlash || strings.TrimSuffix(requestPath, "/") != normalizedRest) {
		return "", auth.ErrInvalidToken
	}
	joined := path.Join(base, strings.TrimPrefix(normalizedRest, "/"))
	if hasTrailingSlash && joined != "/" {
		joined += "/"
	}
	return joined, nil
}

func isWebSocketUpgrade(r *http.Request) bool {
	return strings.EqualFold(r.Header.Get("Upgrade"), "websocket")
}

func writePreviewError(w http.ResponseWriter, code string, status int) {
	http.Error(w, code, status)
}
