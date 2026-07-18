package workerdependencyartifact

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/anthropics/agentsmesh/backend/internal/domain/airesource"
	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	"github.com/anthropics/agentsmesh/backend/internal/domain/workerdependency"
	"github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildBindsCanonicalArtifactToPlanReferenceClosure(t *testing.T) {
	input := validInput(t)
	input.PlanReferences[0], input.PlanReferences[2] =
		input.PlanReferences[2], input.PlanReferences[0]

	artifact, err := Build(input)
	require.NoError(t, err)

	decoded, err := workerdependency.Decode(artifact.JSON())
	require.NoError(t, err)
	assert.Equal(
		t,
		input.Dependencies.ToolModels[0].Binding.UID,
		decoded.Models.Tools[0].Binding.UID,
	)
	assert.Equal(t, artifact.WorkerSpecDigest(), decoded.Worker.SpecDigest)
	expectedDigest, err := workerdependency.Digest(decoded)
	require.NoError(t, err)
	assert.Equal(t, expectedDigest, artifact.Digest())
	decodedSpec, err := workerspec.DecodeSpec(artifact.WorkerSpecJSON())
	require.NoError(t, err)
	assert.Equal(t, input.WorkerSpec.Runtime.WorkerType, decodedSpec.Runtime.WorkerType)

	first := artifact.JSON()
	first[0] = 'x'
	assert.NotEqual(t, first, artifact.JSON())
}

func TestBuildFailsClosedWhenDirectPlanClosureIsIncomplete(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Input)
		match  string
	}{
		{
			name: "artifact contains unplanned reference",
			mutate: func(input *Input) {
				input.PlanReferences = input.PlanReferences[:2]
			},
			match: "unplanned reference",
		},
		{
			name: "Plan contains unused reference",
			mutate: func(input *Input) {
				input.PlanReferences = append(
					input.PlanReferences,
					resolvedReference(resource.KindSkill, "unused-skill"),
				)
			},
			match: "absent from worker dependency artifact",
		},
		{
			name: "Plan reference digest was substituted",
			mutate: func(input *Input) {
				input.PlanReferences[0].Digest = workerdependency.TextDigest("changed")
			},
			match: "absent from worker dependency artifact",
		},
		{
			name: "Plan repeats one immutable reference",
			mutate: func(input *Input) {
				input.PlanReferences = append(
					input.PlanReferences,
					input.PlanReferences[0],
				)
			},
			match: "duplicate planned reference identity revision",
		},
		{
			name: "Plan repeats identity revision with another digest",
			mutate: func(input *Input) {
				duplicate := input.PlanReferences[0]
				duplicate.Digest = workerdependency.TextDigest("substituted")
				input.PlanReferences = append(input.PlanReferences, duplicate)
			},
			match: "duplicate planned reference identity revision",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := validInput(t)
			test.mutate(&input)

			artifact, err := Build(input)

			require.ErrorContains(t, err, test.match)
			assert.Empty(t, artifact.JSON())
			assert.Empty(t, artifact.Digest())
		})
	}
}

func TestBuildRequiresExactToolBindingModelRelationship(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Input)
		match  string
	}{
		{
			name: "missing transitive resolution",
			mutate: func(input *Input) {
				input.Dependencies.ToolModels = nil
			},
			match: "tool models do not match WorkerSpec",
		},
		{
			name: "wrong child kind",
			mutate: func(input *Input) {
				input.Dependencies.ToolModels[0].Model.reference.Kind =
					resource.KindToolBinding
			},
			match: "ModelBinding dependency reference kind",
		},
		{
			name: "wrong parent kind",
			mutate: func(input *Input) {
				input.Dependencies.ToolModels[0].Binding.Kind =
					resource.KindModelBinding
			},
			match: "ToolBinding dependency reference kind",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := validInput(t)
			test.mutate(&input)

			_, err := Build(input)

			require.ErrorContains(t, err, test.match)
		})
	}
}

func TestBuildRejectsScopeSubstitution(t *testing.T) {
	input := validInput(t)
	input.Scope.OrganizationID = 0

	_, err := Build(input)

	require.Error(t, err)
}

func TestBuildReturnsDigestOfTheEncodedBytes(t *testing.T) {
	input := validInput(t)

	artifact, err := Build(input)
	require.NoError(t, err)
	canonicalDigest, err := control.DigestCanonicalJSON(artifact.JSON())
	require.NoError(t, err)

	assert.Equal(t, canonicalDigest, artifact.Digest())
}

func TestBuildRejectsOversizedResolvedFactsBeforeMaterialization(t *testing.T) {
	input := validInput(t)
	input.Dependencies.RuntimeBundles = []RuntimeBundleResolution{{
		Values: make(
			[]RuntimeValueResolution,
			workerdependency.MaxDocumentBytes/16+1,
		),
	}}

	_, err := Build(input)

	require.True(t, errors.Is(err, workerdependency.ErrDocumentTooLarge))
}

func TestBuildRejectsOversizedEnumFieldsBeforeMaterialization(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*testing.T, *Input, string)
	}{
		{
			name: "modality",
			mutate: func(_ *testing.T, input *Input, value string) {
				model := &input.Dependencies.ToolModels[0].Model
				model.Modalities = []airesource.Modality{airesource.Modality(value)}
			},
		},
		{
			name: "capability",
			mutate: func(_ *testing.T, input *Input, value string) {
				model := &input.Dependencies.ToolModels[0].Model
				model.Capabilities = []airesource.Capability{
					airesource.Capability(value),
				}
			},
		},
		{
			name: "resolved knowledge mount mode",
			mutate: func(t *testing.T, input *Input, value string) {
				addWorkspaceDependencies(t, input)
				input.Dependencies.KnowledgeBases[0].Mode =
					workerspec.KnowledgeMountMode(value)
			},
		},
		{
			name: "resolved placement policy",
			mutate: func(_ *testing.T, input *Input, value string) {
				input.Dependencies.Placement.Spec.Policy =
					workerspec.PlacementPolicy(value)
			},
		},
		{
			name: "resolved compute target kind",
			mutate: func(_ *testing.T, input *Input, value string) {
				input.Dependencies.Placement.Spec.ComputeTarget.Kind =
					workerspec.ComputeTargetKind(value)
			},
		},
		{
			name: "resolved deployment mode",
			mutate: func(_ *testing.T, input *Input, value string) {
				input.Dependencies.Placement.Spec.DeploymentMode =
					workerspec.DeploymentMode(value)
			},
		},
		{
			name: "worker interaction mode",
			mutate: func(_ *testing.T, input *Input, value string) {
				input.WorkerSpec.TypeConfig.InteractionMode =
					workerspec.InteractionMode(value)
			},
		},
		{
			name: "worker automation level",
			mutate: func(_ *testing.T, input *Input, value string) {
				input.WorkerSpec.TypeConfig.AutomationLevel =
					workerspec.AutomationLevel(value)
			},
		},
		{
			name: "worker knowledge mount mode",
			mutate: func(t *testing.T, input *Input, value string) {
				addWorkspaceDependencies(t, input)
				input.WorkerSpec.Workspace.KnowledgeMounts[0].Mode =
					workerspec.KnowledgeMountMode(value)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := validInput(t)
			value := strings.Repeat("x", workerdependency.MaxDocumentBytes+1)
			test.mutate(t, &input, value)

			_, err := Build(input)

			require.True(t, errors.Is(err, workerdependency.ErrDocumentTooLarge))
		})
	}
}

func TestBuildRejectsNonTreeJSONValuesBeforeMarshaling(t *testing.T) {
	tests := map[string]any{
		"struct": struct{ Payload string }{
			Payload: strings.Repeat("x", workerdependency.MaxDocumentBytes+1),
		},
		"json marshaler": rejectedJSONMarshaler("small"),
		"text marshaler": rejectedTextMarshaler("small"),
		"byte slice":     []byte("small"),
	}
	for name, value := range tests {
		t.Run(name, func(t *testing.T) {
			input := validInput(t)
			input.WorkerSpec.TypeConfig.Values["value"] = value

			_, err := Build(input)

			require.True(t, errors.Is(err, workerdependency.ErrDocumentTooLarge))
		})
	}
}

func TestBuildRejectsOversizedPlanAndWorkerSpecBeforeNormalization(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Input)
	}{
		{
			name: "Plan references",
			mutate: func(input *Input) {
				input.PlanReferences = make(
					[]control.ResolvedReference,
					workerdependency.MaxDocumentBytes/16+1,
				)
			},
		},
		{
			name: "WorkerSpec values",
			mutate: func(input *Input) {
				input.WorkerSpec.TypeConfig.Values["large"] =
					strings.Repeat("x", workerdependency.MaxDocumentBytes+1)
			},
		},
		{
			name: "WorkerSpec config cardinality",
			mutate: func(input *Input) {
				values := make(
					map[string]any,
					workerdependency.MaxDocumentBytes/16+1,
				)
				for index := 0; index < workerdependency.MaxDocumentBytes/16+1; index++ {
					values[fmt.Sprintf("key-%d", index)] = index
				}
				input.WorkerSpec.TypeConfig.Values = values
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := validInput(t)
			test.mutate(&input)

			_, err := Build(input)

			require.True(t, errors.Is(err, workerdependency.ErrDocumentTooLarge))
		})
	}
}

type rejectedJSONMarshaler string

func (rejectedJSONMarshaler) MarshalJSON() ([]byte, error) {
	panic("custom JSON marshaler must not run")
}

type rejectedTextMarshaler string

func (rejectedTextMarshaler) MarshalText() ([]byte, error) {
	panic("custom text marshaler must not run")
}
