package orchestrationworker

import (
	"context"
	"encoding/json"
	"testing"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	controlservice "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResourceBindingPlannerPlansCanonicalBindingSpec(t *testing.T) {
	cases := []struct {
		kind string
		spec any
	}{
		{resource.KindModelBinding, &resource.ModelBindingSpec{ResourceID: 101}},
		{resource.KindRepository, &resource.RepositoryBindingSpec{RepositoryID: 102}},
		{resource.KindSkill, &resource.SkillBindingSpec{SkillID: 103}},
		{resource.KindKnowledgeBase, &resource.KnowledgeBaseBindingSpec{KnowledgeBaseID: 104}},
		{resource.KindEnvironmentBundle, &resource.EnvironmentBundleBindingSpec{EnvironmentBundleID: 105}},
		{resource.KindComputeTarget, &resource.ComputeTargetBindingSpec{ComputeTargetID: 106}},
		{resource.KindResourceProfile, &resource.ResourceProfileBindingSpec{ResourceProfileID: 107}},
	}

	for _, test := range cases {
		t.Run(test.kind, func(t *testing.T) {
			planner, err := NewResourceBindingPlanner(test.kind)
			require.NoError(t, err)
			specJSON, err := json.Marshal(test.spec)
			require.NoError(t, err)

			references, err := planner.References(test.spec)
			require.NoError(t, err)
			assert.Empty(t, references)

			output, err := planner.Plan(context.Background(), controlservice.TargetPlanInput{
				TypedSpec: test.spec,
				Manifest: resource.Manifest{
					TypeMeta: planner.TypeMeta(),
					Spec:     specJSON,
				},
			})
			require.NoError(t, err)
			assert.Equal(t, test.kind+"Spec", output.ArtifactKind)
			assert.Equal(t, BindingSchemaRevision, output.OptionsRevision)
			assert.JSONEq(t, string(specJSON), string(output.ArtifactJSON))
		})
	}
}

func TestResourceBindingPlannerPinsToolModelReference(t *testing.T) {
	planner, err := NewResourceBindingPlanner(resource.KindToolBinding)
	require.NoError(t, err)
	spec := &resource.ToolBindingSpec{ModelRef: resource.Reference{
		Kind: resource.KindModelBinding,
		Name: resourceBindingName("tool-model"),
	}}

	references, err := planner.References(spec)

	require.NoError(t, err)
	require.Len(t, references, 1)
	assert.Equal(t, "/spec/modelRef", references[0].Path)
	assert.Equal(t, spec.ModelRef, references[0].Reference)
}

func TestResourceBindingPlannerRejectsSubstitutedTypedSpec(t *testing.T) {
	planner, err := NewResourceBindingPlanner(resource.KindModelBinding)
	require.NoError(t, err)

	_, err = planner.Plan(context.Background(), controlservice.TargetPlanInput{
		TypedSpec: &resource.RepositoryBindingSpec{RepositoryID: 102},
		Manifest: resource.Manifest{
			TypeMeta: planner.TypeMeta(),
			Spec:     json.RawMessage(`{"resourceId":101}`),
		},
	})

	assert.ErrorIs(t, err, control.ErrCorrupt)
}

func TestNewResourceBindingPlannerRejectsUnsupportedKind(t *testing.T) {
	_, err := NewResourceBindingPlanner(resource.KindWorkerTemplate)
	assert.Error(t, err)
}
