package orchestrationworker

import (
	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	controlservice "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
)

func buildPromptApplyMutation(
	registry *resource.Registry,
	state controlservice.LockedApplyState,
) (controlservice.ApplyMutation, error) {
	if state.Plan.Target.Kind != resource.KindPrompt ||
		state.Plan.ArtifactKind != "PromptSpec" {
		return controlservice.ApplyMutation{}, control.ErrInvalid
	}
	return buildApplyMutation(registry, state, 0)
}
