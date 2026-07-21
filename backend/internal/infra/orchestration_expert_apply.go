package infra

import (
	"context"
	"errors"

	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	resource "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationresource"
	controlservice "github.com/l8ai-cn/agentcloud/backend/internal/service/orchestrationcontrol"
	"gorm.io/gorm"
)

func (repo *orchestrationResourceRepo) RunExpertApplyTransaction(
	ctx context.Context,
	scope control.Scope,
	planID string,
	build controlservice.ExpertApplyBuilder,
) (controlservice.AppliedExpert, error) {
	if build == nil {
		return controlservice.AppliedExpert{}, control.ErrInvalid
	}
	if err := validateApplyCoordinates(scope, planID); err != nil {
		return controlservice.AppliedExpert{}, err
	}
	var expertID int64
	var snapshotID int64
	head, err := repo.runApplyTransaction(
		ctx,
		scope,
		planID,
		func(
			tx *gorm.DB,
			state controlservice.LockedApplyState,
		) (controlservice.ApplyMutation, error) {
			if state.Plan.Target.Kind != resource.KindExpert ||
				state.Plan.ArtifactKind != resource.KindExpert+"Apply" {
				return controlservice.ApplyMutation{}, control.ErrInvalid
			}
			mutation, err := build(state)
			if err != nil {
				return controlservice.ApplyMutation{}, err
			}
			if err := validateExpertProjection(mutation); err != nil {
				return controlservice.ApplyMutation{}, err
			}
			expertID, err = writeExpertProjection(tx, state, mutation)
			if err != nil {
				return controlservice.ApplyMutation{}, err
			}
			snapshotID = mutation.Projection.WorkerSpecSnapshotID
			return mutation.ApplyMutation, nil
		},
	)
	if errors.Is(err, control.ErrConsumed) {
		return repo.loadAppliedExpert(ctx, scope, planID)
	}
	if err != nil {
		return controlservice.AppliedExpert{}, err
	}
	return controlservice.AppliedExpert{
		Head: head, ExpertID: expertID,
		WorkerSpecSnapshotID: snapshotID,
		ResourceRevision:     head.Revision,
	}, nil
}

func validateExpertProjection(
	mutation controlservice.ExpertApplyMutation,
) error {
	if mutation.Head.Identity.Kind != resource.KindExpert ||
		mutation.Projection.Name == "" ||
		mutation.Projection.WorkerSpecSnapshotID <= 0 ||
		mutation.Revision.WorkerSpecSnapshotID !=
			mutation.Projection.WorkerSpecSnapshotID {
		return control.ErrInvalid
	}
	return nil
}

func (repo *orchestrationResourceRepo) loadAppliedExpert(
	ctx context.Context,
	scope control.Scope,
	planID string,
) (controlservice.AppliedExpert, error) {
	head, err := repo.loadAppliedResourceResult(
		ctx,
		scope,
		planID,
		resource.KindExpert,
		resource.KindExpert+"Apply",
	)
	if err != nil {
		return controlservice.AppliedExpert{}, err
	}
	revision, err := repo.GetRevision(ctx, scope, head.ID, head.Revision)
	if err != nil {
		return controlservice.AppliedExpert{}, err
	}
	var row orchestrationExpertRecord
	err = repo.db.WithContext(ctx).Where(
		"organization_id = ? AND orchestration_resource_id = ?",
		scope.OrganizationID,
		head.ID,
	).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return controlservice.AppliedExpert{}, control.ErrCorrupt
	}
	if err != nil {
		return controlservice.AppliedExpert{}, err
	}
	if row.OrchestrationResourceRevision != head.Revision ||
		row.WorkerSpecSnapshotID != revision.WorkerSpecSnapshotID ||
		row.WorkerSpecSnapshotID <= 0 {
		return controlservice.AppliedExpert{}, control.ErrCorrupt
	}
	return controlservice.AppliedExpert{
		Head: head, ExpertID: row.ID,
		WorkerSpecSnapshotID: row.WorkerSpecSnapshotID,
		ResourceRevision:     row.OrchestrationResourceRevision,
	}, nil
}
