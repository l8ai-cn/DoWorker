package grpc

import (
	"context"

	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	workflowDomain "github.com/anthropics/agentsmesh/backend/internal/domain/workflow"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
)

func (a *GRPCRunnerAdapter) mcpCreateWorkflow(ctx context.Context, tc *middleware.TenantContext, _ string, payload []byte) (interface{}, *mcpError) {
	var params struct {
		mcpResourceApplyRequest
		Enabled bool `json:"enabled"`
	}
	if err := unmarshalPayload(payload, &params); err != nil {
		return nil, err
	}
	if a.workflowPlanApply == nil || a.workflowService == nil {
		return nil, newMcpError(
			503,
			"orchestration workflow apply service unavailable",
		)
	}
	scope, planID, planErr := a.planMCPResource(
		ctx,
		tc,
		params.Resource,
		resource.KindWorkflow,
	)
	if planErr != nil {
		return nil, planErr
	}
	status := workflowDomain.StatusDisabled
	if params.Enabled {
		status = workflowDomain.StatusEnabled
	}
	applied, err := a.workflowPlanApply.ApplyWithStatus(
		ctx,
		scope,
		planID,
		status,
	)
	if err != nil {
		return nil, mapResourceControlError(err)
	}
	workflow, err := a.workflowService.GetByID(ctx, applied.WorkflowID)
	if err != nil {
		return nil, newMcpError(500, "applied workflow is unavailable")
	}

	return map[string]interface{}{
		"workflow": toMCPWorkflowSummary(workflow),
		"resource": mcpAppliedResource(
			applied.Head,
			applied.WorkerSpecSnapshotID,
		),
	}, nil
}
