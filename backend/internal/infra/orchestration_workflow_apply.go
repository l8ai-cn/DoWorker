package infra

import (
	"context"
	"errors"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	controlservice "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
	"gorm.io/gorm"
)

func (repo *orchestrationResourceRepo) RunWorkflowApplyTransaction(
	ctx context.Context,
	scope control.Scope,
	planID string,
	build controlservice.WorkflowApplyBuilder,
) (controlservice.AppliedWorkflow, error) {
	if build == nil {
		return controlservice.AppliedWorkflow{}, control.ErrInvalid
	}
	if err := validateApplyCoordinates(scope, planID); err != nil {
		return controlservice.AppliedWorkflow{}, err
	}
	var workflowID int64
	var snapshotID int64
	head, err := repo.runApplyTransaction(
		ctx,
		scope,
		planID,
		func(
			tx *gorm.DB,
			state controlservice.LockedApplyState,
		) (controlservice.ApplyMutation, error) {
			if state.Plan.Target.Kind != resource.KindWorkflow ||
				state.Plan.ArtifactKind != resource.KindWorkflow+"Apply" {
				return controlservice.ApplyMutation{}, control.ErrInvalid
			}
			mutation, err := build(state)
			if err != nil {
				return controlservice.ApplyMutation{}, err
			}
			if err := validateWorkflowProjection(mutation); err != nil {
				return controlservice.ApplyMutation{}, err
			}
			workflowID, err = writeWorkflowProjection(tx, state, mutation)
			if err != nil {
				return controlservice.ApplyMutation{}, err
			}
			snapshotID = mutation.Projection.WorkerSpecSnapshotID
			return mutation.ApplyMutation, nil
		},
	)
	if errors.Is(err, control.ErrConsumed) {
		return repo.loadAppliedWorkflow(ctx, scope, planID)
	}
	if err != nil {
		return controlservice.AppliedWorkflow{}, err
	}
	return controlservice.AppliedWorkflow{
		Head: head, WorkflowID: workflowID,
		WorkerSpecSnapshotID: snapshotID,
		ResourceRevision:     head.Revision,
	}, nil
}

func validateWorkflowProjection(
	mutation controlservice.WorkflowApplyMutation,
) error {
	if mutation.Head.Identity.Kind != resource.KindWorkflow ||
		mutation.Projection.Name == "" ||
		mutation.Projection.WorkerSpecSnapshotID <= 0 ||
		mutation.Revision.WorkerSpecSnapshotID !=
			mutation.Projection.WorkerSpecSnapshotID {
		return control.ErrInvalid
	}
	return nil
}

func (repo *orchestrationResourceRepo) loadAppliedWorkflow(
	ctx context.Context,
	scope control.Scope,
	planID string,
) (controlservice.AppliedWorkflow, error) {
	head, err := repo.loadAppliedResourceResult(
		ctx,
		scope,
		planID,
		resource.KindWorkflow,
		resource.KindWorkflow+"Apply",
	)
	if err != nil {
		return controlservice.AppliedWorkflow{}, err
	}
	revision, err := repo.GetRevision(ctx, scope, head.ID, head.Revision)
	if err != nil {
		return controlservice.AppliedWorkflow{}, err
	}
	var row orchestrationWorkflowRecord
	err = repo.db.WithContext(ctx).Where(
		"organization_id = ? AND orchestration_resource_id = ?",
		scope.OrganizationID,
		head.ID,
	).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return controlservice.AppliedWorkflow{}, control.ErrCorrupt
	}
	if err != nil {
		return controlservice.AppliedWorkflow{}, err
	}
	if row.OrchestrationResourceRevision != head.Revision ||
		row.WorkerSpecSnapshotID != revision.WorkerSpecSnapshotID ||
		row.WorkerSpecSnapshotID <= 0 {
		return controlservice.AppliedWorkflow{}, control.ErrCorrupt
	}
	return controlservice.AppliedWorkflow{
		Head: head, WorkflowID: row.ID,
		WorkerSpecSnapshotID: row.WorkerSpecSnapshotID,
		ResourceRevision:     row.OrchestrationResourceRevision,
	}, nil
}
