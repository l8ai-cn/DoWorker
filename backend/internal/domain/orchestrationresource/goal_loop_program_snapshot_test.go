package orchestrationresource

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGoalLoopProgramSnapshotPinsCanonicalCustomBlock(t *testing.T) {
	registry := NewRegistry()
	require.NoError(t, RegisterDefinitionSchemas(registry))

	_, err := registry.DecodeAndValidate(goalLoopProgramManifest(t, goalLoopProgramSpec()))

	require.NoError(t, err)
}

func TestGoalLoopProgramSnapshotRejectsMissingOrMismatchedCustomBlock(t *testing.T) {
	registry := NewRegistry()
	require.NoError(t, RegisterDefinitionSchemas(registry))

	for _, test := range []struct {
		name   string
		mutate func(*GoalLoopProgramSnapshot)
	}{
		{
			name: "missing pin",
			mutate: func(snapshot *GoalLoopProgramSnapshot) {
				snapshot.CustomBlock = nil
			},
		},
		{
			name: "mismatched digest",
			mutate: func(snapshot *GoalLoopProgramSnapshot) {
				snapshot.CustomBlock.DefinitionDigest = strings.Repeat("f", 64)
			},
		},
		{
			name: "noncanonical source",
			mutate: func(snapshot *GoalLoopProgramSnapshot) {
				snapshot.CanonicalSource += "\n"
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			spec := goalLoopProgramSpec()
			test.mutate(spec.LoopProgram)

			_, err := registry.DecodeAndValidate(goalLoopProgramManifest(t, spec))

			require.Error(t, err)
		})
	}
}

func TestGoalLoopProgramSnapshotRejectsExecutionFieldDrift(t *testing.T) {
	registry := NewRegistry()
	require.NoError(t, RegisterDefinitionSchemas(registry))
	spec := goalLoopProgramSpec()
	spec.Objective = "Different objective"

	_, err := registry.DecodeAndValidate(goalLoopProgramManifest(t, spec))

	require.ErrorContains(t, err, "does not match GoalLoop task fields")
}

func goalLoopProgramSpec() GoalLoopResourceSpec {
	tokenBudget := int64(80_000)
	return GoalLoopResourceSpec{
		WorkerTemplateRef: Reference{
			Kind: KindWorkerTemplate,
			Name: "reviewer",
		},
		Description:         "Repair checkout deterministically",
		Objective:           "Fix checkout",
		AcceptanceCriteria:  []string{"Tests pass"},
		VerificationCommand: "go test ./...",
		MaxIterations:       5,
		TokenBudget:         &tokenBudget,
		TimeoutMinutes:      60,
		NoProgressLimit:     3,
		SameErrorLimit:      2,
		EscalationPolicy:    "pause",
		LoopProgram: &GoalLoopProgramSnapshot{
			CanonicalSource: goalLoopProgramSource,
			CustomBlock: &GoalLoopCustomBlockPin{
				NodeID:           "n-ppt-step",
				DefinitionID:     "e54112b4-6a22-4ec4-b14d-dc3ac7c527a4",
				Slug:             "ppt-step",
				Version:          2,
				DefinitionDigest: "a1b2c3d4e5f60718293a4b5c6d7e8f90123456789abcdef0123456789abcdef0",
			},
		},
	}
}

func goalLoopProgramManifest(
	t *testing.T,
	spec GoalLoopResourceSpec,
) Manifest {
	t.Helper()
	manifest := goalLoopBoundsManifest(t, spec)
	manifest.Metadata.Name = "checkout-loop"
	return manifest
}

const goalLoopProgramSource = `@id(n-checkout-loop)
loop checkout-loop {
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
