package agentpod

import (
	"context"
	"errors"
	"strings"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
)

func (o *PodOrchestrator) inheritWorkerSpecSnapshot(
	ctx context.Context,
	req *OrchestrateCreatePodRequest,
	source *podDomain.Pod,
) error {
	if source.WorkerSpecSnapshotID == nil {
		return nil
	}
	if o.workerSpecs == nil {
		return ErrWorkerSpecSnapshotUnavailable
	}
	snapshotID := *source.WorkerSpecSnapshotID
	snapshot, err := o.workerSpecs.GetByID(
		ctx,
		source.OrganizationID,
		snapshotID,
	)
	if err != nil {
		return errors.Join(ErrWorkerSpecSnapshotUnavailable, err)
	}
	spec, err := specdomain.NormalizeAndValidate(snapshot.Spec)
	if err != nil {
		return errors.Join(ErrWorkerSpecSnapshotMismatch, err)
	}
	if snapshot.ID != snapshotID ||
		snapshot.OrganizationID != source.OrganizationID ||
		!sourceMatchesWorkerSpec(source, spec) {
		return ErrWorkerSpecSnapshotMismatch
	}
	if o.workerCreation == nil {
		return ErrWorkerCreationUnavailable
	}
	if err := o.workerCreation.ValidateWorkerTypeSnapshot(
		ctx,
		specservice.Scope{
			OrgID:  source.OrganizationID,
			UserID: req.UserID,
		},
		spec.Runtime.WorkerType,
	); err != nil {
		return errors.Join(ErrWorkerSpecDefinitionChanged, err)
	}
	if source.ActiveConfigRevision == nil ||
		source.ActiveConfigRevisionID == nil ||
		source.ActiveConfigRevision.ID != *source.ActiveConfigRevisionID ||
		strings.TrimSpace(source.ActiveConfigRevision.AgentfileLayer) == "" {
		return ErrWorkerSpecSnapshotMismatch
	}
	if req.AgentfileLayer != nil &&
		*req.AgentfileLayer != source.ActiveConfigRevision.AgentfileLayer {
		return ErrWorkerSpecSnapshotMismatch
	}
	req.AgentfileLayer = workerSpecStringPointer(
		source.ActiveConfigRevision.AgentfileLayer,
	)
	req.workerSpecSnapshotID = workerSpecInt64Pointer(snapshotID)
	req.preparedWorkerSpec = &spec
	return nil
}

func sourceMatchesWorkerSpec(
	source *podDomain.Pod,
	spec specdomain.Spec,
) bool {
	return source.AgentSlug == spec.Runtime.WorkerType.Slug.String() &&
		int64PointersEqual(
			source.ModelResourceID,
			workerSpecInt64Pointer(spec.Runtime.ModelBinding.ResourceID),
		) &&
		int64PointersEqual(source.RepositoryID, spec.Workspace.RepositoryID) &&
		stringPointerMatches(source.BranchName, spec.Workspace.Branch) &&
		source.InteractionMode == string(spec.TypeConfig.InteractionMode) &&
		source.AutomationLevel == string(spec.TypeConfig.AutomationLevel)
}
