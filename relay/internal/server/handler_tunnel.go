package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"

	"github.com/anthropics/agentsmesh/relay/internal/auth"
	"github.com/anthropics/agentsmesh/relay/internal/protocol/tunnelframe"
	"github.com/anthropics/agentsmesh/relay/internal/tunnel"
)

// TunnelHandler accepts runner-initiated tunnel WebSocket connections on
// /runner/tunnel and registers them in the tunnel registry.
type TunnelHandler struct {
	validator     *auth.TokenValidator
	registry      *tunnel.Registry
	originChecker *auth.OriginChecker
	window        int
	logger        *slog.Logger
}

// NewTunnelHandler constructs a TunnelHandler.
func NewTunnelHandler(v *auth.TokenValidator, registry *tunnel.Registry, oc *auth.OriginChecker, window int) *TunnelHandler {
	if oc == nil {
		oc = auth.NewOriginChecker(nil)
	}
	return &TunnelHandler{
		validator:     v,
		registry:      registry,
		originChecker: oc,
		window:        window,
		logger:        slog.With("component", "tunnel_handler"),
	}
}

// HandleTunnelWS handles a runner tunnel connection.
// Path: /runner/tunnel?token=<tunnel-jwt>
func (h *TunnelHandler) HandleTunnelWS(w http.ResponseWriter, r *http.Request) {
	if !h.originChecker.Allowed(r.Header.Get("Origin")) {
		http.Error(w, "forbidden origin", http.StatusForbidden)
		return
	}

	tokenStr := r.URL.Query().Get("token")
	if tokenStr == "" {
		http.Error(w, "token required", http.StatusUnauthorized)
		return
	}
	claims, err := h.validator.ValidateToken(tokenStr)
	if err != nil {
		h.logger.Warn("invalid tunnel token", "error", err)
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}
	if claims.ResolvedType() != auth.TokenTypeTunnel {
		h.logger.Warn("non-tunnel token on tunnel endpoint", "type", claims.ResolvedType())
		http.Error(w, "invalid token type", http.StatusUnauthorized)
		return
	}

	up := websocket.Upgrader{
		ReadBufferSize:  1024 * 64,
		WriteBufferSize: 1024 * 64,
		CheckOrigin:     func(*http.Request) bool { return true },
	}
	conn, err := up.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("tunnel upgrade failed", "error", err)
		return
	}

	// Read and verify the HELLO frame before registering.
	_ = conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	_, data, err := conn.ReadMessage()
	if err != nil {
		h.logger.Warn("tunnel hello read failed", "error", err)
		_ = conn.Close()
		return
	}
	f, derr := tunnelframe.Decode(data)
	if derr != nil || f.Type != tunnelframe.TypeHello {
		h.logger.Warn("tunnel first frame not HELLO")
		_ = conn.Close()
		return
	}
	var hello tunnelframe.HelloPayload
	if len(f.Payload) > 0 {
		_ = json.Unmarshal(f.Payload, &hello)
	}
	// The verified tunnel token is the authoritative source of the runner id
	// (routing is by JWT claim, never a lookup). HELLO.runner_id is advisory: if
	// present it must match, but an empty value is accepted for runners that do
	// not know their numeric id.
	if hello.RunnerID != "" && hello.RunnerID != strconv.FormatInt(claims.RunnerID, 10) {
		h.logger.Warn("tunnel hello runner mismatch", "hello", hello.RunnerID, "claim", claims.RunnerID)
		_ = conn.Close()
		return
	}
	_ = conn.SetReadDeadline(time.Time{})

	tun := tunnel.NewTunnel(conn, claims.RunnerID, claims.OrgID, h.window, h.logger)
	h.registry.Register(tun)
	if err := tun.WriteFrame(tunnelframe.Frame{Type: tunnelframe.TypeHelloAck}); err != nil {
		h.registry.Unregister(tun)
		tun.Close()
		h.logger.Warn("tunnel HELLO_ACK write failed", "runner_id", claims.RunnerID, "error", err)
		return
	}
	h.logger.Info("tunnel registered", "runner_id", claims.RunnerID)

	tun.Start() // blocks until the tunnel closes
	h.registry.Unregister(tun)
	h.logger.Info("tunnel unregistered", "runner_id", claims.RunnerID)
}
