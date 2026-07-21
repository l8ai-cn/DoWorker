package agentpod

import (
	"context"
	"errors"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/gitprovider"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/workerdependency"
	specdomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
	workercreation "github.com/l8ai-cn/agentcloud/backend/internal/service/workercreation"
	specservice "github.com/l8ai-cn/agentcloud/backend/internal/service/workerspec"
)

type WorkerSnapshotPreparer interface {
	PrepareSnapshotWithDependencies(
		context.Context,
		specservice.Scope,
		specdomain.Snapshot,
		workerdependency.Document,
	) (workercreation.PreparedSnapshot, error)
}

type WorkerSpecDependencyArtifactLoader interface {
	GetBySnapshotID(
		context.Context,
		int64,
		int64,
	) (workerdependency.Document, error)
}

func (o *PodOrchestrator) prepareSnapshotWorkerCreate(
	ctx context.Context,
	req *OrchestrateCreatePodRequest,
) error {
	if req.WorkerSpecSnapshotID == nil {
		return nil
	}
	if hasConflictingSnapshotWorkerInput(req) {
		return ErrConflictingWorkerCreateInput
	}
	if o.workerSpecs == nil {
		return ErrWorkerSpecSnapshotUnavailable
	}
	if o.workerSnapshots == nil {
		return ErrWorkerCreationUnavailable
	}
	if o.workerDependencies == nil {
		return ErrWorkerSpecDependencyUnavailable
	}
	snapshotID := *req.WorkerSpecSnapshotID
	if snapshotID <= 0 {
		return ErrWorkerSpecSnapshotMismatch
	}
	snapshot, err := o.workerSpecs.GetByID(
		ctx,
		req.OrganizationID,
		snapshotID,
	)
	if err != nil {
		return errors.Join(ErrWorkerSpecSnapshotUnavailable, err)
	}
	if snapshot.ID != snapshotID ||
		snapshot.OrganizationID != req.OrganizationID {
		return ErrWorkerSpecSnapshotMismatch
	}
	artifact, err := o.loadSnapshotDependencyArtifact(ctx, req.OrganizationID, snapshot)
	if err != nil {
		return err
	}
	prepared, err := o.workerSnapshots.PrepareSnapshotWithDependencies(
		ctx,
		specservice.Scope{OrgID: req.OrganizationID, UserID: req.UserID},
		snapshot,
		artifact,
	)
	if err != nil {
		return err
	}
	if err := validatePreparedWorkerSnapshot(snapshot, prepared); err != nil {
		return err
	}
	aliasOverride := req.Alias
	if err := projectPreparedWorkerSnapshot(req, prepared); err != nil {
		return err
	}
	if aliasOverride != nil {
		req.Alias = aliasOverride
	}
	appendWorkerSpecPromptOverride(req)
	req.workerSpecSnapshotID = workerSpecInt64Pointer(snapshotID)
	return nil
}

func hasConflictingSnapshotWorkerInput(req *OrchestrateCreatePodRequest) bool {
	return req.WorkerSpecDraft != nil ||
		req.AgentSlug != "" ||
		req.RepositoryID != nil ||
		req.AgentfileLayer != nil ||
		req.AutomationLevel != "" ||
		req.BranchName != nil ||
		req.ModelResourceID != nil ||
		req.Perpetual ||
		len(req.KnowledgeMounts) > 0 ||
		len(req.ModelResourceEnv) > 0 ||
		len(req.ModelResourceArgs) > 0
}

func projectPreparedWorkerSpec(
	req *OrchestrateCreatePodRequest,
	prepared workercreation.Prepared,
) {
	projectWorkerSpec(
		req,
		prepared.Spec,
		prepared.AgentfileLayer,
		prepared.Repository,
	)
	req.resolvedWorkerSpec = &prepared.Snapshot
	req.preResolvedDependencies = prepared.Dependencies
	req.preResolvedArtifact = prepared.Artifact
}

func projectPreparedWorkerSnapshot(
	req *OrchestrateCreatePodRequest,
	prepared workercreation.PreparedSnapshot,
) error {
	projectWorkerSpec(
		req,
		prepared.Spec,
		prepared.AgentfileLayer,
		prepared.Repository,
	)
	req.preResolvedDependencies = prepared.Dependencies
	if prepared.Dependencies != nil {
		encoded, digest, err := workerdependency.EncodeAndDigest(
			*prepared.Dependencies,
		)
		if err != nil {
			return err
		}
		req.preResolvedArtifactJSON = encoded
		req.preResolvedArtifactDigest = digest
	}
	return nil
}

func projectWorkerSpec(
	req *OrchestrateCreatePodRequest,
	spec specdomain.Spec,
	agentfileLayer string,
	repository *gitprovider.Repository,
) {
	req.AgentSlug = spec.Runtime.WorkerType.Slug.String()
	req.ModelResourceID = workerSpecModelResourcePointer(spec.Runtime.ModelBinding)
	req.RepositoryID = cloneWorkerSpecInt64Pointer(spec.Workspace.RepositoryID)
	req.BranchName = workerSpecStringPointer(spec.Workspace.Branch)
	req.Alias = workerSpecStringPointer(spec.Metadata.Alias)
	req.AutomationLevel = string(spec.TypeConfig.AutomationLevel)
	req.Perpetual = spec.Lifecycle.TerminationPolicy == specdomain.TerminationPolicyManual
	req.AgentfileLayer = workerSpecStringPointer(agentfileLayer)
	req.preparedWorkerSpec = &spec
	req.preResolvedRepository = repository
	if repository != nil {
		req.preResolvedRepositorySlug = repository.Slug
	}
}

func workerSpecStringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
