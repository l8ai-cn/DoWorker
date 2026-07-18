package agentpod

import (
	"context"
	"errors"
	"fmt"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
)

var (
	ErrDeferredPodDispatchInvalid     = errors.New("deferred pod dispatch is invalid")
	ErrDeferredPodDispatchUnavailable = errors.New(
		"deferred pod dispatch is unavailable",
	)
	ErrDeferredPodStatusTransition = errors.New(
		"deferred pod status transition failed",
	)
)

func (o *PodOrchestrator) DispatchDeferredPod(
	ctx context.Context,
	req *OrchestrateCreatePodRequest,
	result *OrchestrateCreatePodResult,
) (*OrchestrateCreatePodResult, error) {
	if req == nil || !req.DeferRunnerDispatch ||
		req.QueueIfUnavailable ||
		result == nil || result.Pod == nil ||
		result.DeferredCreateCommand == nil ||
		result.Pod.OrganizationID != req.OrganizationID ||
		result.Pod.RunnerID != req.RunnerID ||
		result.Pod.PodKey == "" ||
		result.Pod.Status != podDomain.StatusQueued ||
		result.DeferredCreateCommand.GetPodKey() != result.Pod.PodKey {
		return nil, ErrDeferredPodDispatchInvalid
	}
	if o == nil || o.podCoordinator == nil || o.podService == nil {
		return nil, ErrDeferredPodDispatchUnavailable
	}
	if err := o.podService.transitionPodStatus(
		ctx,
		result.Pod.PodKey,
		podDomain.StatusQueued,
		podDomain.StatusInitializing,
	); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDeferredPodStatusTransition, err)
	}
	result.Pod.Status = podDomain.StatusInitializing

	dispatchReq := *req
	dispatchReq.DeferRunnerDispatch = false
	sessionID := ""
	if result.Pod.SessionID != nil {
		sessionID = *result.Pod.SessionID
	}
	dispatched, err := o.dispatchCreatedPod(
		ctx,
		&dispatchReq,
		result.Pod,
		result.DeferredCreateCommand,
		sessionID,
		req.SourcePodKey != "",
	)
	if err != nil {
		return nil, err
	}
	result.DeferredCreateCommand = nil
	return dispatched, nil
}
