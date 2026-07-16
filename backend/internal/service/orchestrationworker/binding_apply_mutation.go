package orchestrationworker

import (
	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	controlservice "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
)

func buildBindingApplyMutation(
	registry *resource.Registry,
	state controlservice.LockedApplyState,
) (controlservice.ApplyMutation, error) {
	if !IsResourceBindingKind(state.Plan.Target.Kind) ||
		state.Plan.ArtifactKind != state.Plan.Target.Kind+"Spec" {
		return controlservice.ApplyMutation{}, control.ErrInvalid
	}
	return buildApplyMutation(registry, state, 0)
}

func IsResourceBindingKind(kind string) bool {
	return resource.IsBindingKind(kind)
}
