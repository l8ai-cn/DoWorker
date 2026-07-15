package sessionapi

import (
	"errors"
	"net/http"

	"github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
	agentsessionsvc "github.com/anthropics/agentsmesh/backend/internal/service/agentsession"
	"github.com/anthropics/agentsmesh/backend/internal/service/airesource"
	"github.com/anthropics/agentsmesh/backend/internal/service/billing"
	"github.com/gin-gonic/gin"
)

func writeOrchestratorError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, agentpod.ErrMissingAgentSlug):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "invalid_agent"})
	case errors.Is(err, agentpod.ErrNoAvailableRunner):
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error(), "code": "runner_unavailable"})
	case errors.Is(err, agentpod.ErrRunnerDispatchFailed):
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error(), "code": "runner_dispatch_failed"})
	case errors.Is(err, agentsessionsvc.ErrSessionBindingChanged):
		c.JSON(http.StatusConflict, gin.H{
			"error": "session changed while rebuilding; refresh and try again",
			"code":  "session_binding_changed",
		})
	case errors.Is(err, airesource.ErrDisabled):
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "selected model resource is disabled",
			"code":  "model_resource_disabled",
		})
	case errors.Is(err, billing.ErrQuotaExceeded):
		c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error(), "code": "quota_exceeded"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session", "code": "internal_error"})
	}
}
