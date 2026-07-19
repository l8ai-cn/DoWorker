package agentpod

import (
	"context"
	"errors"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	"github.com/anthropics/agentsmesh/backend/internal/domain/workerdependency"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	"github.com/anthropics/agentsmesh/backend/internal/service/workerdependencyartifact"
)

func (o *PodOrchestrator) loadSnapshotDependencyArtifact(
	ctx context.Context,
	organizationID int64,
	snapshot specdomain.Snapshot,
) (workerdependency.Document, error) {
	artifact, err := o.workerDependencies.GetBySnapshotID(
		ctx,
		organizationID,
		snapshot.ID,
	)
	if err != nil {
		return workerdependency.Document{}, errors.Join(ErrWorkerSpecDependencyUnavailable, err)
	}
	if artifact.OrganizationID != organizationID {
		return workerdependency.Document{}, ErrWorkerSpecSnapshotMismatch
	}
	digest, err := workerSpecDigest(snapshot.Spec)
	if err != nil {
		return workerdependency.Document{}, errors.Join(ErrWorkerSpecSnapshotMismatch, err)
	}
	if artifact.Worker.SpecDigest != digest {
		return workerdependency.Document{}, ErrWorkerSpecSnapshotMismatch
	}
	if err := workerdependencyartifact.ValidateWorkerSpecConsistency(
		snapshot.Spec,
		artifact,
	); err != nil {
		return workerdependency.Document{}, errors.Join(ErrWorkerSpecSnapshotMismatch, err)
	}
	return artifact, nil
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
