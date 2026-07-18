package orchestrationresource

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGoalLoopResourceSchemaMatchesDatabaseBounds(t *testing.T) {
	registry := NewRegistry()
	require.NoError(t, RegisterDefinitionSchemas(registry))

	for _, test := range []struct {
		name string
		spec GoalLoopResourceSpec
	}{
		{
			name: "minimum",
			spec: goalLoopBoundsSpec(1, 1, 1, 1),
		},
		{
			name: "maximum",
			spec: goalLoopBoundsSpec(100, 1440, 20, 20),
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := registry.DecodeAndValidate(goalLoopBoundsManifest(t, test.spec))
			require.NoError(t, err)
		})
	}
}

func TestGoalLoopResourceSchemaRejectsValuesOutsideDatabaseBounds(t *testing.T) {
	registry := NewRegistry()
	require.NoError(t, RegisterDefinitionSchemas(registry))

	for _, test := range []struct {
		name  string
		spec  GoalLoopResourceSpec
		field string
	}{
		{"max iterations below", goalLoopBoundsSpec(0, 1, 1, 1), "maxIterations"},
		{"max iterations above", goalLoopBoundsSpec(101, 1, 1, 1), "maxIterations"},
		{"timeout below", goalLoopBoundsSpec(1, 0, 1, 1), "timeoutMinutes"},
		{"timeout above", goalLoopBoundsSpec(1, 1441, 1, 1), "timeoutMinutes"},
		{"no progress below", goalLoopBoundsSpec(1, 1, 0, 1), "noProgressLimit"},
		{"no progress above", goalLoopBoundsSpec(1, 1, 21, 1), "noProgressLimit"},
		{"same error below", goalLoopBoundsSpec(1, 1, 1, 0), "sameErrorLimit"},
		{"same error above", goalLoopBoundsSpec(1, 1, 1, 21), "sameErrorLimit"},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := registry.DecodeAndValidate(goalLoopBoundsManifest(t, test.spec))
			require.ErrorContains(t, err, test.field)
		})
	}
}

func TestGoalLoopResourceSchemaPreservesRequiredDescriptionInCanonicalJSON(t *testing.T) {
	registry := NewRegistry()
	require.NoError(t, RegisterDefinitionSchemas(registry))
	spec := goalLoopBoundsSpec(1, 1, 1, 1)
	spec.Description = ""

	_, err := registry.DecodeAndValidate(goalLoopBoundsManifest(t, spec))
	require.NoError(t, err)
	canonical, err := json.Marshal(spec)
	require.NoError(t, err)
	require.Contains(t, string(canonical), `"description":""`)
}

func goalLoopBoundsSpec(
	maxIterations, timeoutMinutes, noProgressLimit, sameErrorLimit int,
) GoalLoopResourceSpec {
	return GoalLoopResourceSpec{
		WorkerTemplateRef: Reference{
			Kind: KindWorkerTemplate,
			Name: "reviewer",
		},
		Description:         "Repair checkout deterministically",
		Objective:           "Fix checkout",
		AcceptanceCriteria:  []string{"Tests pass"},
		VerificationCommand: "go test ./...",
		MaxIterations:       maxIterations,
		TimeoutMinutes:      timeoutMinutes,
		NoProgressLimit:     noProgressLimit,
		SameErrorLimit:      sameErrorLimit,
		EscalationPolicy:    "pause",
	}
}

func goalLoopBoundsManifest(t *testing.T, spec GoalLoopResourceSpec) Manifest {
	t.Helper()
	raw, err := json.Marshal(spec)
	require.NoError(t, err)
	return Manifest{
		TypeMeta: TypeMeta{
			APIVersion: APIVersionV1Alpha1,
			Kind:       KindGoalLoop,
		},
		Metadata: Metadata{
			Name:      "checkout-loop",
			Namespace: "acme",
		},
		Spec: raw,
	}
}
