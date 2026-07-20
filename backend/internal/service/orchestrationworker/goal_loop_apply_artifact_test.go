package orchestrationworker

import (
	"context"
	"strings"
	"testing"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	controlservice "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
	"github.com/stretchr/testify/require"
)

func TestDefinitionPlannerPinsGoalLoopProgramSnapshot(t *testing.T) {
	planner, err := NewDefinitionPlanner(
		resource.KindGoalLoop,
		&definitionResolverStub{snapshotID: 91},
		&goalLoopSlugCheckerStub{},
	)
	require.NoError(t, err)
	scope := definitionScope()
	workerRef := reference(resource.KindWorkerTemplate, "reviewer")
	spec := goalLoopArtifactSpec(workerRef)
	rawArtifact, err := definitionApplyArtifact(
		resource.KindGoalLoop,
		91,
		spec,
	)
	require.NoError(t, err)
	require.NotEmpty(t, rawArtifact)
	output, err := planner.Plan(context.Background(), controlservice.TargetPlanInput{
		Scope: scope,
		Manifest: resource.Manifest{
			TypeMeta: planner.TypeMeta(),
			Metadata: resource.Metadata{Name: "checkout-loop"},
		},
		TypedSpec: spec,
		ResolvedReferences: []control.ResolvedReference{
			resolvedReference(scope, workerRef, 4),
		},
	})

	require.NoError(t, err)
	artifact, err := decodeGoalLoopApplyArtifact(output.ArtifactJSON)
	require.NoError(t, err)
	require.NotEmpty(t, artifact.LoopProgramDigest)
	require.NoError(t, validateGoalLoopApplyArtifact(spec, artifact))
}

func TestGoalLoopApplyRejectsSubstitutedProgramSnapshot(t *testing.T) {
	state := goalLoopApplyCreateState(t)
	spec := goalLoopArtifactSpec(resource.Reference{
		Kind: resource.KindWorkerTemplate,
		Name: "review-worker",
	})
	manifest := resource.Manifest{
		TypeMeta: resource.TypeMeta{
			APIVersion: resource.APIVersionV1Alpha1,
			Kind:       resource.KindGoalLoop,
		},
		Metadata: resource.Metadata{
			Name: "checkout-recovery", Namespace: "team-alpha",
			DisplayName: "Checkout Recovery", Labels: map[string]string{},
		},
		Spec: canonicalApplyJSON(t, spec),
	}
	state.Plan.CanonicalManifest = canonicalApplyJSON(t, manifest)
	state.Plan.ArtifactJSON = canonicalApplyJSON(t, GoalLoopApplyArtifact{
		WorkerSpecSnapshotID: 901,
		LoopProgramDigest:    strings.Repeat("f", 64),
	})
	state.Plan.ArtifactDigest = digestApplyJSON(t, state.Plan.ArtifactJSON)

	_, err := buildGoalLoopApplyMutation(workerApplyRegistry(t), state)

	require.ErrorIs(t, err, control.ErrCorrupt)
}

func goalLoopArtifactSpec(workerRef resource.Reference) *resource.GoalLoopResourceSpec {
	tokenBudget := int64(80_000)
	return &resource.GoalLoopResourceSpec{
		WorkerTemplateRef:   workerRef,
		Description:         "Restore checkout reliability",
		Objective:           "Fix checkout",
		AcceptanceCriteria:  []string{"Tests pass"},
		VerificationCommand: "go test ./...",
		MaxIterations:       5,
		TokenBudget:         &tokenBudget,
		TimeoutMinutes:      60,
		NoProgressLimit:     3,
		SameErrorLimit:      2,
		EscalationPolicy:    "pause",
		LoopProgram: &resource.GoalLoopProgramSnapshot{
			CanonicalSource: goalLoopArtifactSource,
			CustomBlock: &resource.GoalLoopCustomBlockPin{
				NodeID:           "n-ppt-step",
				DefinitionID:     "e54112b4-6a22-4ec4-b14d-dc3ac7c527a4",
				Slug:             "ppt-step",
				Version:          2,
				DefinitionDigest: "a1b2c3d4e5f60718293a4b5c6d7e8f90123456789abcdef0123456789abcdef0",
			},
		},
	}
}

const goalLoopArtifactSource = `@id(n-checkout-recovery)
loop checkout-recovery {
  limits(iterations: 5, tokens: 80000, timeout: 60m, no_progress: 3, same_error: 2)
  @id(n-fix-cycle)
  repeat fix-cycle(max: 5, until: tests.passed) {
    custom_block(node_id: n-ppt-step, definition_id: "e54112b4-6a22-4ec4-b14d-dc3ac7c527a4", slug: ppt-step, version: 2, digest: "a1b2c3d4e5f60718293a4b5c6d7e8f90123456789abcdef0123456789abcdef0")
    @id(n-ppt-step-task)
    agent ppt-step-task { prompt """Fix checkout""" }
    @id(n-tests)
    verify tests { command "go test ./..." accept "Tests pass" }
  }
  on_failure pause
}`
