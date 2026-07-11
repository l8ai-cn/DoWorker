package v1

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	workflowService "github.com/anthropics/agentsmesh/backend/internal/service/workflow"
	"github.com/anthropics/agentsmesh/backend/pkg/apierr"
	"github.com/gin-gonic/gin"
)

func (h *WorkflowHandler) ListWorkflowRuns(c *gin.Context) {
	var req listRunsQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		apierr.ValidationError(c, err.Error())
		return
	}

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

	limit := req.Limit
	if limit == 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	runsOffset := req.Offset
	if runsOffset < 0 {
		runsOffset = 0
	}

	runs, total, err := h.workflowRunService.ListWorkflowRuns(c.Request.Context(), &workflowService.ListWorkflowRunsFilter{
		WorkflowID: workflow.ID,
		Status:     req.Status,
		Limit:      limit,
		Offset:     runsOffset,
	})
	if err != nil {
		apierr.InternalError(c, "Failed to list runs")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"runs":   runs,
		"total":  total,
		"limit":  limit,
		"offset": runsOffset,
	})
}

func (h *WorkflowHandler) GetRun(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	workflowSlug := c.Param("workflow_slug")
	runIDStr := c.Param("run_id")

	workflow, err := h.workflowService.GetBySlug(c.Request.Context(), tenant.OrganizationID, workflowSlug)
	if err != nil {
		if errors.Is(err, workflowService.ErrWorkflowNotFound) {
			apierr.ResourceNotFound(c, "Workflow not found")
		} else {
			apierr.InternalError(c, "Failed to get workflow")
		}
		return
	}

	runID, err := strconv.ParseInt(runIDStr, 10, 64)
	if err != nil {
		apierr.ValidationError(c, "Invalid run ID")
		return
	}

	run, err := h.workflowRunService.GetByID(c.Request.Context(), runID)
	if err != nil {
		if errors.Is(err, workflowService.ErrRunNotFound) {
			apierr.ResourceNotFound(c, "Run not found")
		} else {
			apierr.InternalError(c, "Failed to get run")
		}
		return
	}

	if run.WorkflowID != workflow.ID {
		apierr.ResourceNotFound(c, "Run not found")
		return
	}

	c.JSON(http.StatusOK, gin.H{"run": run})
}

func (h *WorkflowHandler) CancelRun(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	workflowSlug := c.Param("workflow_slug")
	runIDStr := c.Param("run_id")

	workflow, err := h.workflowService.GetBySlug(c.Request.Context(), tenant.OrganizationID, workflowSlug)
	if err != nil {
		if errors.Is(err, workflowService.ErrWorkflowNotFound) {
			apierr.ResourceNotFound(c, "Workflow not found")
		} else {
			apierr.InternalError(c, "Failed to get workflow")
		}
		return
	}

	runID, err := strconv.ParseInt(runIDStr, 10, 64)
	if err != nil {
		apierr.ValidationError(c, "Invalid run ID")
		return
	}

	run, err := h.workflowRunService.GetByID(c.Request.Context(), runID)
	if err != nil {
		if errors.Is(err, workflowService.ErrRunNotFound) {
			apierr.ResourceNotFound(c, "Run not found")
		} else {
			apierr.InternalError(c, "Failed to get run")
		}
		return
	}

	if run.WorkflowID != workflow.ID {
		apierr.ResourceNotFound(c, "Run not found")
		return
	}

	if run.IsTerminal() {
		apierr.BadRequest(c, apierr.VALIDATION_FAILED, "Run is already in terminal state")
		return
	}

	// SSOT: cancel by terminating the Pod — run status will be derived from Pod state
	if run.PodKey != nil && h.podTerminator != nil {
		if err := h.podTerminator.TerminatePod(c.Request.Context(), *run.PodKey); err != nil {
			apierr.InternalError(c, "Failed to terminate pod")
			return
		}
	} else {
		if err := h.orchestrator.MarkRunCancelled(c.Request.Context(), runID, "Cancelled by user"); err != nil {
			apierr.InternalError(c, "Failed to cancel run")
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Run cancelled"})
}
