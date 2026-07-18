package workerdependencyartifact

import (
	"testing"

	"github.com/anthropics/agentsmesh/backend/internal/domain/envbundle"
	"github.com/anthropics/agentsmesh/backend/internal/domain/workerdependency"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildDerivesWorkerSnapshotFromVerifiedDefinition(t *testing.T) {
	input := validInput(t)
	input.AgentfileLayer = `CONFIG approval_mode = "never"`

	artifact, err := Build(input)
	require.NoError(t, err)
	document, err := workerdependency.Decode(artifact.JSON())
	require.NoError(t, err)

	assert.Equal(t, input.Definition.AdapterID, document.Worker.AdapterID.String())
	assert.Equal(t, input.Definition.DefinitionHash, document.Worker.DefinitionHash)
	assert.Contains(t, document.Worker.AgentfileSource, `CONFIG approval_mode = "never"`)
}

func TestBuildRejectsEnvironmentDeclarationInAgentfileLayer(t *testing.T) {
	input := validInput(t)
	input.AgentfileLayer = `ENV UNDECLARED_TOKEN = "opaque-value"`

	_, err := Build(input)

	require.ErrorContains(t, err, "layer must not declare ENV fields")
}

func TestBuildRejectsWorkerDefinitionIntegritySubstitution(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Input)
		match  string
	}{
		{
			name: "adapter projection",
			mutate: func(input *Input) {
				input.Definition.AdapterID = "other-adapter"
			},
			match: "projection does not match",
		},
		{
			name: "AgentFile source",
			mutate: func(input *Input) {
				input.Definition.AgentFile = "AGENT other\nMODE pty\n"
			},
			match: "snapshot is invalid",
		},
		{
			name: "definition source",
			mutate: func(input *Input) {
				input.Definition.DefinitionSource[0] = '['
			},
			match: "snapshot is invalid",
		},
		{
			name: "WorkerSpec definition hash",
			mutate: func(input *Input) {
				input.WorkerSpec.Runtime.WorkerType.DefinitionHash =
					workerdependency.TextDigest("other")[len("sha256:"):]
			},
			match: "does not match Plan WorkerSpec",
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

func TestBuildRejectsDefinitionWorkspaceSubstitution(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Input)
		match  string
	}{
		{
			name: "config metadata",
			mutate: func(input *Input) {
				input.Dependencies.RuntimeBundles[1].ConfigDocument.TargetPath =
					"other/settings.json"
			},
			match: "config document",
		},
		{
			name: "Secret bundle key",
			mutate: func(input *Input) {
				input.Dependencies.SecretReferences[0].BundleKey = "OTHER_KEY"
			},
			match: "does not match worker definition",
		},
		{
			name: "user Secret owner",
			mutate: func(input *Input) {
				input.Dependencies.SecretReferences[0].OwnerScope =
					envbundle.OwnerScopeUser
				input.Dependencies.SecretReferences[0].OwnerID = 99
			},
			match: "does not match Plan actor",
		},
		{
			name: "tool model provider",
			mutate: func(input *Input) {
				input.WorkerSpec.Runtime.ToolModelBindings[0].
					ModelBinding.ProviderKey = "other-provider"
				input.Dependencies.ToolModels[0].Model.ProviderKey = "other-provider"
			},
			match: "rejects tool model role",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := validInput(t)
			addWorkspaceDependencies(t, &input)
			test.mutate(&input)

			_, err := Build(input)

			require.ErrorContains(t, err, test.match)
		})
	}
}
