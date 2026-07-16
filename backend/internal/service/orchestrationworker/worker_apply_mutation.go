package orchestrationworker

import (
	"context"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	controlservice "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
)

func buildWorkerApplyMutation(
	ctx context.Context,
	registry *resource.Registry,
	resolver DefinitionResolver,
	state controlservice.LockedApplyState,
) (WorkerApplyMutation, error) {
	if state.Plan.Operation != control.PlanOperationCreate ||
		state.Plan.Target.Kind != resource.KindWorker ||
		state.Plan.ArtifactKind != resource.KindWorker+"Apply" {
		return WorkerApplyMutation{}, control.ErrInvalid
	}
	artifact, err := decodeDefinitionApplyArtifact(state.Plan.ArtifactJSON)
	if err != nil {
		return WorkerApplyMutation{}, err
	}
	manifest, _, err := plannedApplyManifest(registry, state)
	if err != nil {
		return WorkerApplyMutation{}, err
	}
	value, err := registry.DecodeAndValidate(manifest)
	if err != nil {
		return WorkerApplyMutation{}, control.ErrCorrupt
	}
	spec, ok := value.(*resource.WorkerInvocationSpec)
	if !ok || spec == nil {
		return WorkerApplyMutation{}, control.ErrCorrupt
	}
	prompt, err := resolveWorkerPrompt(ctx, resolver, state, spec)
	if err != nil {
		return WorkerApplyMutation{}, err
	}
	mutation, err := buildApplyMutation(
		registry,
		state,
		artifact.WorkerSpecSnapshotID,
	)
	if err != nil {
		return WorkerApplyMutation{}, err
	}
	return WorkerApplyMutation{
		ApplyMutation: mutation,
		Launch: WorkerLaunchProjection{
			WorkerSpecSnapshotID: artifact.WorkerSpecSnapshotID,
			Prompt:               prompt,
			Alias:                spec.Alias,
		},
	}, nil
}

func resolveWorkerPrompt(
	ctx context.Context,
	resolver DefinitionResolver,
	state controlservice.LockedApplyState,
	spec *resource.WorkerInvocationSpec,
) (*string, error) {
	if spec.PromptRef == nil {
		return nil, nil
	}
	pins, err := newPinnedReferenceIndex(
		state.Plan.Scope,
		state.Plan.ResolvedReferences,
	)
	if err != nil {
		return nil, control.ErrCorrupt
	}
	pinned, err := pins.resolve(*spec.PromptRef)
	if err != nil {
		return nil, control.ErrCorrupt
	}
	prompt, err := resolver.ResolvePromptSpec(ctx, state.Plan.Scope, pinned)
	if err != nil {
		return nil, err
	}
	rendered, err := renderPrompt(prompt, spec.Inputs)
	if err != nil {
		return nil, err
	}
	return &rendered, nil
}
