package webhooks

import (
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	domainimbridge "github.com/l8ai-cn/agentcloud/backend/internal/domain/imbridge"
	imbridgesvc "github.com/l8ai-cn/agentcloud/backend/internal/service/imbridge"
)

func (r *WebhookRouter) handleIMBridgeWebhook(c *gin.Context) {
	if r.imBridge == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "im bridge not configured"})
		return
	}
	provider := c.Param("provider")
	connectionID, err := strconv.ParseInt(c.Param("connection_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid connection id"})
		return
	}
	token := c.Query("token")
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "read body failed"})
		return
	}
	resp, err := r.imBridge.HandleWebhookDeliver(c.Request.Context(), provider, connectionID, token, c.Request.Header, body)
	if err != nil {
		switch err {
		case imbridgesvc.ErrNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "connection not found"})
		case imbridgesvc.ErrConnectionPaused:
			c.JSON(http.StatusConflict, gin.H{"error": "connection not active"})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}
	if m, ok := resp.(map[string]string); ok {
		if ch, exists := m["challenge"]; exists && provider == domainimbridge.ProviderSlack {
			c.Data(http.StatusOK, "text/plain", []byte(ch))
			return
		}
		c.JSON(http.StatusOK, resp)
		return
	}
	if s, ok := resp.(string); ok {
		c.Data(http.StatusOK, "text/plain", []byte(s))
		return
	}
	c.JSON(http.StatusOK, resp)
}
