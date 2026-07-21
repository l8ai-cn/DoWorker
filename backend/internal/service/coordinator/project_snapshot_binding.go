package coordinator

import (
	"context"
	"errors"

	workerspecdom "github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
)

var ErrCoordinatorWorkerSpecArtifactRequired = errors.New(
	"coordinator: worker spec dependency artifact is required",
)
var ErrCoordinatorWorkerSpecSnapshotMismatch = errors.New(
	"coordinator: worker spec snapshot does not match its dependency artifact",
)
var ErrCoordinatorAgentSlugDerived = errors.New(
	"coordinator: agent slug is derived from worker spec snapshot",
)

func (s *Service) projectSnapshotWorkerSlug(
	ctx context.Context,
	organizationID int64,
	snapshotID int64,
) (string, error) {
	if snapshotID <= 0 {
		return "", ErrWorkerSpecSnapshotRequired
	}
	if s.snapshots == nil {
		return "", ErrCoordinatorWorkerSpecSnapshotStoreRequired
	}
	snapshot, err := s.snapshots.GetByID(ctx, organizationID, snapshotID)
	if err != nil {
		return "", err
	}
	if snapshot.ID != snapshotID || snapshot.OrganizationID != organizationID {
		return "", ErrCoordinatorWorkerSpecSnapshotMismatch
	}
	if s.artifacts == nil {
		return "", ErrCoordinatorWorkerSpecArtifactRequired
	}
	artifact, err := s.artifacts.GetBySnapshotID(ctx, organizationID, snapshotID)
	if err != nil {
		return "", ErrCoordinatorWorkerSpecArtifactRequired
	}
	return validateCoordinatorSnapshotArtifact(snapshot, artifact.Worker.WorkerType.String())
}

func validateCoordinatorSnapshotArtifact(
	snapshot workerspecdom.Snapshot,
	artifactWorkerType string,
) (string, error) {
	workerType := snapshot.Spec.Runtime.WorkerType.Slug.String()
	if workerType == "" || artifactWorkerType != workerType {
		return "", ErrCoordinatorWorkerSpecSnapshotMismatch
	}
	return workerType, nil
}
