package grpc

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	workflowDomain "github.com/anthropics/agentsmesh/backend/internal/domain/workflow"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	workflowService "github.com/anthropics/agentsmesh/backend/internal/service/workflow"
)

func (a *GRPCRunnerAdapter) mcpListWorkflows(ctx context.Context, tc *middleware.TenantContext, payload []byte) (interface{}, *mcpError) {
	if a.workflowService == nil {
		return nil, newMcpError(500, "workflow service not available")
	}

	var params struct {
		Status string `json:"status"`
		Query  string `json:"query"`
		Limit  int    `json:"limit"`
		Offset int    `json:"offset"`
	}
	if err := unmarshalPayload(payload, &params); err != nil {
		return nil, err
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset := params.Offset
	if offset < 0 {
		offset = 0
	}

	workflows, _, err := a.workflowService.List(ctx, &workflowDomain.ListWorkflowsFilter{
		OrganizationID: tc.OrganizationID,
		Status:         params.Status,
		Query:          params.Query,
		Limit:          limit,
		Offset:         offset,
	})
	if err != nil {
		return nil, newMcpError(500, "failed to list workflows")
	}

	if len(workflows) > 0 && a.workflowRunService != nil {
		workflowIDs := make([]int64, len(workflows))
		for i, l := range workflows {
			workflowIDs[i] = l.ID
		}
		if counts, err := a.workflowRunService.CountActiveRunsByWorkflowIDs(ctx, workflowIDs); err == nil {
			for _, l := range workflows {
				if count, ok := counts[l.ID]; ok {
					l.ActiveRunCount = int(count)
				}
			}
		}
	}

	summaries := make([]*mcpWorkflowSummary, len(workflows))
	for i, l := range workflows {
		summaries[i] = toMCPWorkflowSummary(l)
	}

	return map[string]interface{}{"workflows": summaries}, nil
}

func (a *GRPCRunnerAdapter) mcpTriggerWorkflow(ctx context.Context, tc *middleware.TenantContext, payload []byte) (interface{}, *mcpError) {
	if a.workflowService == nil || a.workflowOrchestrator == nil {
		return nil, newMcpError(500, "workflow service not available")
	}

	var params struct {
		WorkflowSlug string          `json:"workflow_slug"`
		Variables    json.RawMessage `json:"variables"`
	}
	if err := unmarshalPayload(payload, &params); err != nil {
		return nil, err
	}

	if params.WorkflowSlug == "" {
		return nil, newMcpError(400, "workflow_slug is required")
	}

	workflow, err := a.workflowService.GetBySlug(ctx, tc.OrganizationID, params.WorkflowSlug)
	if err != nil {
		if errors.Is(err, workflowService.ErrWorkflowNotFound) {
			return nil, newMcpError(404, "workflow not found")
		}
		return nil, newMcpError(500, "failed to get workflow")
	}

	result, err := a.workflowOrchestrator.TriggerRun(ctx, &workflowService.TriggerRunRequest{
		WorkflowID:    workflow.ID,
		TriggerType:   workflowDomain.RunTriggerManual,
		TriggerSource: "pod:" + strconv.FormatInt(tc.UserID, 10),
		TriggerParams: params.Variables,
	})
	if err != nil {
		if errors.Is(err, workflowService.ErrWorkflowDisabled) {
			return nil, newMcpError(400, "workflow is disabled")
		}
		if errors.Is(err, workflowDomain.ErrWorkflowResourceRequired) {
			return nil, newMcpError(409, err.Error())
		}
		return nil, newMcpError(500, "failed to trigger workflow")
	}

	if result.Skipped {
		return map[string]interface{}{
			"run":     toMCPRunSummary(result.Run),
			"skipped": true,
			"reason":  result.Reason,
		}, nil
	}

	startCtx, startCancel := context.WithTimeout(context.Background(), 5*time.Minute)
	go func() {
		defer startCancel()
		a.workflowOrchestrator.StartRun(startCtx, result.Workflow, result.Run, tc.UserID)
	}()

	return map[string]interface{}{
		"run": toMCPRunSummary(result.Run),
	}, nil
}
