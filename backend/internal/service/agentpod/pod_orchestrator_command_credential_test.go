package agentpod

import (
	"context"
	"testing"

	"github.com/anthropics/agentsmesh/backend/internal/domain/workerdependency"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPodCommandGitCredentialUsesNoneWithoutArtifactRepository(t *testing.T) {
	orchestrator := NewPodOrchestrator(&PodOrchestratorDeps{})
	credential, err := orchestrator.podCommandGitCredential(
		context.Background(),
		&OrchestrateCreatePodRequest{
			preResolvedDependencies: &workerdependency.Document{},
		},
	)

	require.NoError(t, err)
	assert.Equal(t, workerdependency.RepositoryCredentialTypeNone, credential.credentialType)
}

func TestPodCommandGitCredentialUsesRunnerLocalWithoutLegacyCredential(t *testing.T) {
	orchestrator := NewPodOrchestrator(&PodOrchestratorDeps{})
	credential, err := orchestrator.podCommandGitCredential(
		context.Background(),
		&OrchestrateCreatePodRequest{},
	)

	require.NoError(t, err)
	assert.Equal(t, "runner_local", credential.credentialType)
}

func TestPodCommandRepositoryUsesArtifactCommit(t *testing.T) {
	config, err := podCommandRepository(
		&OrchestrateCreatePodRequest{
			preResolvedDependencies: &workerdependency.Document{
				Repository: &workerdependency.Repository{
					HTTPCloneURL: "https://example.test/acme/repository.git",
					Branch:       "main",
					CommitSHA:    "0123456789abcdef0123456789abcdef01234567",
				},
			},
		},
		&agentfileResolved{},
		nil,
	)

	require.NoError(t, err)
	assert.Equal(t, "main", config.sourceBranch)
	assert.Equal(t, "0123456789abcdef0123456789abcdef01234567", config.sourceCommitSha)
}
