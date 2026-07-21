package sessionapi

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/l8ai-cn/agentcloud/backend/internal/service/agentpod"
	agentsessionsvc "github.com/l8ai-cn/agentcloud/backend/internal/service/agentsession"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/airesource"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/billing"
	specservice "github.com/l8ai-cn/agentcloud/backend/internal/service/workerspec"
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
	case errors.Is(err, specservice.ErrInvalidDraft):
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"code":  "validation_failed",
			"field": specservice.InvalidDraftField(err),
		})
	case errors.Is(err, billing.ErrQuotaExceeded):
		c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error(), "code": "quota_exceeded"})
	default:
		slog.ErrorContext(c.Request.Context(), "session pod orchestration failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session", "code": "internal_error"})
	}
}
