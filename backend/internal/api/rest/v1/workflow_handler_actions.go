package v1

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	workflowDomain "github.com/anthropics/agentsmesh/backend/internal/domain/workflow"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	workflowService "github.com/anthropics/agentsmesh/backend/internal/service/workflow"
	"github.com/anthropics/agentsmesh/backend/pkg/apierr"
	"github.com/gin-gonic/gin"
)

// TriggerWorkflow manually triggers a workflow run.
// POST /api/v1/orgs/:slug/workflows/:workflow_slug/trigger
func (h *WorkflowHandler) TriggerWorkflow(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	workflowSlug := c.Param("workflow_slug")

	var body struct {
		Variables json.RawMessage `json:"variables"`
	}
	_ = c.ShouldBindJSON(&body)

	workflow, err := h.workflowService.GetBySlug(c.Request.Context(), tenant.OrganizationID, workflowSlug)
	if err != nil {
		if errors.Is(err, workflowService.ErrWorkflowNotFound) {
			apierr.ResourceNotFound(c, "Workflow not found")
		} else {
			apierr.InternalError(c, "Failed to get workflow")
		}
		return
	}

	result, err := h.orchestrator.TriggerRun(c.Request.Context(), &workflowService.TriggerRunRequest{
		WorkflowID:    workflow.ID,
		TriggerType:   workflowDomain.RunTriggerManual,
		TriggerSource: "user:" + strconv.FormatInt(tenant.UserID, 10),
		TriggerParams: body.Variables,
	})
	if err != nil {
		if errors.Is(err, workflowService.ErrWorkflowDisabled) {
			apierr.BadRequest(c, apierr.VALIDATION_FAILED, "Workflow is disabled")
		} else {
			apierr.InternalError(c, "Failed to trigger workflow")
		}
		return
	}

	if result.Skipped {
		c.JSON(http.StatusOK, gin.H{
			"run":     result.Run,
			"skipped": true,
			"reason":  result.Reason,
		})
		return
	}

	// Run start is async — orchestrator handles Pod creation + Autopilot setup.
	// Timeout prevents goroutine leak if Pod creation hangs indefinitely.
	startCtx, startCancel := context.WithTimeout(context.Background(), 5*time.Minute)
	go func() {
		defer startCancel()
		h.orchestrator.StartRun(startCtx, result.Workflow, result.Run, tenant.UserID)
	}()

	c.JSON(http.StatusCreated, gin.H{"run": result.Run})
}
