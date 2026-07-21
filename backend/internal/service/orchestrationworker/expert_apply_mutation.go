package orchestrationworker

import (
	"context"
	"strings"

	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	resource "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationresource"
	controlservice "github.com/l8ai-cn/agentcloud/backend/internal/service/orchestrationcontrol"
)

func buildExpertApplyMutation(
	ctx context.Context,
	registry *resource.Registry,
	resolver DefinitionResolver,
	state controlservice.LockedApplyState,
) (ExpertApplyMutation, error) {
	if state.Plan.Target.Kind != resource.KindExpert ||
		state.Plan.ArtifactKind != resource.KindExpert+"Apply" {
		return ExpertApplyMutation{}, control.ErrInvalid
	}
	artifact, err := decodeDefinitionApplyArtifact(state.Plan.ArtifactJSON)
	if err != nil {
		return ExpertApplyMutation{}, err
	}
	manifest, _, err := plannedApplyManifest(registry, state)
	if err != nil {
		return ExpertApplyMutation{}, err
	}
	value, err := registry.DecodeAndValidate(manifest)
	if err != nil {
		return ExpertApplyMutation{}, control.ErrCorrupt
	}
	spec, ok := value.(*resource.ExpertResourceSpec)
	if !ok || spec == nil {
		return ExpertApplyMutation{}, control.ErrCorrupt
	}
	prompt, err := resolveExpertPrompt(ctx, resolver, state, spec)
	if err != nil {
		return ExpertApplyMutation{}, err
	}
	mutation, err := buildApplyMutation(
		registry,
		state,
		artifact.WorkerSpecSnapshotID,
	)
	if err != nil {
		return ExpertApplyMutation{}, err
	}
	name := strings.TrimSpace(manifest.Metadata.DisplayName)
	if name == "" {
		name = manifest.Metadata.Name.String()
	}
	return ExpertApplyMutation{
		ApplyMutation: mutation,
		Projection: ExpertApplyProjection{
			Name: name, Description: spec.Description,
			Category: spec.Category, ReleaseNotes: spec.ReleaseNotes,
			Prompt: prompt, WorkerSpecSnapshotID: artifact.WorkerSpecSnapshotID,
		},
	}, nil
}

func resolveExpertPrompt(
	ctx context.Context,
	resolver DefinitionResolver,
	state controlservice.LockedApplyState,
	spec *resource.ExpertResourceSpec,
) (string, error) {
	if spec.PromptRef == nil {
		return "", nil
	}
	pins, err := newPinnedReferenceIndex(
		state.Plan.Scope,
		state.Plan.ResolvedReferences,
	)
	if err != nil {
		return "", control.ErrCorrupt
	}
	pinned, err := pins.resolve(*spec.PromptRef)
	if err != nil {
		return "", control.ErrCorrupt
	}
	prompt, err := resolver.ResolvePromptSpec(ctx, state.Plan.Scope, pinned)
	if err != nil {
		return "", err
	}
	return renderPrompt(prompt, map[string]string{})
}
