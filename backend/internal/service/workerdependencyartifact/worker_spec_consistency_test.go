package workerdependencyartifact

import (
	"testing"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	"github.com/anthropics/agentsmesh/backend/internal/domain/workerdependency"
	"github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildAcceptsEveryDirectDependencyClass(t *testing.T) {
	input := validInput(t)
	addWorkspaceDependencies(t, &input)

	artifact, err := Build(input)
	require.NoError(t, err)

	document, err := workerdependency.Decode(artifact.JSON())
	require.NoError(t, err)
	assert.NotNil(t, document.Models.Primary)
	assert.NotNil(t, document.Repository)
	assert.Len(t, document.Skills, 1)
	assert.Len(t, document.KnowledgeBases, 1)
	assert.Len(t, document.RuntimeBundles, 2)
	assert.Len(t, document.SecretReferences, 1)
}

func TestBuildRejectsWorkerSpecDependencySubstitution(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Input)
		match  string
	}{
		{
			name: "worker identity",
			mutate: func(input *Input) {
				input.WorkerSpec.Runtime.WorkerType.DefinitionHash =
					workerdependency.TextDigest("other")[len("sha256:"):]
			},
			match: "worker definition does not match",
		},
		{
			name: "placement",
			mutate: func(input *Input) {
				input.WorkerSpec.Placement.ComputeTarget.ID = 99
			},
			match: "placement does not match",
		},
		{
			name: "primary model",
			mutate: func(input *Input) {
				input.WorkerSpec.Runtime.ModelBinding = workerSpecModelBinding(
					materializeModel(input.Dependencies.ToolModels[0].Model),
				)
			},
			match: "primary model does not match",
		},
		{
			name: "tool model",
			mutate: func(input *Input) {
				input.WorkerSpec.Runtime.ToolModelBindings[0].
					ModelBinding.ModelID = "substituted-model"
			},
			match: "tool model",
		},
		{
			name: "repository",
			mutate: func(input *Input) {
				id := int64(61)
				input.WorkerSpec.Workspace.RepositoryID = &id
				input.WorkerSpec.Workspace.Branch = "main"
			},
			match: "repository",
		},
		{
			name: "Skill",
			mutate: func(input *Input) {
				input.WorkerSpec.Workspace.SkillIDs = []int64{71}
			},
			match: "Skills",
		},
		{
			name: "KnowledgeBase",
			mutate: func(input *Input) {
				input.WorkerSpec.Workspace.KnowledgeMounts =
					[]workerspec.KnowledgeMount{{
						KnowledgeBaseID: 81,
						Mode:            workerspec.KnowledgeMountReadOnly,
					}}
			},
			match: "KnowledgeBases",
		},
		{
			name: "runtime bundle",
			mutate: func(input *Input) {
				input.WorkerSpec.Workspace.EnvBundleIDs =
					[]workerspec.RuntimeEnvBundleID{91}
			},
			match: "runtime bundles",
		},
		{
			name: "config document",
			mutate: func(input *Input) {
				input.WorkerSpec.Workspace.ConfigDocumentBindings =
					[]workerspec.ConfigDocumentBinding{{
						DocumentID: "settings", ConfigBundleID: 92,
					}}
			},
			match: "config documents",
		},
		{
			name: "Secret reference",
			mutate: func(input *Input) {
				input.WorkerSpec.TypeConfig.SecretRefs["CURSOR_API_KEY"] =
					workerspec.SecretReference{
						Kind: slugkit.MustNewForTest("env-bundle"),
						ID:   93,
					}
			},
			match: "Secret references",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := validInput(t)
			test.mutate(&input)

			artifact, err := Build(input)

			require.ErrorContains(t, err, test.match)
			assert.Empty(t, artifact.JSON())
		})
	}
}

func TestBuildRejectsInvalidPlanWorkerSpecBeforeReferenceChecks(t *testing.T) {
	input := validInput(t)
	input.WorkerSpec.Version = 0
	input.PlanReferences = []control.ResolvedReference{
		resolvedReference(resource.KindSkill, "irrelevant"),
	}

	_, err := Build(input)

	require.ErrorContains(t, err, "validate Plan WorkerSpec")
}
