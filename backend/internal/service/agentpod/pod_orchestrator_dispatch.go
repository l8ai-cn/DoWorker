package agentpod

import (
	"context"
	"errors"
	"log/slog"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
)

func (o *PodOrchestrator) dispatchCreatedPod(
	ctx context.Context,
	req *OrchestrateCreatePodRequest,
	pod *podDomain.Pod,
	podCmd *runnerv1.CreatePodCommand,
	sessionID string,
	isResumeMode bool,
) (*OrchestrateCreatePodResult, error) {
	if o.podCoordinator == nil || req.DeferRunnerDispatch {
		slog.WarnContext(ctx, "PodCoordinator is nil, cannot dispatch create_pod", "pod_key", pod.PodKey)
		return &OrchestrateCreatePodResult{Pod: pod}, nil
	}

	slog.InfoContext(ctx, "dispatching create_pod to runner", "runner_id", req.RunnerID, "pod_key", pod.PodKey, "session_id", sessionID, "resume", isResumeMode)
	dispatchErr := o.podCoordinator.CreatePodOrQueue(ctx, req.RunnerID, podCmd, podDomain.CreatePodQueueOpts{
		Queue: req.QueueIfUnavailable,
		TTL:   req.QueueTTL,
		OrgID: req.OrganizationID,
	})
	switch {
	case dispatchErr == nil:
		slog.InfoContext(ctx, "create_pod dispatched", "pod_key", pod.PodKey)
		// TOCTOU: pod row may have been created as `queued` (runner looked
		// unavailable) but the runner came online before dispatch. Align
		// status so error/timeout paths (which expect `initializing`) work.
		if pod.Status == podDomain.StatusQueued {
			_ = o.podService.UpdatePodStatus(ctx, pod.PodKey, podDomain.StatusInitializing)
			pod.Status = podDomain.StatusInitializing
		}
	case podDomain.IsPodQueued(dispatchErr):
		slog.InfoContext(ctx, "create_pod queued for runner", "pod_key", pod.PodKey, "runner_id", req.RunnerID)
		if pod.Status != podDomain.StatusQueued {
			_ = o.podService.UpdatePodStatus(ctx, pod.PodKey, podDomain.StatusQueued)
			pod.Status = podDomain.StatusQueued
		}
		return &OrchestrateCreatePodResult{Pod: pod, Queued: true}, nil
	case errors.Is(dispatchErr, podDomain.ErrQueueFull):
		if markErr := o.podService.MarkDispatchFailed(ctx, pod.PodKey, errCodeQueueFull,
			"Runner pending queue is full"); markErr != nil {
			slog.ErrorContext(ctx, "failed to mark pod after queue-full", "pod_key", pod.PodKey, "error", markErr)
		}
		return nil, podDomain.ErrQueueFull
	default:
		slog.ErrorContext(ctx, "failed to dispatch create_pod", "pod_key", pod.PodKey, "error", dispatchErr)
		if markErr := o.podService.MarkDispatchFailed(ctx, pod.PodKey, errCodeRunnerUnreachable,
			"Failed to dispatch pod to runner: "+dispatchErr.Error()); markErr != nil {
			slog.ErrorContext(ctx, "failed to mark pod as dispatch failed", "pod_key", pod.PodKey, "error", markErr)
		}
		return nil, ErrRunnerDispatchFailed
	}
	return &OrchestrateCreatePodResult{Pod: pod}, nil
}
