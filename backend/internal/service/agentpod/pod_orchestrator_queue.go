package agentpod

import (
	"context"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
)

type dispatchReadiness interface {
	ShouldDispatchNow(ctx context.Context, runnerID int64) bool
}

func (o *PodOrchestrator) initialPodStatus(req *OrchestrateCreatePodRequest) string {
	if req.DeferRunnerDispatch && req.OrchestrationWorkerLaunchID != nil {
		return podDomain.StatusQueued
	}
	if !req.QueueIfUnavailable || req.RunnerID == 0 || o.podCoordinator == nil {
		return ""
	}
	dc, ok := o.podCoordinator.(dispatchReadiness)
	if !ok {
		return podDomain.StatusQueued
	}
	if dc.ShouldDispatchNow(context.Background(), req.RunnerID) {
		return ""
	}
	return podDomain.StatusQueued
}
