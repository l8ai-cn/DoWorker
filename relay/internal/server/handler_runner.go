package server

import (
	"net/http"

	"github.com/gorilla/websocket"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"github.com/anthropics/agentsmesh/relay/internal/auth"
	"github.com/anthropics/agentsmesh/relay/internal/protocol"
)

func (h *Handler) HandleRunnerWS(w http.ResponseWriter, r *http.Request) {
	_, span := otel.Tracer("agentsmesh-relay").Start(r.Context(), "relay.ws.runner")
	defer span.End()

	if !h.acceptingConnections.Load() {
		http.Error(w, "server shutting down", http.StatusServiceUnavailable)
		return
	}

	tokenStr := r.URL.Query().Get("token")
	if tokenStr == "" {
		h.logger.Warn("Runner connection missing token")
		http.Error(w, "token required", http.StatusUnauthorized)
		return
	}

	claims, err := h.tokenValidator.ValidateToken(tokenStr)
	if err != nil {
		h.logger.Warn("Invalid runner token", "error", err)
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}
	if claims.ResolvedType() != auth.TokenTypeRunner {
		h.logger.Warn("Non-runner token used on runner endpoint", "user_id", claims.UserID, "pod_key", claims.PodKey)
		http.Error(w, "invalid token type", http.StatusUnauthorized)
		return
	}

	podKey := claims.PodKey
	if podKey == "" {
		h.logger.Warn("Runner token missing pod_key")
		http.Error(w, "invalid token claims", http.StatusBadRequest)
		return
	}
	span.SetAttributes(attribute.String("pod.key", podKey))

	conn, ok := h.upgrade(w, r)
	if !ok {
		return
	}
	h.logger.Info("Publisher (runner) connecting", "pod_key", podKey, "runner_id", claims.RunnerID)

	if err := conn.WriteMessage(websocket.BinaryMessage, protocol.EncodePublisherReady()); err != nil {
		h.logger.Error("Failed to acknowledge publisher readiness", "error", err, "pod_key", podKey)
		_ = conn.Close()
		return
	}
	if err := h.channelManager.HandlePublisherConnect(podKey, conn); err != nil {
		h.logger.Error("Failed to handle publisher connect", "error", err, "pod_key", podKey)
		_ = conn.Close()
	}
}
