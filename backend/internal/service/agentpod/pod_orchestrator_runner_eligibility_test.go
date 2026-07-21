package agentpod

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	podDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agentpod"
	runnerDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/runner"
)

func TestCreatePod_ExplicitRunner_UsesEligibilityResolver(t *testing.T) {
	coord := &mockPodCoordinator{}
	selector := &mockRunnerSelector{
		err:           errors.New("affinity selection must not be called"),
		resolveRunner: &runnerDomain.Runner{ID: 5},
	}
	orch, _, _ := setupOrchestrator(t,
		withCoordinator(coord),
		withRunnerSelector(selector),
	)

	result, err := createPodWithPlanSourceForTest(t, orch, context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID:     17,
		UserID:             23,
		RunnerID:           5,
		AgentSlug:          "claude-code",
		ModelResourceID:    testModelResourceID(),
		AgentfileLayer:     ptrStr("CONFIG mcp_enabled = true"),
		QueueIfUnavailable: true,
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, int64(5), result.Pod.RunnerID)
	assert.Equal(t, &runnerResolveCall{
		RunnerID:         5,
		OrganizationID:   17,
		UserID:           23,
		AgentSlug:        "claude-code",
		AllowUnavailable: true,
	}, selector.resolveCall)
	assert.False(t, selector.selectCalled)
	assert.True(t, coord.createPodCalled)
}

func TestCreatePod_ExplicitRunner_ResolverRejectsBeforePersistenceOrDispatch(t *testing.T) {
	coord := &mockPodCoordinator{}
	selector := &mockRunnerSelector{resolveErr: errors.New("runner not eligible")}
	orch, _, db := setupOrchestrator(t,
		withCoordinator(coord),
		withRunnerSelector(selector),
	)

	_, err := createPodWithPlanSourceForTest(t, orch, context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID:  17,
		UserID:          23,
		RunnerID:        5,
		AgentSlug:       "claude-code",
		ModelResourceID: testModelResourceID(),
	})

	require.ErrorIs(t, err, ErrNoAvailableRunner)
	assert.Equal(t, &runnerResolveCall{
		RunnerID:       5,
		OrganizationID: 17,
		UserID:         23,
		AgentSlug:      "claude-code",
	}, selector.resolveCall)
	var podCount int64
	require.NoError(t, db.Model(&podDomain.Pod{}).Count(&podCount).Error)
	assert.Zero(t, podCount)
	assert.False(t, coord.createPodCalled)
}

func TestCreatePod_ExplicitRunner_RejectsBeforeAgentOrRepositoryResolution(t *testing.T) {
	coord := &mockPodCoordinator{}
	selector := &mockRunnerSelector{resolveErr: errors.New("runner not eligible")}
	agentResolver := &mockAgentResolver{err: errors.New("agent resolution must not be called")}
	repoSvc := &mockRepoService{}
	orch, _, db := setupOrchestrator(t,
		withCoordinator(coord),
		withRunnerSelector(selector),
		withAgentResolver(agentResolver),
		withRepoSvc(repoSvc),
	)

	result, err := createPodWithPlanSourceForTest(t, orch, context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID:  17,
		UserID:          23,
		RunnerID:        5,
		AgentSlug:       "claude-code",
		ModelResourceID: testModelResourceID(),
		RepositoryID:    intPtr(61),
	})

	assert.ErrorIs(t, err, ErrNoAvailableRunner)
	assert.Nil(t, result)
	assert.Zero(t, agentResolver.calls)
	assert.Empty(t, repoSvc.getAccessibleCalls)
	assert.Empty(t, repoSvc.findAccessibleCalls)
	assertNoPodOrDispatch(t, db, coord)
}

func TestCreatePod_ExplicitRunner_MissingResolverRejectsBeforePersistenceOrDispatch(t *testing.T) {
	coord := &mockPodCoordinator{}
	orch, _, db := setupOrchestrator(t,
		withCoordinator(coord),
		withRunnerSelector(nil),
	)

	_, err := createPodWithPlanSourceForTest(t, orch, context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID:  17,
		UserID:          23,
		RunnerID:        5,
		AgentSlug:       "claude-code",
		ModelResourceID: testModelResourceID(),
	})

	require.ErrorIs(t, err, ErrNoAvailableRunner)
	var podCount int64
	require.NoError(t, db.Model(&podDomain.Pod{}).Count(&podCount).Error)
	assert.Zero(t, podCount)
	assert.False(t, coord.createPodCalled)
}
