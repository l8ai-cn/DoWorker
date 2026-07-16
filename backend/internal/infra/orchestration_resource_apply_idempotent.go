package infra

import (
	"context"
	"errors"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	controlservice "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
)

func (repo *orchestrationResourceRepo) runIdempotentApplyTransaction(
	ctx context.Context,
	scope control.Scope,
	planID string,
	expectedKind string,
	expectedArtifactKind string,
	build controlservice.ApplyBuilder,
) (control.ResourceHead, error) {
	if err := validateApplyRequest(scope, planID, build); err != nil {
		return control.ResourceHead{}, err
	}
	head, err := repo.RunApplyTransaction(
		ctx,
		scope,
		planID,
		func(state controlservice.LockedApplyState) (
			controlservice.ApplyMutation,
			error,
		) {
			if state.Plan.Target.Kind != expectedKind ||
				state.Plan.ArtifactKind != expectedArtifactKind {
				return controlservice.ApplyMutation{}, control.ErrInvalid
			}
			return build(state)
		},
	)
	if !errors.Is(err, control.ErrConsumed) {
		return head, err
	}
	return repo.loadAppliedResourceResult(
		ctx,
		scope,
		planID,
		expectedKind,
		expectedArtifactKind,
	)
}

func (repo *orchestrationResourceRepo) loadAppliedResourceResult(
	ctx context.Context,
	scope control.Scope,
	planID string,
	expectedKind string,
	expectedArtifactKind string,
) (control.ResourceHead, error) {
	plan, err := repo.GetPlan(ctx, scope, planID)
	if err != nil {
		return control.ResourceHead{}, err
	}
	if plan.Target.Kind != expectedKind ||
		plan.ArtifactKind != expectedArtifactKind {
		return control.ResourceHead{}, control.ErrInvalid
	}
	if plan.Status != control.PlanStatusApplied || plan.ResultIdentity == nil {
		return control.ResourceHead{}, control.ErrConsumed
	}
	head, err := repo.GetResource(ctx, scope, plan.Target)
	if err != nil {
		return control.ResourceHead{}, err
	}
	if head.ID != plan.ResultResourceID ||
		head.Identity != *plan.ResultIdentity ||
		head.Revision != plan.ResultRevision ||
		head.ResourceVersion != plan.ResultResourceVersion {
		return control.ResourceHead{}, control.ErrCorrupt
	}
	return head, nil
}
