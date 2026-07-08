package server

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/anthropics/agentsmesh/relay/internal/auth"
	"github.com/anthropics/agentsmesh/relay/internal/proxy"
	"github.com/anthropics/agentsmesh/relay/internal/tunnel"
)

// previewCookieName is the session cookie set by HandlePreviewSession after a
// one-shot token exchange, scoped to /preview/{podKey}.
const previewCookieName = "gw_preview"

// PreviewConfig bounds the preview HTTP entrypoint's runtime behaviour.
type PreviewConfig struct {
	ReconnectGrace    time.Duration // WaitForTunnel grace window
	StreamTimeout     time.Duration // single-stream timeout (never closes the whole tunnel)
	StreamWindowBytes int           // credit window per stream
	CookieSecure      bool          // Secure flag on the gw_preview cookie
}

// PreviewHandler serves ANY /preview/{podKey}/* by resolving the pod's runner
// tunnel from the JWT claim (no table lookup) and proxying the request over
// it. Routing is entirely claim-driven: the token embeds runner_id and the
// loopback target the backend already validated.
type PreviewHandler struct {
	validator *auth.TokenValidator
	registry  *tunnel.Registry
	limiter   *tunnel.PodLimiter
	cfg       PreviewConfig
	logger    *slog.Logger
}

// NewPreviewHandler constructs a PreviewHandler.
func NewPreviewHandler(v *auth.TokenValidator, registry *tunnel.Registry, limiter *tunnel.PodLimiter, cfg PreviewConfig) *PreviewHandler {
	return &PreviewHandler{
		validator: v,
		registry:  registry,
		limiter:   limiter,
		cfg:       cfg,
		logger:    slog.With("component", "preview_handler"),
	}
}

// route dispatches all /preview/ traffic: the __session sub-path performs the
// token->cookie exchange, everything else is proxied.
func (h *PreviewHandler) route(w http.ResponseWriter, r *http.Request) {
	_, rest, ok := parsePreviewPath(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}
	if rest == "__session" {
		h.HandlePreviewSession(w, r)
		return
	}
	h.HandlePreview(w, r)
}

// parsePreviewPath splits "/preview/{podKey}/{rest...}" into podKey and rest
// (rest may be empty for the pod root, e.g. "/preview/pod1").
func parsePreviewPath(p string) (podKey, rest string, ok bool) {
	trimmed := strings.TrimPrefix(p, "/preview/")
	if trimmed == "" || trimmed == p {
		return "", "", false
	}
	if idx := strings.Index(trimmed, "/"); idx >= 0 {
		return trimmed[:idx], trimmed[idx+1:], true
	}
	return trimmed, "", true
}

// extractToken prefers the session cookie (set after __session exchange) and
// falls back to a query-string token for the initial/direct request.
func (h *PreviewHandler) extractToken(r *http.Request) string {
	if c, err := r.Cookie(previewCookieName); err == nil && c.Value != "" {
		return c.Value
	}
	return r.URL.Query().Get("token")
}

// HandlePreview authenticates and proxies a single preview HTTP request (or,
// once upgraded, a WebSocket) over the pod's runner tunnel.
func (h *PreviewHandler) HandlePreview(w http.ResponseWriter, r *http.Request) {
	podKey, rest, ok := parsePreviewPath(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}

	claims, ok := h.authenticate(w, r, podKey)
	if !ok {
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
		PodKey:      podKey,
		Target:      claims.PreviewTarget,
		Path:        "/" + rest,
		WindowBytes: h.cfg.StreamWindowBytes,
		Timeout:     h.cfg.StreamTimeout,
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

// authenticate validates the preview token (cookie or query param), enforcing
// token_type=preview and that the claim's pod_key matches the requested path.
func (h *PreviewHandler) authenticate(w http.ResponseWriter, r *http.Request, podKey string) (*auth.RelayClaims, bool) {
	tokenStr := h.extractToken(r)
	if tokenStr == "" {
		writePreviewError(w, "token_required", http.StatusUnauthorized)
		return nil, false
	}
	claims, err := h.validator.ValidateToken(tokenStr)
	if err != nil {
		writePreviewError(w, "invalid_token", http.StatusUnauthorized)
		return nil, false
	}
	if claims.ResolvedType() != auth.TokenTypePreview || claims.PodKey != podKey {
		// Deliberately vague to avoid leaking whether the pod exists.
		writePreviewError(w, "invalid_token", http.StatusUnauthorized)
		return nil, false
	}
	return claims, true
}

func isWebSocketUpgrade(r *http.Request) bool {
	return strings.EqualFold(r.Header.Get("Upgrade"), "websocket")
}

func writePreviewError(w http.ResponseWriter, code string, status int) {
	http.Error(w, code, status)
}
