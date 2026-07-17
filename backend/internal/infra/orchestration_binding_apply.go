package infra

import (
	"context"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	controlservice "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
	"gorm.io/gorm"
)

func (repo *orchestrationResourceRepo) RunBindingApplyTransaction(
	ctx context.Context,
	scope control.Scope,
	planID string,
	build controlservice.BindingApplyBuilder,
) (control.ResourceHead, error) {
	if build == nil {
		return control.ResourceHead{}, control.ErrInvalid
	}
	if err := validateApplyCoordinates(scope, planID); err != nil {
		return control.ResourceHead{}, err
	}
	return repo.runApplyTransaction(
		ctx,
		scope,
		planID,
		func(
			_ *gorm.DB,
			state controlservice.LockedApplyState,
		) (controlservice.ApplyMutation, error) {
			if !resource.IsBindingKind(state.Plan.Target.Kind) ||
				state.Plan.ArtifactKind != state.Plan.Target.Kind+"Spec" {
				return controlservice.ApplyMutation{}, control.ErrInvalid
			}
			return build(state)
		},
	)
}
