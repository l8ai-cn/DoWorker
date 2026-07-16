package infra

import (
	"github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	orchestrationservice "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
	"gorm.io/gorm"
)

func writeApplyMutation(
	tx *gorm.DB,
	state orchestrationservice.LockedApplyState,
	mutation orchestrationservice.ApplyMutation,
) error {
	appliedPlan, err := state.Plan.Apply(
		state.AppliedAt,
		state.Plan.ActorID,
		mutation.Head.ID,
		mutation.Head.Identity,
		mutation.Head.ResourceVersion,
		mutation.Head.Revision,
	)
	if err != nil {
		return err
	}
	if err := consumeApplyPlan(tx, appliedPlan); err != nil {
		return err
	}
	if state.Head == nil {
		return insertApplyMutation(tx, state.Plan.Scope, mutation)
	}
	return updateApplyMutation(tx, state, mutation)
}

func consumeApplyPlan(
	tx *gorm.DB,
	plan orchestrationcontrol.Plan,
) error {
	result := tx.Model(&orchestrationPlanRecord{}).
		Where(
			"id = ? AND organization_id = ? AND actor_id = ? AND consumed_at IS NULL",
			plan.ID,
			plan.Scope.OrganizationID,
			plan.ActorID,
		).
		Updates(map[string]any{
			"consumed_at":             *plan.ConsumedAt,
			"consumed_by_id":          plan.ConsumedByID,
			"consumption_result":      string(plan.Status),
			"result_resource_id":      plan.ResultResourceID,
			"result_resource_uid":     plan.ResultIdentity.UID,
			"result_resource_version": plan.ResultResourceVersion,
			"result_revision":         plan.ResultRevision,
			"result_json":             plan.ResultJSON,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return orchestrationcontrol.ErrConsumed
	}
	return nil
}

func insertApplyMutation(
	tx *gorm.DB,
	scope orchestrationcontrol.Scope,
	mutation orchestrationservice.ApplyMutation,
) error {
	head, err := orchestrationResourceRecordFromDomain(mutation.Head, scope)
	if err != nil {
		return err
	}
	if err := tx.Create(&head).Error; err != nil {
		return err
	}
	revision, err := orchestrationRevisionRecordFromDomain(
		mutation.Revision,
		scope,
	)
	if err != nil {
		return err
	}
	return tx.Create(&revision).Error
}

func updateApplyMutation(
	tx *gorm.DB,
	state orchestrationservice.LockedApplyState,
	mutation orchestrationservice.ApplyMutation,
) error {
	revision, err := orchestrationRevisionRecordFromDomain(
		mutation.Revision,
		state.Plan.Scope,
	)
	if err != nil {
		return err
	}
	if err := tx.Create(&revision).Error; err != nil {
		return err
	}
	head, err := orchestrationResourceRecordFromDomain(
		mutation.Head,
		state.Plan.Scope,
	)
	if err != nil {
		return err
	}
	result := tx.Model(&orchestrationResourceRecord{}).
		Where(
			"organization_id = ? AND id = ? AND resource_version = ?",
			state.Plan.Scope.OrganizationID,
			head.ID,
			state.Head.ResourceVersion,
		).
		Updates(map[string]any{
			"display_name": head.DisplayName, "labels": head.Labels,
			"status": head.Status, "generation": head.Generation,
			"resource_version": head.ResourceVersion,
			"active_revision":  head.ActiveRevision,
			"updated_by_id":    head.UpdatedByID, "updated_at": head.UpdatedAt,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return orchestrationcontrol.ErrStale
	}
	return nil
}
