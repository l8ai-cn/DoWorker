package agentpod

import (
	"context"
	"errors"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
)

var (
	ErrConflictingWorkerSpecPersistence = errors.New(
		"cannot create and inherit a workerspec snapshot at the same time",
	)
	ErrWorkerSpecPersistenceUnavailable = errors.New(
		"atomic workerspec pod persistence is unavailable",
	)
)

type workerSpecPodRepository interface {
	CreateWithConfigAndWorkerSpec(
		context.Context,
		*agentpod.Pod,
		*agentpod.PodConfigRevision,
		specservice.ResolvedSnapshot,
	) error
}

func (service *PodService) persistPodWithWorkerSpec(
	ctx context.Context,
	req *CreatePodRequest,
	pod *agentpod.Pod,
	revision *agentpod.PodConfigRevision,
) error {
	if req.ResolvedWorkerSpec != nil && req.WorkerSpecSnapshotID != nil {
		return ErrConflictingWorkerSpecPersistence
	}
	if req.ResolvedWorkerSpec == nil {
		return service.repo.CreateWithConfig(ctx, pod, revision)
	}
	repository, ok := service.repo.(workerSpecPodRepository)
	if !ok {
		return ErrWorkerSpecPersistenceUnavailable
	}
	return repository.CreateWithConfigAndWorkerSpec(
		ctx,
		pod,
		revision,
		*req.ResolvedWorkerSpec,
	)
}
