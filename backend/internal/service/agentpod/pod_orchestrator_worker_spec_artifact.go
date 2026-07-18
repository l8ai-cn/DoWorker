package agentpod

import (
	"context"
	"errors"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
)

func (o *PodOrchestrator) validateSnapshotDependencyArtifact(
	ctx context.Context,
	organizationID int64,
	snapshot specdomain.Snapshot,
) error {
	artifact, err := o.workerDependencies.GetBySnapshotID(
		ctx,
		organizationID,
		snapshot.ID,
	)
	if err != nil {
		return errors.Join(ErrWorkerSpecDependencyUnavailable, err)
	}
	if artifact.OrganizationID != organizationID {
		return ErrWorkerSpecSnapshotMismatch
	}
	digest, err := workerSpecDigest(snapshot.Spec)
	if err != nil {
		return errors.Join(ErrWorkerSpecSnapshotMismatch, err)
	}
	if artifact.Worker.SpecDigest != digest {
		return ErrWorkerSpecSnapshotMismatch
	}
	return nil
}

func workerSpecDigest(spec specdomain.Spec) (string, error) {
	encoded, err := specdomain.EncodeSpec(spec)
	if err != nil {
		return "", err
	}
	canonical, err := control.CanonicalJSONObject(encoded)
	if err != nil {
		return "", err
	}
	return control.DigestCanonicalJSON(canonical)
}
