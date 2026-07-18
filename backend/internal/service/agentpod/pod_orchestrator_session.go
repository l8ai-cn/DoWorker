package agentpod

import (
	"context"
	"errors"
	"log/slog"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	sessionDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentsession"
)

func (o *PodOrchestrator) prepareSessionBeforeDispatch(
	ctx context.Context,
	req *OrchestrateCreatePodRequest,
	pod *podDomain.Pod,
) (*sessionDomain.ProvisionReceipt, error) {
	if req.SessionProvision == nil {
		return nil, nil
	}
	receipt, err := o.prepareSession(ctx, pod, *req.SessionProvision)
	if err != nil {
		o.markSessionPreparationFailed(ctx, pod, err)
		return nil, err
	}
	if req.PrepareSession == nil {
		return receipt, nil
	}
	if err := req.PrepareSession(ctx, receipt.Session); err != nil {
		preparationErr := errors.Join(ErrSessionPreparationFailed, err)
		o.markSessionPreparationFailed(ctx, pod, preparationErr)
		return nil, o.rollbackSessionProvision(ctx, pod, receipt, preparationErr)
	}
	return receipt, nil
}

func (o *PodOrchestrator) prepareSession(
	ctx context.Context,
	pod *podDomain.Pod,
	spec sessionDomain.ProvisionSpec,
) (*sessionDomain.ProvisionReceipt, error) {
	if o.sessionProvisioner == nil {
		return nil, ErrSessionProvisionerUnavailable
	}
	receipt, err := o.sessionProvisioner.PrepareForPod(ctx, pod, spec)
	if err != nil {
		return nil, errors.Join(ErrSessionProvisionFailed, err)
	}
	return receipt, nil
}

func (o *PodOrchestrator) rollbackSessionProvision(
	ctx context.Context,
	pod *podDomain.Pod,
	receipt *sessionDomain.ProvisionReceipt,
	cause error,
) error {
	if receipt == nil {
		return cause
	}
	cleanupCtx, cancel := detachedCleanupContext(ctx)
	defer cancel()
	if err := o.sessionProvisioner.RollbackProvision(cleanupCtx, receipt); err != nil {
		slog.ErrorContext(ctx, "failed to roll back session provision",
			"pod_key", pod.PodKey, "error", err)
		return errors.Join(cause, ErrSessionRollbackFailed, err)
	}
	return cause
}

func (o *PodOrchestrator) markSessionPreparationFailed(
	ctx context.Context,
	pod *podDomain.Pod,
	err error,
) {
	cleanupCtx, cancel := detachedCleanupContext(ctx)
	defer cancel()
	if markErr := o.podService.MarkDispatchFailed(
		cleanupCtx,
		pod.PodKey,
		errCodeSessionProvision,
		err.Error(),
	); markErr != nil {
		slog.ErrorContext(ctx, "failed to mark pod after session preparation failure",
			"pod_key", pod.PodKey, "error", markErr)
	}
}
