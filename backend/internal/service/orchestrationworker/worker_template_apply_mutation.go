package orchestrationworker

import (
	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	controlservice "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
	"github.com/anthropics/agentsmesh/backend/internal/service/workerdependencyartifact"
)

func buildWorkerTemplateApplyMutation(
	registry *resource.Registry,
	state controlservice.LockedApplyState,
	snapshotID int64,
) (controlservice.ApplyMutation, error) {
	if state.Plan.Target.Kind != resource.KindWorkerTemplate ||
		state.Plan.ArtifactKind != workerdependencyartifact.PlanArtifactKind ||
		snapshotID <= 0 {
		return controlservice.ApplyMutation{}, control.ErrInvalid
	}
	return buildApplyMutation(registry, state, snapshotID)
}
