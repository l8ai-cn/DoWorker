package orchestrationworker

import (
	"context"
	"strings"
	"time"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	workflowDomain "github.com/anthropics/agentsmesh/backend/internal/domain/workflow"
	controlservice "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
	"github.com/robfig/cron/v3"
)

var workflowCronParser = cron.NewParser(
	cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow,
)

func buildWorkflowApplyMutation(
	ctx context.Context,
	registry *resource.Registry,
	resolver DefinitionResolver,
	state controlservice.LockedApplyState,
	status string,
) (WorkflowApplyMutation, error) {
	if state.Plan.Target.Kind != resource.KindWorkflow ||
		state.Plan.ArtifactKind != resource.KindWorkflow+"Apply" {
		return WorkflowApplyMutation{}, control.ErrInvalid
	}
	artifact, err := decodeDefinitionApplyArtifact(state.Plan.ArtifactJSON)
	if err != nil {
		return WorkflowApplyMutation{}, err
	}
	manifest, _, err := plannedApplyManifest(registry, state)
	if err != nil {
		return WorkflowApplyMutation{}, err
	}
	value, err := registry.DecodeAndValidate(manifest)
	if err != nil {
		return WorkflowApplyMutation{}, control.ErrCorrupt
	}
	spec, ok := value.(*resource.WorkflowResourceSpec)
	if !ok || spec == nil {
		return WorkflowApplyMutation{}, control.ErrCorrupt
	}
	prompt, err := resolveWorkflowPrompt(ctx, resolver, state, spec)
	if err != nil {
		return WorkflowApplyMutation{}, err
	}
	nextRunAt, err := workflowNextRunAt(spec.CronExpression, state.AppliedAt)
	if err != nil {
		return WorkflowApplyMutation{}, control.ErrCorrupt
	}
	if status == workflowDomain.StatusDisabled {
		nextRunAt = nil
	}
	mutation, err := buildApplyMutation(
		registry,
		state,
		artifact.WorkerSpecSnapshotID,
	)
	if err != nil {
		return WorkflowApplyMutation{}, err
	}
	name := strings.TrimSpace(manifest.Metadata.DisplayName)
	if name == "" {
		name = manifest.Metadata.Name.String()
	}
	return WorkflowApplyMutation{
		ApplyMutation: mutation,
		Projection: WorkflowApplyProjection{
			Name: name, Prompt: prompt,
			Status:               status,
			ExecutionMode:        spec.ExecutionMode,
			CronExpression:       spec.CronExpression,
			SandboxStrategy:      spec.SandboxStrategy,
			SessionPersistence:   spec.SessionPersistence,
			ConcurrencyPolicy:    spec.ConcurrencyPolicy,
			MaxConcurrentRuns:    spec.MaxConcurrentRuns,
			MaxRetainedRuns:      spec.MaxRetainedRuns,
			TimeoutMinutes:       spec.TimeoutMinutes,
			IdleTimeoutSeconds:   spec.IdleTimeoutSeconds,
			CallbackURL:          spec.CallbackURL,
			WorkerSpecSnapshotID: artifact.WorkerSpecSnapshotID,
			NextRunAt:            nextRunAt,
		},
	}, nil
}

func resolveWorkflowPrompt(
	ctx context.Context,
	resolver DefinitionResolver,
	state controlservice.LockedApplyState,
	spec *resource.WorkflowResourceSpec,
) (string, error) {
	pins, err := newPinnedReferenceIndex(
		state.Plan.Scope,
		state.Plan.ResolvedReferences,
	)
	if err != nil {
		return "", control.ErrCorrupt
	}
	pinned, err := pins.resolve(spec.PromptRef)
	if err != nil {
		return "", control.ErrCorrupt
	}
	prompt, err := resolver.ResolvePromptSpec(ctx, state.Plan.Scope, pinned)
	if err != nil {
		return "", err
	}
	return renderPrompt(prompt, spec.Inputs)
}

func workflowNextRunAt(expression string, from time.Time) (*time.Time, error) {
	if expression == "" {
		return nil, nil
	}
	schedule, err := workflowCronParser.Parse(expression)
	if err != nil {
		return nil, err
	}
	next := schedule.Next(from)
	return &next, nil
}
