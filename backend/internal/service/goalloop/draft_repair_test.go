package goalloop

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDraftGeneratorRepairsExactIntegerDiagnostic(t *testing.T) {
	source := strings.Replace(draftLoopSource, "max: 5", "max: 6", 1)
	resources := &draftResourceResolver{resolved: draftResolvedResource(t)}
	provider := &draftJSONGenerator{output: []byte(`{"value":4}`)}
	generator := NewDraftGenerator(resources, provider)

	proposal, err := generator.Repair(
		context.Background(),
		DraftGenerationScope{OrganizationID: 7, UserID: 9},
		DraftRepairInput{
			Source: source, ModelResourceID: 42, Locale: "zh-CN",
			DiagnosticCode: "loop.repeat.max-exceeds-limit",
			NodeID:         "n-fix-cycle", FieldPath: "repeat.max",
			Prompt: "保持修复轮数充足",
		},
	)

	require.NoError(t, err)
	require.Equal(t, int64(6), proposal.Patch.OldValue)
	require.Equal(t, int64(4), proposal.Patch.NewValue)
	require.Equal(t, int64(4), proposal.Program.Loop.Repeat.Max)
	require.Equal(t, "pnpm test", proposal.Program.Loop.Repeat.Verifier.Command)
	require.NotContains(t, provider.userPrompt, "pnpm test")
	require.NotContains(t, provider.userPrompt, source)
	require.Equal(t, int64(42), resources.resourceID)
}

func TestDraftGeneratorRejectsStaleOrUnsupportedRepairTarget(t *testing.T) {
	source := strings.Replace(draftLoopSource, "max: 5", "max: 6", 1)
	tests := []struct {
		name      string
		code      string
		fieldPath string
		want      error
	}{
		{
			name: "stale target", code: "loop.value.out-of-range",
			fieldPath: "repeat.max", want: ErrDraftRepairTargetStale,
		},
		{
			name: "unsupported field", code: "loop.repeat.max-exceeds-limit",
			fieldPath: "verifier.command", want: ErrDraftRepairTargetStale,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			provider := &draftJSONGenerator{output: []byte(`{"value":4}`)}
			generator := NewDraftGenerator(
				&draftResourceResolver{resolved: draftResolvedResource(t)},
				provider,
			)

			_, err := generator.Repair(
				context.Background(),
				DraftGenerationScope{OrganizationID: 7, UserID: 9},
				DraftRepairInput{
					Source: source, ModelResourceID: 42, Locale: "zh-CN",
					DiagnosticCode: test.code, NodeID: "n-fix-cycle",
					FieldPath: test.fieldPath,
				},
			)

			require.ErrorIs(t, err, test.want)
			require.False(t, provider.called)
		})
	}
}

func TestDraftGeneratorRejectsInvalidRepairValue(t *testing.T) {
	source := strings.Replace(draftLoopSource, "max: 5", "max: 6", 1)
	generator := NewDraftGenerator(
		&draftResourceResolver{resolved: draftResolvedResource(t)},
		&draftJSONGenerator{output: []byte(`{"value":100}`)},
	)

	_, err := generator.Repair(
		context.Background(),
		DraftGenerationScope{OrganizationID: 7, UserID: 9},
		DraftRepairInput{
			Source: source, ModelResourceID: 42, Locale: "zh-CN",
			DiagnosticCode: "loop.repeat.max-exceeds-limit",
			NodeID:         "n-fix-cycle", FieldPath: "repeat.max",
		},
	)

	require.ErrorIs(t, err, ErrGeneratedDraftInvalid)
}

func TestDraftGeneratorRejectsSecretLikeRepairIntentBeforeProvider(t *testing.T) {
	source := strings.Replace(draftLoopSource, "max: 5", "max: 6", 1)
	provider := &draftJSONGenerator{}
	generator := NewDraftGenerator(
		&draftResourceResolver{resolved: draftResolvedResource(t)},
		provider,
	)

	_, err := generator.Repair(
		context.Background(),
		DraftGenerationScope{OrganizationID: 7, UserID: 9},
		DraftRepairInput{
			Source: source, ModelResourceID: 42, Locale: "zh-CN",
			DiagnosticCode: "loop.repeat.max-exceeds-limit",
			NodeID:         "n-fix-cycle", FieldPath: "repeat.max",
			Prompt: "Bearer eyJhbGciOiJIUzI1NiJ9.payload.signature",
		},
	)

	require.ErrorIs(t, err, ErrDraftContainsSecret)
	require.False(t, provider.called)
}
