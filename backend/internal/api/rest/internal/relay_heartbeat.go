package internal

import (
	"net/http"
	"time"

	"github.com/l8ai-cn/agentcloud/backend/pkg/apierr"
	"github.com/gin-gonic/gin"
)

func (h *RelayHandler) Heartbeat(c *gin.Context) {
	var req HeartbeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.ValidationError(c, err.Error())
		return
	}

	if err := h.relayManager.HeartbeatWithLatency(req.RelayID, req.Connections, req.CPUUsage, req.MemoryUsage, req.LatencyMs); err != nil {
		h.logger.Warn("Heartbeat from unknown relay",
			"relay_id", req.RelayID,
			"error", err)
		apierr.ResourceNotFound(c, "relay not found")
		return
	}

	if req.ActiveTunnels > 0 || req.ActiveStreams > 0 {
		h.logger.Debug("relay HTTP data-plane tunnel stats",
			"relay_id", req.RelayID,
			"active_tunnels", req.ActiveTunnels,
			"active_streams", req.ActiveStreams)
	}

	response := HeartbeatResponse{Status: "ok"}

	if req.NeedCert && h.acmeManager != nil {
		cert, key, expiry, err := h.acmeManager.GetCertificatePEM()
		if err == nil && cert != "" {
			response.TLSCert = cert
			response.TLSKey = key
			response.TLSExpiry = expiry.Format(time.RFC3339)
			h.logger.Info("TLS certificate included in heartbeat response",
				"relay_id", req.RelayID,
				"cert_expiry", expiry)
		}
	}

	c.JSON(http.StatusOK, response)
}
