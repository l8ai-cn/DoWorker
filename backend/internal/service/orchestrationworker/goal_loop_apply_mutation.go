package orchestrationworker

import (
	"strings"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	controlservice "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
)

func buildGoalLoopApplyMutation(
	registry *resource.Registry,
	state controlservice.LockedApplyState,
) (GoalLoopApplyMutation, error) {
	if state.Plan.Target.Kind != resource.KindGoalLoop ||
		state.Plan.ArtifactKind != resource.KindGoalLoop+"Apply" {
		return GoalLoopApplyMutation{}, control.ErrInvalid
	}
	artifact, err := decodeGoalLoopApplyArtifact(state.Plan.ArtifactJSON)
	if err != nil {
		return GoalLoopApplyMutation{}, err
	}
	manifest, _, err := plannedApplyManifest(registry, state)
	if err != nil {
		return GoalLoopApplyMutation{}, err
	}
	value, err := registry.DecodeAndValidate(manifest)
	if err != nil {
		return GoalLoopApplyMutation{}, control.ErrCorrupt
	}
	spec, ok := value.(*resource.GoalLoopResourceSpec)
	if !ok || spec == nil {
		return GoalLoopApplyMutation{}, control.ErrCorrupt
	}
	if err := validateGoalLoopApplyArtifact(spec, artifact); err != nil {
		return GoalLoopApplyMutation{}, err
	}
	mutation, err := buildApplyMutation(
		registry,
		state,
		artifact.WorkerSpecSnapshotID,
	)
	if err != nil {
		return GoalLoopApplyMutation{}, err
	}
	name := strings.TrimSpace(manifest.Metadata.DisplayName)
	if name == "" {
		name = manifest.Metadata.Name.String()
	}
	return GoalLoopApplyMutation{
		ApplyMutation: mutation,
		Projection: GoalLoopApplyProjection{
			Name: name, Description: spec.Description,
			Objective: spec.Objective,
			AcceptanceCriteria: append(
				[]string{},
				spec.AcceptanceCriteria...,
			),
			VerificationCommand:  spec.VerificationCommand,
			MaxIterations:        spec.MaxIterations,
			TokenBudget:          copyTokenBudget(spec.TokenBudget),
			TimeoutMinutes:       spec.TimeoutMinutes,
			NoProgressLimit:      spec.NoProgressLimit,
			SameErrorLimit:       spec.SameErrorLimit,
			EscalationPolicy:     spec.EscalationPolicy,
			WorkerSpecSnapshotID: artifact.WorkerSpecSnapshotID,
		},
	}, nil
}

func copyTokenBudget(value *int64) *int64 {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}
