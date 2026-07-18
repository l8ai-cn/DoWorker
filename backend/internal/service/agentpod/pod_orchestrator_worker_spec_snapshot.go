package agentpod

import (
	"context"
	"errors"
	"reflect"
	"strings"

	"github.com/anthropics/agentsmesh/backend/internal/domain/gitprovider"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	workercreation "github.com/anthropics/agentsmesh/backend/internal/service/workercreation"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
)

type WorkerSnapshotPreparer interface {
	PrepareSnapshot(
		context.Context,
		specservice.Scope,
		specdomain.Snapshot,
	) (workercreation.PreparedSnapshot, error)
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
	prepared, err := o.workerSnapshots.PrepareSnapshot(
		ctx,
		specservice.Scope{OrgID: req.OrganizationID, UserID: req.UserID},
		snapshot,
	)
	if err != nil {
		return err
	}
	if err := validatePreparedWorkerSnapshot(snapshot, prepared); err != nil {
		return err
	}
	aliasOverride := req.Alias
	projectPreparedWorkerSnapshot(req, prepared)
	if aliasOverride != nil {
		req.Alias = aliasOverride
	}
	appendWorkerSpecPromptOverride(req)
	req.workerSpecSnapshotID = workerSpecInt64Pointer(snapshotID)
	return nil
}

func hasConflictingSnapshotWorkerInput(req *OrchestrateCreatePodRequest) bool {
	return req.WorkerSpecDraft != nil ||
		req.RunnerID != 0 ||
		req.AgentSlug != "" ||
		req.RepositoryID != nil ||
		req.AgentfileLayer != nil ||
		req.AutomationLevel != "" ||
		req.BranchName != nil ||
		req.ModelResourceID != nil ||
		req.Perpetual ||
		req.LocalPath != "" ||
		len(req.KnowledgeMounts) > 0 ||
		len(req.ModelResourceEnv) > 0 ||
		len(req.ModelResourceArgs) > 0
}

func validatePreparedWorkerSnapshot(
	snapshot specdomain.Snapshot,
	prepared workercreation.PreparedSnapshot,
) error {
	spec, err := specdomain.NormalizeAndValidate(snapshot.Spec)
	if err != nil {
		return errors.Join(ErrWorkerSpecSnapshotMismatch, err)
	}
	if strings.TrimSpace(prepared.AgentfileLayer) == "" ||
		!reflect.DeepEqual(spec, prepared.Spec) {
		return ErrInvalidPreparedWorkerSpec
	}
	return validatePreparedWorkerRepository(
		snapshot.OrganizationID,
		prepared.Spec.Workspace.RepositoryID,
		prepared.Repository,
	)
}

func validatePreparedWorkerRepository(
	organizationID int64,
	repositoryID *int64,
	repository *gitprovider.Repository,
) error {
	if repositoryID == nil {
		if repository != nil {
			return ErrInvalidPreparedWorkerSpec
		}
		return nil
	}
	if repository == nil ||
		repository.ID != *repositoryID ||
		repository.OrganizationID != organizationID ||
		!repository.IsActive {
		return ErrInvalidPreparedWorkerSpec
	}
	return nil
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
}

func projectPreparedWorkerSnapshot(
	req *OrchestrateCreatePodRequest,
	prepared workercreation.PreparedSnapshot,
) {
	projectWorkerSpec(
		req,
		prepared.Spec,
		prepared.AgentfileLayer,
		prepared.Repository,
	)
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
