package server

import (
	"log/slog"
	"net/http"
	"strconv"
	"sync/atomic"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.opentelemetry.io/otel"

	"github.com/l8ai-cn/agentcloud/relay/internal/auth"
	"github.com/l8ai-cn/agentcloud/relay/internal/channel"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024 * 64, // 64KB
	WriteBufferSize: 1024 * 64, // 64KB
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins in development, should be restricted in production
		return true
	},
}

// Handler handles WebSocket connections
type Handler struct {
	channelManager       *channel.ChannelManager
	tokenValidator       *auth.TokenValidator
	originChecker        *auth.OriginChecker
	acceptingConnections *atomic.Bool // shared with Server for graceful shutdown
	logger               *slog.Logger
}

// NewHandler creates a new WebSocket handler with an allow-all origin checker
// (backward compatible with existing deployments/tests).
func NewHandler(channelManager *channel.ChannelManager, tokenValidator *auth.TokenValidator) *Handler {
	return NewHandlerWithOrigin(channelManager, tokenValidator, auth.NewOriginChecker(nil))
}

// NewHandlerWithOrigin creates a new WebSocket handler with an explicit origin checker.
func NewHandlerWithOrigin(channelManager *channel.ChannelManager, tokenValidator *auth.TokenValidator, originChecker *auth.OriginChecker) *Handler {
	if originChecker == nil {
		originChecker = auth.NewOriginChecker(nil)
	}
	h := &Handler{
		channelManager:       channelManager,
		tokenValidator:       tokenValidator,
		originChecker:        originChecker,
		acceptingConnections: &atomic.Bool{},
		logger:               slog.With("component", "ws_handler"),
	}
	h.acceptingConnections.Store(true)
	return h
}

// upgrade performs the WebSocket upgrade after enforcing the Origin allowlist.
func (h *Handler) upgrade(w http.ResponseWriter, r *http.Request) (*websocket.Conn, bool) {
	if !h.originChecker.Allowed(r.Header.Get("Origin")) {
		h.logger.Warn("Rejected websocket with disallowed origin", "origin", r.Header.Get("Origin"))
		http.Error(w, "forbidden origin", http.StatusForbidden)
		return nil, false
	}
	up := websocket.Upgrader{
		ReadBufferSize:  1024 * 64,
		WriteBufferSize: 1024 * 64,
		CheckOrigin:     func(*http.Request) bool { return true }, // already checked above
	}
	conn, err := up.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("upgrade failed", "error", err)
		return nil, false
	}
	return conn, true
}

// HandleBrowserWS handles browser WebSocket connections (Subscriber)
// Path: /browser/relay?token=xxx
// Channel is identified by pod_key from the token (not session_id)
func (h *Handler) HandleBrowserWS(w http.ResponseWriter, r *http.Request) {
	_, span := otel.Tracer("agentcloud-relay").Start(r.Context(), "relay.ws.browser")
	defer span.End()

	if !h.acceptingConnections.Load() {
		http.Error(w, "server shutting down", http.StatusServiceUnavailable)
		return
	}

	tokenStr := r.URL.Query().Get("token")

	if tokenStr == "" {
		h.logger.Warn("Browser connection missing token")
		http.Error(w, "token required", http.StatusUnauthorized)
		return
	}

	// Validate token
	claims, err := h.tokenValidator.ValidateToken(tokenStr)
	if err != nil {
		h.logger.Warn("Invalid token", "error", err)
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	// Enforce token type: browser endpoint only accepts browser tokens.
	// Prevents runner tokens from being used to subscribe to terminal output.
	if claims.ResolvedType() != auth.TokenTypeBrowser {
		h.logger.Warn("Non-browser token used on browser endpoint", "pod_key", claims.PodKey)
		http.Error(w, "invalid token type", http.StatusUnauthorized)
		return
	}

	// Use pod_key from token as channel identifier
	podKey := claims.PodKey

	if podKey == "" {
		h.logger.Warn("Browser token missing pod_key")
		http.Error(w, "invalid token claims", http.StatusBadRequest)
		return
	}

	conn, ok := h.upgrade(w, r)
	if !ok {
		return
	}

	// Generate subscriber ID for this browser connection
	subscriberID := uuid.New().String()

	h.logger.Info("Subscriber (browser) connecting",
		"pod_key", podKey,
		"subscriber_id", subscriberID,
		"user_id", claims.UserID)

	if err := h.channelManager.HandleSubscriberConnect(podKey, subscriberID, conn); err != nil {
		h.logger.Error("Failed to handle subscriber connect", "error", err, "pod_key", podKey)

		// Send error message before closing
		if _, ok := err.(*channel.MaxSubscribersError); ok {
			_ = conn.WriteMessage(websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "max subscribers reached"))
		}
		_ = conn.Close()
		return
	}
}

// HandleHealth handles health check requests
func (h *Handler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

// HandleStats handles stats requests
func (h *Handler) HandleStats(w http.ResponseWriter, r *http.Request) {
	stats := h.channelManager.Stats()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"active_channels":` + strconv.Itoa(stats.ActiveChannels) +
		`,"total_subscribers":` + strconv.Itoa(stats.TotalSubscribers) +
		`,"pending_publishers":` + strconv.Itoa(stats.PendingPublishers) +
		`,"pending_subscribers":` + strconv.Itoa(stats.PendingSubscribers) + `}`))
}
