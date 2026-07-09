package agentpod

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	"github.com/anthropics/agentsmesh/backend/internal/domain/gitprovider"
	"github.com/anthropics/agentsmesh/backend/internal/service/repository"
)

func TestCreatePod_RejectsUnavailableDirectRepository(t *testing.T) {
	repoID := int64(41)
	repoSvc := &mockRepoService{err: repository.ErrNoPermission}
	coord := &mockPodCoordinator{}
	orch, _, db := setupOrchestrator(t, withRepoSvc(repoSvc), withCoordinator(coord))

	result, err := orch.CreatePod(context.Background(), repositoryCreateRequest(&repoID, nil))

	require.ErrorIs(t, err, ErrCreateResourceUnavailable)
	require.Equal(t, ErrCreateResourceUnavailable, err)
	assert.Nil(t, result)
	assertNoPodOrDispatch(t, db, coord)
}

func TestCreatePod_RejectsUnavailableAgentFileRepository(t *testing.T) {
	repoSvc := &mockRepoService{err: repository.ErrNoPermission}
	coord := &mockPodCoordinator{}
	orch, _, db := setupOrchestrator(t, withRepoSvc(repoSvc), withCoordinator(coord))

	result, err := orch.CreatePod(context.Background(), repositoryCreateRequest(nil, ptrStr(`REPO "org/private"`)))

	require.ErrorIs(t, err, ErrCreateResourceUnavailable)
	require.Equal(t, ErrCreateResourceUnavailable, err)
	assert.Nil(t, result)
	assertNoPodOrDispatch(t, db, coord)
}

func TestCreatePod_RejectsAmbiguousAgentFileRepository(t *testing.T) {
	repoSvc := &mockRepoService{err: repository.ErrAmbiguousRepositorySlug}
	coord := &mockPodCoordinator{}
	orch, _, db := setupOrchestrator(t, withRepoSvc(repoSvc), withCoordinator(coord))

	result, err := orch.CreatePod(context.Background(), repositoryCreateRequest(nil, ptrStr(`REPO "org/shared"`)))

	require.ErrorIs(t, err, ErrCreateResourceUnavailable)
	require.Equal(t, ErrCreateResourceUnavailable, err)
	assert.Nil(t, result)
	assertNoPodOrDispatch(t, db, coord)
}

func TestCreatePod_RejectsRepositoryWithoutResolver(t *testing.T) {
	repoID := int64(42)
	coord := &mockPodCoordinator{}
	orch, _, db := setupOrchestrator(t, withCoordinator(coord))

	result, err := orch.CreatePod(context.Background(), repositoryCreateRequest(&repoID, nil))

	require.ErrorIs(t, err, ErrCreateResourceUnavailable)
	require.Equal(t, ErrCreateResourceUnavailable, err)
	assert.Nil(t, result)
	assertNoPodOrDispatch(t, db, coord)
}

func TestCreatePod_UsesScopedAgentFileRepositoryOnce(t *testing.T) {
	prepScript := "pnpm install"
	prepTimeout := 480
	repoSvc := &mockRepoService{repo: &gitprovider.Repository{
		ID:                 43,
		HttpCloneURL:       "https://example.com/org/project.git",
		SshCloneURL:        "git@example.com:org/project.git",
		DefaultBranch:      "develop",
		PreparationScript:  &prepScript,
		PreparationTimeout: &prepTimeout,
	}}
	coord := &mockPodCoordinator{}
	orch, _, _ := setupOrchestrator(t, withRepoSvc(repoSvc), withCoordinator(coord))

	result, err := orch.CreatePod(context.Background(), repositoryCreateRequest(nil, ptrStr(`REPO "org/project"`)))

	require.NoError(t, err)
	require.NotNil(t, result.Pod.RepositoryID)
	assert.Equal(t, int64(43), *result.Pod.RepositoryID)
	assert.Empty(t, repoSvc.getAccessibleCalls)
	require.Equal(t, []repositorySlugAccessCall{{OrganizationID: 7, UserID: 11, Slug: "org/project"}}, repoSvc.findAccessibleCalls)
	require.NotNil(t, coord.lastCmd)
	require.NotNil(t, coord.lastCmd.SandboxConfig)
	assert.Equal(t, "https://example.com/org/project.git", coord.lastCmd.SandboxConfig.HttpCloneUrl)
	assert.Equal(t, "git@example.com:org/project.git", coord.lastCmd.SandboxConfig.SshCloneUrl)
	assert.Equal(t, "develop", coord.lastCmd.SandboxConfig.SourceBranch)
	assert.Equal(t, "pnpm install", coord.lastCmd.SandboxConfig.PreparationScript)
	assert.Equal(t, int32(480), coord.lastCmd.SandboxConfig.PreparationTimeout)
}

func TestCreatePod_RejectsRepositoryLookupInfrastructureErrorsUnchanged(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
	}{
		{name: "database", err: errors.New("repository database unavailable")},
		{name: "context canceled", err: context.Canceled},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repoID := int64(44)
			repoSvc := &mockRepoService{err: tc.err}
			coord := &mockPodCoordinator{}
			orch, _, db := setupOrchestrator(t, withRepoSvc(repoSvc), withCoordinator(coord))

			result, err := orch.CreatePod(context.Background(), repositoryCreateRequest(&repoID, nil))

			require.ErrorIs(t, err, tc.err)
			require.Equal(t, tc.err, err)
			assert.NotErrorIs(t, err, ErrCreateResourceUnavailable)
			assert.Nil(t, result)
			assertNoPodOrDispatch(t, db, coord)
		})
	}
}

func TestCreatePod_UsesNoRepositoryWithoutResolver(t *testing.T) {
	coord := &mockPodCoordinator{}
	orch, _, _ := setupOrchestrator(t, withCoordinator(coord))

	result, err := orch.CreatePod(context.Background(), repositoryCreateRequest(nil, nil))

	require.NoError(t, err)
	require.NotNil(t, result.Pod)
	assert.Nil(t, result.Pod.RepositoryID)
	assert.True(t, coord.createPodCalled)
}

func repositoryCreateRequest(repoID *int64, layer *string) *OrchestrateCreatePodRequest {
	return &OrchestrateCreatePodRequest{
		OrganizationID: 7,
		UserID:         11,
		RunnerID:       1,
		AgentSlug:      "claude-code",
		RepositoryID:   repoID,
		AgentfileLayer: layer,
	}
}

func assertNoPodOrDispatch(t *testing.T, db *gorm.DB, coord *mockPodCoordinator) {
	t.Helper()
	var count int64
	require.NoError(t, db.Model(&podDomain.Pod{}).Count(&count).Error)
	assert.Zero(t, count)
	assert.False(t, coord.createPodCalled)
}
