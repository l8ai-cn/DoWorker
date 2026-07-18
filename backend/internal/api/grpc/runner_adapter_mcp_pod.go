package grpc

import (
	"context"

	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
)

func (a *GRPCRunnerAdapter) mcpCreatePod(ctx context.Context, tc *middleware.TenantContext, payload []byte) (interface{}, *mcpError) {
	var params mcpResourceApplyRequest
	if err := unmarshalPayload(payload, &params); err != nil {
		return nil, err
	}
	if a.workerPlanApply == nil || a.podService == nil {
		return nil, newMcpError(
			503,
			"orchestration worker apply service unavailable",
		)
	}
	scope, planID, planErr := a.planMCPResource(
		ctx,
		tc,
		params.Resource,
		resource.KindWorker,
	)
	if planErr != nil {
		return nil, planErr
	}
	applied, err := a.workerPlanApply.Apply(ctx, scope, planID)
	if err != nil {
		return nil, mapResourceControlError(err)
	}
	pod, err := a.podService.GetPodByKey(ctx, applied.PodKey)
	if err != nil {
		return nil, newMcpError(500, "applied worker pod is unavailable")
	}

	return map[string]interface{}{
		"pod": map[string]interface{}{
			"pod_key": pod.PodKey,
			"status":  pod.Status,
		},
		"resource": mcpAppliedResource(
			applied.Head,
			applied.WorkerSpecSnapshotID,
		),
	}, nil
}
