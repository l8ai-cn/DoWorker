package agentpod

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anthropics/agentsmesh/backend/internal/domain/gitprovider"
	runnerDomain "github.com/anthropics/agentsmesh/backend/internal/domain/runner"
)

type routingRepositoryService struct {
	byID      map[int64]*gitprovider.Repository
	getCalls  []int64
	findCalls []string
}

func (s *routingRepositoryService) GetAccessibleByID(_ context.Context, id, _, _ int64) (*gitprovider.Repository, error) {
	s.getCalls = append(s.getCalls, id)
	return s.byID[id], nil
}

func (s *routingRepositoryService) FindAccessibleByOrgSlug(_ context.Context, _, _ int64, slug string) (*gitprovider.Repository, error) {
	s.findCalls = append(s.findCalls, slug)
	return nil, ErrCreateResourceUnavailable
}

func TestCreatePod_RepositoryOverride_EmptyLayerSuppressesBase(t *testing.T) {
	repoSvc := &routingRepositoryService{}
	selector := &mockRunnerSelector{runner: &runnerDomain.Runner{ID: 23}}
	coord := &mockPodCoordinator{}
	orch, podSvc, _ := setupOrchestrator(t,
		withRepoSvc(repoSvc),
		withAgentResolver(&mockAgentResolver{err: errors.New("current agent changed")}),
		withRunnerSelector(selector),
		withCoordinator(coord),
	)
	req := repositoryOverrideRequest(ptrStr(`REPO ""`), nil)

	result, err := createPodWithPlanSourceForTest(t, orch, context.Background(), req)

	require.NoError(t, err)
	assert.Empty(t, repoSvc.findCalls)
	assert.Empty(t, repoSvc.getCalls)
	require.NotNil(t, selector.selectHints)
	assert.Nil(t, selector.selectHints.RepositoryID)
	persisted, err := podSvc.GetPod(context.Background(), result.Pod.PodKey)
	require.NoError(t, err)
	assert.Nil(t, persisted.RepositoryID)
	require.NotNil(t, coord.lastCmd)
	if coord.lastCmd.SandboxConfig != nil {
		assert.Empty(t, coord.lastCmd.SandboxConfig.HttpCloneUrl)
		assert.Empty(t, coord.lastCmd.SandboxConfig.SshCloneUrl)
	}
}

func TestCreatePod_RepositoryOverride_EmptyLayerFallsBackToDirectID(t *testing.T) {
	directID := int64(62)
	prepScript := "make prepare"
	prepTimeout := 300
	directRepo := &gitprovider.Repository{
		ID:                 directID,
		HttpCloneURL:       "https://example.com/direct.git",
		DefaultBranch:      "direct-main",
		PreparationScript:  &prepScript,
		PreparationTimeout: &prepTimeout,
	}
	repoSvc := &routingRepositoryService{
		byID: map[int64]*gitprovider.Repository{directID: directRepo},
	}
	selector := &mockRunnerSelector{runner: &runnerDomain.Runner{ID: 23}}
	coord := &mockPodCoordinator{}
	orch, podSvc, _ := setupOrchestrator(t,
		withRepoSvc(repoSvc),
		withAgentResolver(&mockAgentResolver{err: errors.New("current agent changed")}),
		withRunnerSelector(selector),
		withCoordinator(coord),
	)
	req := repositoryOverrideRequest(ptrStr(`REPO ""`), &directID)

	result, err := createPodWithPlanSourceForTest(t, orch, context.Background(), req)

	require.NoError(t, err)
	assert.Empty(t, repoSvc.findCalls)
	assert.Equal(t, []int64{directID}, repoSvc.getCalls)
	require.NotNil(t, selector.selectHints)
	require.NotNil(t, selector.selectHints.RepositoryID)
	assert.Equal(t, directID, *selector.selectHints.RepositoryID)
	persisted, err := podSvc.GetPod(context.Background(), result.Pod.PodKey)
	require.NoError(t, err)
	require.NotNil(t, persisted.RepositoryID)
	assert.Equal(t, directID, *persisted.RepositoryID)
	require.NotNil(t, coord.lastCmd.SandboxConfig)
	assert.Equal(t, "https://example.com/direct.git", coord.lastCmd.SandboxConfig.HttpCloneUrl)
	assert.Equal(t, "direct-main", coord.lastCmd.SandboxConfig.SourceBranch)
	assert.Equal(t, "make prepare", coord.lastCmd.SandboxConfig.PreparationScript)
}

func TestCreatePod_RepositoryOverride_UsesPlanPinnedRepository(t *testing.T) {
	repoA := &gitprovider.Repository{ID: 71, HttpCloneURL: "https://example.com/a.git", DefaultBranch: "a-main"}
	repoB := &gitprovider.Repository{ID: 72, HttpCloneURL: "https://example.com/b.git", DefaultBranch: "b-main"}
	repoSvc := &routingRepositoryService{byID: map[int64]*gitprovider.Repository{
		repoA.ID: repoA,
		repoB.ID: repoB,
	}}
	selector := &mockRunnerSelector{runner: &runnerDomain.Runner{ID: 23}}
	coord := &mockPodCoordinator{}
	orch, podSvc, _ := setupOrchestrator(t,
		withRepoSvc(repoSvc),
		withAgentResolver(&mockAgentResolver{err: errors.New("current agent changed")}),
		withRunnerSelector(selector),
		withCoordinator(coord),
	)
	repositoryID := repoA.ID

	result, err := createPodWithPlanSourceForTest(
		t,
		orch,
		context.Background(),
		repositoryOverrideRequest(nil, &repositoryID),
	)

	require.NoError(t, err)
	assert.Equal(t, []int64{repoA.ID}, repoSvc.getCalls)
	assert.Empty(t, repoSvc.findCalls)
	require.NotNil(t, selector.selectHints.RepositoryID)
	assert.Equal(t, int64(71), *selector.selectHints.RepositoryID)
	persisted, err := podSvc.GetPod(context.Background(), result.Pod.PodKey)
	require.NoError(t, err)
	require.NotNil(t, persisted.RepositoryID)
	assert.Equal(t, int64(71), *persisted.RepositoryID)
	require.NotNil(t, coord.lastCmd.SandboxConfig)
	assert.Equal(t, "https://example.com/a.git", coord.lastCmd.SandboxConfig.HttpCloneUrl)
	assert.Equal(t, "a-main", coord.lastCmd.SandboxConfig.SourceBranch)
}

func TestCreatePod_RepositoryOverride_RejectsMalformedLayerBeforePlacement(t *testing.T) {
	repoSvc := &routingRepositoryService{}
	selector := &mockRunnerSelector{runner: &runnerDomain.Runner{ID: 23}}
	coord := &mockPodCoordinator{}
	orch, _, db := setupOrchestrator(t,
		withRepoSvc(repoSvc),
		withAgentResolver(&mockAgentResolver{err: errors.New("current agent changed")}),
		withRunnerSelector(selector),
		withCoordinator(coord),
	)
	req := repositoryOverrideRequest(ptrStr(`INVALID @@@ not valid syntax`), nil)

	result, err := createPodWithPlanSourceForTest(t, orch, context.Background(), req)

	require.ErrorIs(t, err, ErrInvalidAgentfileLayer)
	assert.Nil(t, result)
	assert.False(t, selector.selectCalled)
	assert.Empty(t, repoSvc.findCalls)
	assert.Empty(t, repoSvc.getCalls)
	assertNoPodOrDispatch(t, db, coord)
}

func repositoryOverrideRequest(layer *string, repositoryID *int64) *OrchestrateCreatePodRequest {
	return &OrchestrateCreatePodRequest{
		OrganizationID:  7,
		UserID:          11,
		RunnerID:        0,
		AgentSlug:       "claude-code",
		ModelResourceID: testModelResourceID(),
		RepositoryID:    repositoryID,
		AgentfileLayer:  layer,
	}
}
