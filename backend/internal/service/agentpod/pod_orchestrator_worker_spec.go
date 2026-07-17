package agentpod

import (
	"bytes"
	"context"
	"errors"
	"strings"

	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	workercreation "github.com/anthropics/agentsmesh/backend/internal/service/workercreation"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
)

var (
	ErrWorkerCreationUnavailable    = errors.New("worker creation service is not configured")
	ErrConflictingWorkerCreateInput = errors.New(
		"structured workerspec cannot be combined with legacy worker fields",
	)
	ErrInvalidPreparedWorkerSpec = errors.New(
		"worker creation service returned an inconsistent prepared workerspec",
	)
	ErrWorkerSpecSnapshotUnavailable = errors.New(
		"source workerspec snapshot is unavailable",
	)
	ErrWorkerSpecSnapshotMismatch = errors.New(
		"source pod does not match its workerspec snapshot",
	)
	ErrWorkerSpecDefinitionChanged = errors.New(
		"workerspec worker type definition changed",
	)
)

type WorkerCreationPreparer interface {
	Prepare(
		context.Context,
		specservice.Scope,
		workercreation.Draft,
	) (workercreation.Prepared, error)
	ValidateWorkerTypeSnapshot(
		context.Context,
		specservice.Scope,
		specdomain.WorkerType,
	) error
}

type WorkerSpecSnapshotLoader interface {
	GetByID(
		context.Context,
		int64,
		int64,
	) (specdomain.Snapshot, error)
}

func (o *PodOrchestrator) prepareStructuredWorkerCreate(
	ctx context.Context,
	req *OrchestrateCreatePodRequest,
) error {
	if req.WorkerSpecDraft == nil {
		return nil
	}
	if hasConflictingWorkerCreateInput(req) {
		return ErrConflictingWorkerCreateInput
	}
	if o.workerCreation == nil {
		return ErrWorkerCreationUnavailable
	}
	prepared, err := o.workerCreation.Prepare(
		ctx,
		specservice.Scope{OrgID: req.OrganizationID, UserID: req.UserID},
		*req.WorkerSpecDraft,
	)
	if err != nil {
		return err
	}
	if err := validatePreparedWorkerSpec(req.OrganizationID, prepared); err != nil {
		return err
	}
	projectPreparedWorkerSpec(req, prepared)
	return nil
}

func (o *PodOrchestrator) validatePreparedWorkerType(
	ctx context.Context,
	req *OrchestrateCreatePodRequest,
) error {
	if req.preparedWorkerSpec == nil {
		return nil
	}
	if o.workerCreation == nil {
		return ErrWorkerCreationUnavailable
	}
	if err := o.workerCreation.ValidateWorkerTypeSnapshot(
		ctx,
		specservice.Scope{OrgID: req.OrganizationID, UserID: req.UserID},
		req.preparedWorkerSpec.Runtime.WorkerType,
	); err != nil {
		return errors.Join(ErrWorkerSpecDefinitionChanged, err)
	}
	return nil
}

func hasConflictingWorkerCreateInput(req *OrchestrateCreatePodRequest) bool {
	return req.AgentSlug != "" ||
		req.RepositoryID != nil ||
		req.Alias != nil ||
		req.AgentfileLayer != nil ||
		req.AutomationLevel != "" ||
		req.BranchName != nil ||
		req.ModelResourceID != nil ||
		req.TokenBudget != nil ||
		req.Perpetual ||
		req.LocalPath != "" ||
		len(req.KnowledgeMounts) > 0 ||
		len(req.ModelResourceEnv) > 0 ||
		len(req.ModelResourceArgs) > 0
}

func validatePreparedWorkerSpec(
	organizationID int64,
	prepared workercreation.Prepared,
) error {
	if prepared.Snapshot.OrganizationID() != organizationID ||
		strings.TrimSpace(prepared.AgentfileLayer) == "" {
		return ErrInvalidPreparedWorkerSpec
	}
	encoded, err := specdomain.EncodeSpec(prepared.Spec)
	if err != nil {
		return errors.Join(ErrInvalidPreparedWorkerSpec, err)
	}
	if !bytes.Equal(encoded, prepared.Snapshot.SpecJSON()) {
		return ErrInvalidPreparedWorkerSpec
	}
	return validatePreparedWorkerRepository(
		organizationID,
		prepared.Spec.Workspace.RepositoryID,
		prepared.Repository,
	)
}

func int64PointersEqual(left, right *int64) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func stringPointerMatches(value *string, expected string) bool {
	if value == nil {
		return expected == ""
	}
	return *value == expected
}

func cloneWorkerSpecInt64Pointer(value *int64) *int64 {
	if value == nil {
		return nil
	}
	return workerSpecInt64Pointer(*value)
}

func workerSpecInt64Pointer(value int64) *int64 {
	return &value
}

func workerSpecModelResourcePointer(binding specdomain.ModelBinding) *int64 {
	if binding.IsEmpty() {
		return nil
	}
	return workerSpecInt64Pointer(binding.ResourceID)
}

func workerSpecStringPointer(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
