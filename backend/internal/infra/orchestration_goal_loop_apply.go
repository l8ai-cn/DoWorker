package infra

import (
	"context"
	"errors"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	controlservice "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
	"gorm.io/gorm"
)

func (repo *orchestrationResourceRepo) RunGoalLoopApplyTransaction(
	ctx context.Context,
	scope control.Scope,
	planID string,
	build controlservice.GoalLoopApplyBuilder,
) (controlservice.AppliedGoalLoop, error) {
	if build == nil {
		return controlservice.AppliedGoalLoop{}, control.ErrInvalid
	}
	if err := validateApplyCoordinates(scope, planID); err != nil {
		return controlservice.AppliedGoalLoop{}, err
	}
	var goalLoopID int64
	var snapshotID int64
	head, err := repo.runApplyTransaction(
		ctx,
		scope,
		planID,
		func(
			tx *gorm.DB,
			state controlservice.LockedApplyState,
		) (controlservice.ApplyMutation, error) {
			if state.Plan.Target.Kind != resource.KindGoalLoop ||
				state.Plan.ArtifactKind != resource.KindGoalLoop+"Apply" ||
				state.Head != nil {
				return controlservice.ApplyMutation{}, control.ErrInvalid
			}
			mutation, err := build(state)
			if err != nil {
				return controlservice.ApplyMutation{}, err
			}
			if err := validateGoalLoopProjection(mutation); err != nil {
				return controlservice.ApplyMutation{}, err
			}
			goalLoopID, err = writeGoalLoopProjection(tx, state, mutation)
			if err != nil {
				return controlservice.ApplyMutation{}, err
			}
			snapshotID = mutation.Projection.WorkerSpecSnapshotID
			return mutation.ApplyMutation, nil
		},
	)
	if errors.Is(err, control.ErrConsumed) {
		return repo.loadAppliedGoalLoop(ctx, scope, planID)
	}
	if err != nil {
		return controlservice.AppliedGoalLoop{}, err
	}
	return controlservice.AppliedGoalLoop{
		Head: head, GoalLoopID: goalLoopID,
		WorkerSpecSnapshotID: snapshotID,
		ResourceRevision:     head.Revision,
	}, nil
}

func validateGoalLoopProjection(
	mutation controlservice.GoalLoopApplyMutation,
) error {
	if mutation.Head.Identity.Kind != resource.KindGoalLoop ||
		mutation.Projection.Name == "" ||
		mutation.Projection.WorkerSpecSnapshotID <= 0 ||
		mutation.Revision.WorkerSpecSnapshotID !=
			mutation.Projection.WorkerSpecSnapshotID {
		return control.ErrInvalid
	}
	return nil
}

func (repo *orchestrationResourceRepo) loadAppliedGoalLoop(
	ctx context.Context,
	scope control.Scope,
	planID string,
) (controlservice.AppliedGoalLoop, error) {
	head, err := repo.loadAppliedResourceResult(
		ctx,
		scope,
		planID,
		resource.KindGoalLoop,
		resource.KindGoalLoop+"Apply",
	)
	if err != nil {
		return controlservice.AppliedGoalLoop{}, err
	}
	revision, err := repo.GetRevision(ctx, scope, head.ID, head.Revision)
	if err != nil {
		return controlservice.AppliedGoalLoop{}, err
	}
	var row orchestrationGoalLoopRecord
	err = repo.db.WithContext(ctx).Where(
		"organization_id = ? AND orchestration_resource_id = ?",
		scope.OrganizationID,
		head.ID,
	).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return controlservice.AppliedGoalLoop{}, control.ErrCorrupt
	}
	if err != nil {
		return controlservice.AppliedGoalLoop{}, err
	}
	if row.OrchestrationResourceRevision != head.Revision ||
		row.WorkerSpecSnapshotID != revision.WorkerSpecSnapshotID ||
		row.WorkerSpecSnapshotID <= 0 {
		return controlservice.AppliedGoalLoop{}, control.ErrCorrupt
	}
	return controlservice.AppliedGoalLoop{
		Head: head, GoalLoopID: row.ID,
		WorkerSpecSnapshotID: row.WorkerSpecSnapshotID,
		ResourceRevision:     row.OrchestrationResourceRevision,
	}, nil
}
