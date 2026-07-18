package workerdependencyartifact

import (
	"testing"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecodeApplyPlanReturnsBoundImmutableDocuments(t *testing.T) {
	input := validInput(t)
	addWorkspaceDependencies(t, &input)
	built, err := Build(input)
	require.NoError(t, err)
	plan := controlPlanForArtifact(t, input, built)

	artifact, err := DecodeApplyPlan(plan)
	require.NoError(t, err)

	assert.Equal(t, built.WorkerSpecJSON(), artifact.WorkerSpecJSON())
	assert.Equal(t, built.JSON(), artifact.DependenciesJSON())
	assert.Equal(t, built.Digest(), artifact.DependenciesDigest())
	require.Len(t, artifact.SecretReferences(), 1)

	workerSpec := artifact.WorkerSpecJSON()
	workerSpec[0] = 'x'
	assert.NotEqual(t, workerSpec, artifact.WorkerSpecJSON())
	secrets := artifact.SecretReferences()
	secrets[0].BundleKey = "MUTATED"
	assert.Equal(t, "CURSOR_API_KEY", artifact.SecretReferences()[0].BundleKey)
}

func TestDecodeApplyPlanRejectsPlanEnvelopeSubstitution(t *testing.T) {
	tests := []struct {
		name  string
		plan  func(*control.Plan)
		match string
	}{
		{
			name: "wrong artifact kind",
			plan: func(plan *control.Plan) {
				plan.ArtifactKind = "WorkerSpec"
			},
			match: "WorkerTemplate build",
		},
		{
			name: "missing direct reference",
			plan: func(plan *control.Plan) {
				plan.ResolvedReferences = plan.ResolvedReferences[:2]
			},
			match: "unplanned reference",
		},
		{
			name: "wrong target",
			plan: func(plan *control.Plan) {
				plan.Target.Kind = resource.KindWorker
			},
			match: "authenticated target",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := validInput(t)
			built, err := Build(input)
			require.NoError(t, err)
			plan := controlPlanForArtifact(t, input, built)
			test.plan(&plan)
			plan.PlanHash, err = control.ComputePlanHash(plan.HashInput())
			require.NoError(t, err)

			_, err = DecodeApplyPlan(plan)

			require.ErrorContains(t, err, test.match)
		})
	}
}
