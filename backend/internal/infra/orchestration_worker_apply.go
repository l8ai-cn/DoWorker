package infra

import (
	"context"
	"errors"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	controlservice "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
	"gorm.io/gorm"
)

func (repo *orchestrationResourceRepo) RunWorkerApplyTransaction(
	ctx context.Context,
	scope control.Scope,
	planID string,
	build controlservice.WorkerApplyBuilder,
) (controlservice.AppliedWorker, error) {
	if build == nil {
		return controlservice.AppliedWorker{}, control.ErrInvalid
	}
	if err := validateApplyCoordinates(scope, planID); err != nil {
		return controlservice.AppliedWorker{}, err
	}
	var launch orchestrationWorkerLaunchRecord
	head, err := repo.runApplyTransaction(
		ctx,
		scope,
		planID,
		func(
			tx *gorm.DB,
			state controlservice.LockedApplyState,
		) (controlservice.ApplyMutation, error) {
			if state.Plan.Operation != control.PlanOperationCreate ||
				state.Plan.Target.Kind != resource.KindWorker ||
				state.Plan.ArtifactKind != resource.KindWorker+"Apply" {
				return controlservice.ApplyMutation{}, control.ErrInvalid
			}
			mutation, err := build(state)
			if err != nil {
				return controlservice.ApplyMutation{}, err
			}
			if err := validateWorkerApplyMutation(mutation); err != nil {
				return controlservice.ApplyMutation{}, err
			}
			launch = newWorkerLaunchRecord(
				state.Plan.Scope,
				state.Plan.ID,
				state.AppliedAt,
				mutation,
			)
			if err := tx.Create(&launch).Error; err != nil {
				return controlservice.ApplyMutation{}, err
			}
			return mutation.ApplyMutation, nil
		},
	)
	if errors.Is(err, control.ErrConsumed) {
		return repo.loadAppliedWorker(ctx, scope, planID)
	}
	if err != nil {
		return controlservice.AppliedWorker{}, err
	}
	return appliedWorkerFromLaunch(head, launch), nil
}

func validateWorkerApplyMutation(
	mutation controlservice.WorkerApplyMutation,
) error {
	if mutation.Head.Identity.Kind != resource.KindWorker ||
		mutation.Head.Revision != 1 ||
		mutation.Launch.WorkerSpecSnapshotID <= 0 ||
		mutation.Revision.WorkerSpecSnapshotID !=
			mutation.Launch.WorkerSpecSnapshotID {
		return control.ErrInvalid
	}
	return nil
}

func (repo *orchestrationResourceRepo) loadAppliedWorker(
	ctx context.Context,
	scope control.Scope,
	planID string,
) (controlservice.AppliedWorker, error) {
	head, err := repo.loadAppliedResourceResult(
		ctx,
		scope,
		planID,
		resource.KindWorker,
		resource.KindWorker+"Apply",
	)
	if err != nil {
		return controlservice.AppliedWorker{}, err
	}
	revision, err := repo.GetRevision(ctx, scope, head.ID, head.Revision)
	if err != nil {
		return controlservice.AppliedWorker{}, err
	}
	var launch orchestrationWorkerLaunchRecord
	err = repo.db.WithContext(ctx).Where(
		"organization_id = ? AND plan_id = ?",
		scope.OrganizationID,
		planID,
	).First(&launch).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return controlservice.AppliedWorker{}, control.ErrCorrupt
	}
	if err != nil {
		return controlservice.AppliedWorker{}, err
	}
	if launch.ResourceID != head.ID ||
		launch.ResourceRevision != head.Revision ||
		launch.WorkerSpecSnapshotID != revision.WorkerSpecSnapshotID ||
		launch.WorkerSpecSnapshotID <= 0 {
		return controlservice.AppliedWorker{}, control.ErrCorrupt
	}
	return appliedWorkerFromLaunch(head, launch), nil
}

func appliedWorkerFromLaunch(
	head control.ResourceHead,
	launch orchestrationWorkerLaunchRecord,
) controlservice.AppliedWorker {
	result := controlservice.AppliedWorker{
		Head: head, LaunchID: launch.ID,
		WorkerSpecSnapshotID: launch.WorkerSpecSnapshotID,
		ResourceRevision:     launch.ResourceRevision,
	}
	if launch.PodID != nil {
		result.PodID = *launch.PodID
	}
	if launch.PodKey != nil {
		result.PodKey = *launch.PodKey
	}
	if launch.RunnerID != nil {
		result.RunnerID = *launch.RunnerID
	}
	return result
}
