package agentpod

import (
	"context"
	"errors"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
)

var (
	ErrWorkerLaunchPodMismatch = errors.New(
		"orchestration worker launch does not match its existing pod",
	)
	ErrWorkerLaunchPodPersistenceUnavailable = errors.New(
		"orchestration worker launch pod persistence is unavailable",
	)
	ErrWorkerLaunchRequiresDeferredDispatch = errors.New(
		"orchestration worker launch requires durable deferred dispatch",
	)
)

type workerLaunchPodRepository interface {
	GetByOrchestrationWorkerLaunchID(
		context.Context,
		int64,
		int64,
	) (*podDomain.Pod, error)
}

func (o *PodOrchestrator) bindExistingWorkerLaunchPod(
	ctx context.Context,
	req *OrchestrateCreatePodRequest,
) (*podDomain.Pod, error) {
	if req.OrchestrationWorkerLaunchID == nil {
		return nil, nil
	}
	if o.podService == nil || *req.OrchestrationWorkerLaunchID <= 0 ||
		req.WorkerSpecSnapshotID == nil ||
		*req.WorkerSpecSnapshotID <= 0 {
		return nil, ErrWorkerLaunchPodMismatch
	}
	pod, err := o.podService.GetByOrchestrationWorkerLaunchID(
		ctx,
		req.OrganizationID,
		*req.OrchestrationWorkerLaunchID,
	)
	if err != nil || pod == nil {
		return pod, err
	}
	if pod.OrganizationID != req.OrganizationID ||
		pod.CreatedByID != req.UserID ||
		!optionalInt64sEqual(
			pod.WorkerSpecSnapshotID,
			req.WorkerSpecSnapshotID,
		) ||
		!optionalInt64sEqual(
			pod.OrchestrationWorkerLaunchID,
			req.OrchestrationWorkerLaunchID,
		) {
		return nil, ErrWorkerLaunchPodMismatch
	}
	req.RunnerID = pod.RunnerID
	req.clusterID = pod.ClusterID
	return pod, nil
}

func (service *PodService) GetByOrchestrationWorkerLaunchID(
	ctx context.Context,
	organizationID int64,
	launchID int64,
) (*podDomain.Pod, error) {
	repository, ok := service.repo.(workerLaunchPodRepository)
	if !ok {
		return nil, ErrWorkerLaunchPodPersistenceUnavailable
	}
	return repository.GetByOrchestrationWorkerLaunchID(
		ctx,
		organizationID,
		launchID,
	)
}

func (service *PodService) persistOrReuseWorkerLaunchPod(
	ctx context.Context,
	req *CreatePodRequest,
	pod *podDomain.Pod,
	revision *podDomain.PodConfigRevision,
) (*podDomain.Pod, error) {
	if req.OrchestrationWorkerLaunchID == nil {
		if err := service.persistPodWithWorkerSpec(
			ctx,
			req,
			pod,
			revision,
		); err != nil {
			return nil, err
		}
		return pod, nil
	}
	if *req.OrchestrationWorkerLaunchID <= 0 ||
		req.WorkerSpecSnapshotID == nil ||
		*req.WorkerSpecSnapshotID <= 0 {
		return nil, ErrWorkerLaunchPodMismatch
	}
	existing, err := service.GetByOrchestrationWorkerLaunchID(
		ctx,
		req.OrganizationID,
		*req.OrchestrationWorkerLaunchID,
	)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return validateWorkerLaunchPod(existing, pod)
	}
	err = service.persistPodWithWorkerSpec(ctx, req, pod, revision)
	if err == nil {
		return pod, nil
	}
	if !errors.Is(err, podDomain.ErrWorkerLaunchPodAlreadyExists) {
		return nil, err
	}
	existing, loadErr := service.GetByOrchestrationWorkerLaunchID(
		ctx,
		req.OrganizationID,
		*req.OrchestrationWorkerLaunchID,
	)
	if loadErr != nil {
		return nil, loadErr
	}
	if existing == nil {
		return nil, ErrWorkerLaunchPodMismatch
	}
	return validateWorkerLaunchPod(existing, pod)
}

func validateWorkerLaunchPod(
	existing *podDomain.Pod,
	expected *podDomain.Pod,
) (*podDomain.Pod, error) {
	if existing.OrganizationID != expected.OrganizationID ||
		existing.RunnerID != expected.RunnerID ||
		existing.ClusterID != expected.ClusterID ||
		existing.AgentSlug != expected.AgentSlug ||
		existing.CreatedByID != expected.CreatedByID ||
		existing.Prompt != expected.Prompt ||
		!optionalStringsEqual(existing.Alias, expected.Alias) ||
		!optionalStringsEqual(existing.SessionID, expected.SessionID) ||
		existing.InteractionMode != expected.InteractionMode ||
		existing.AutomationLevel != expected.AutomationLevel ||
		existing.Perpetual != expected.Perpetual ||
		!optionalInt64sEqual(
			existing.WorkerSpecSnapshotID,
			expected.WorkerSpecSnapshotID,
		) ||
		!optionalInt64sEqual(
			existing.OrchestrationWorkerLaunchID,
			expected.OrchestrationWorkerLaunchID,
		) {
		return nil, ErrWorkerLaunchPodMismatch
	}
	return existing, nil
}

func optionalStringsEqual(left, right *string) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func optionalInt64sEqual(left, right *int64) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}
