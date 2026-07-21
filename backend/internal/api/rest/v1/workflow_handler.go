package v1

import (
	"context"
	"errors"
	"net/http"

	"github.com/l8ai-cn/agentcloud/backend/internal/middleware"
	workflowService "github.com/l8ai-cn/agentcloud/backend/internal/service/workflow"
	"github.com/l8ai-cn/agentcloud/backend/pkg/apierr"
	"github.com/gin-gonic/gin"
)

// PodTerminatorForWorkflow defines the minimal interface needed by WorkflowHandler
// to terminate Pods (used for cancel run). Follows ISP — handler only needs TerminatePod.
type PodTerminatorForWorkflow interface {
	TerminatePod(ctx context.Context, podKey string) error
}

// WorkflowHandler handles workflow-related requests
type WorkflowHandler struct {
	workflowService    *workflowService.WorkflowService
	workflowRunService *workflowService.WorkflowRunService
	orchestrator       *workflowService.WorkflowOrchestrator
	podTerminator      PodTerminatorForWorkflow
}

// NewWorkflowHandler creates a new workflow handler
func NewWorkflowHandler(
	ls *workflowService.WorkflowService,
	lrs *workflowService.WorkflowRunService,
	orch *workflowService.WorkflowOrchestrator,
	podTerminator PodTerminatorForWorkflow,
) *WorkflowHandler {
	return &WorkflowHandler{
		workflowService:    ls,
		workflowRunService: lrs,
		orchestrator:       orch,
		podTerminator:      podTerminator,
	}
}

// ========== Workflow CRUD ==========

// ListWorkflows lists workflows for an organization
// GET /api/v1/orgs/:slug/workflows
func (h *WorkflowHandler) ListWorkflows(c *gin.Context) {
	var req listWorkflowsQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		apierr.ValidationError(c, err.Error())
		return
	}

	tenant := middleware.GetTenant(c)
	limit := req.Limit
	if limit == 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	workflows, total, err := h.workflowService.List(c.Request.Context(), &workflowService.ListWorkflowsFilter{
		OrganizationID: tenant.OrganizationID,
		Status:         req.Status,
		ExecutionMode:  req.ExecutionMode,
		CronEnabled:    req.CronEnabled,
		Query:          req.Query,
		Limit:          limit,
		Offset:         offset,
	})
	if err != nil {
		apierr.InternalError(c, "Failed to list workflows")
		return
	}

	// Enrich with active run counts (H2)
	if len(workflows) > 0 {
		workflowIDs := make([]int64, len(workflows))
		for i, l := range workflows {
			workflowIDs[i] = l.ID
		}
		if counts, err := h.workflowRunService.CountActiveRunsByWorkflowIDs(c.Request.Context(), workflowIDs); err == nil {
			for _, l := range workflows {
				if count, ok := counts[l.ID]; ok {
					l.ActiveRunCount = int(count)
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"workflows": workflows,
		"total":     total,
		"limit":     limit,
		"offset":    req.Offset,
	})
}

// GetWorkflow gets a workflow by slug
// GET /api/v1/orgs/:slug/workflows/:workflow_slug
func (h *WorkflowHandler) GetWorkflow(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	workflowSlug := c.Param("workflow_slug")

	workflow, err := h.workflowService.GetBySlug(c.Request.Context(), tenant.OrganizationID, workflowSlug)
	if err != nil {
		if errors.Is(err, workflowService.ErrWorkflowNotFound) {
			apierr.ResourceNotFound(c, "Workflow not found")
		} else {
			apierr.InternalError(c, "Failed to get workflow")
		}
		return
	}

	// Enrich with active run count (H2)
	if counts, err := h.workflowRunService.CountActiveRunsByWorkflowIDs(c.Request.Context(), []int64{workflow.ID}); err == nil {
		if count, ok := counts[workflow.ID]; ok {
			workflow.ActiveRunCount = int(count)
		}
	}

	// Enrich with average duration (M5)
	if avg, err := h.workflowRunService.GetAvgDuration(c.Request.Context(), workflow.ID); err == nil && avg != nil {
		workflow.AvgDurationSec = avg
	}

	c.JSON(http.StatusOK, gin.H{"workflow": workflow})
}
