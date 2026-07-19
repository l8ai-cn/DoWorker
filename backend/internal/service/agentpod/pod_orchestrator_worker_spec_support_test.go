package agentpod

import (
	"context"

	poddomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	workercreation "github.com/anthropics/agentsmesh/backend/internal/service/workercreation"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
	"gorm.io/gorm"
)

type workerCreationPreparer struct {
	prepared      workercreation.Prepared
	err           error
	validate      error
	calls         int
	validateCalls int
	scope         specservice.Scope
}

func (preparer *workerCreationPreparer) Prepare(
	_ context.Context,
	scope specservice.Scope,
	_ workercreation.Draft,
) (workercreation.Prepared, error) {
	preparer.calls++
	preparer.scope = scope
	return preparer.prepared, preparer.err
}

func (preparer *workerCreationPreparer) ValidateWorkerTypeSnapshot(
	context.Context,
	specservice.Scope,
	specdomain.WorkerType,
) error {
	preparer.validateCalls++
	return preparer.validate
}

type workerSpecDispatchObserver struct {
	db               *gorm.DB
	observedSnapshot bool
	err              error
}

type workerSpecSnapshotLoader struct {
	snapshot       specdomain.Snapshot
	err            error
	organizationID int64
	snapshotID     int64
}

func (loader *workerSpecSnapshotLoader) GetByID(
	_ context.Context,
	organizationID, snapshotID int64,
) (specdomain.Snapshot, error) {
	loader.organizationID = organizationID
	loader.snapshotID = snapshotID
	return loader.snapshot, loader.err
}

func (observer *workerSpecDispatchObserver) CreatePod(
	ctx context.Context,
	_ int64,
	command *runnerv1.CreatePodCommand,
) error {
	var pod poddomain.Pod
	if err := observer.db.WithContext(ctx).
		Where("pod_key = ?", command.PodKey).
		First(&pod).Error; err != nil {
		observer.err = err
		return err
	}
	observer.observedSnapshot = pod.WorkerSpecSnapshotID != nil
	return nil
}

func (observer *workerSpecDispatchObserver) CreatePodOrQueue(
	ctx context.Context,
	runnerID int64,
	command *runnerv1.CreatePodCommand,
	_ poddomain.CreatePodQueueOpts,
) error {
	return observer.CreatePod(ctx, runnerID, command)
}
