package agentpod

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	"github.com/anthropics/agentsmesh/backend/internal/domain/gitprovider"
)

func TestCreatePod_ResumeMode_AgentSlugMismatch_Rejected(t *testing.T) {
	coord := &mockPodCoordinator{}
	orch, podSvc, db := setupOrchestrator(t, withCoordinator(coord))

	sourcePod, err := podSvc.CreatePod(context.Background(), &CreatePodRequest{
		OrganizationID:  1,
		RunnerID:        1,
		AgentSlug:       "claude-code",
		ModelResourceID: testModelResourceID(),
		CreatedByID:     1,
		SessionID:       "session-1",
	})
	require.NoError(t, err)
	db.Exec("UPDATE pods SET status = ? WHERE pod_key = ?", podDomain.StatusTerminated, sourcePod.PodKey)

	result, err := createPodWithPlanSourceForTest(t, orch, context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID: 1,
		UserID:         1,
		AgentSlug:      "codex-cli", // Different agent than source pod
		SourcePodKey:   sourcePod.PodKey,
	})

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrResumeAgentMismatch))
	assert.Nil(t, result)
	assert.False(t, coord.createPodCalled, "no runner dispatch should happen on cross-agent resume")
}

func TestCreatePod_ResumeMode_AgentSlugMatch_Accepted(t *testing.T) {
	coord := &mockPodCoordinator{}
	orch, podSvc, db := setupOrchestrator(t, withCoordinator(coord))

	sourcePod := createImmutableResumeSource(t, orch, podSvc, db, &OrchestrateCreatePodRequest{
		OrganizationID: 1,
		UserID:         1,
		AgentSlug:      "claude-code",
	})

	// Explicit AgentSlug matching source — should be accepted (not rejected).
	result, err := createPodWithPlanSourceForTest(t, orch, context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID:  1,
		UserID:          1,
		AgentSlug:       "claude-code",
		ModelResourceID: testModelResourceID(),
		SourcePodKey:    sourcePod.PodKey,
	})

	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestCreatePod_ResumeMode_Success(t *testing.T) {
	coord := &mockPodCoordinator{}
	orch, podSvc, db := setupOrchestrator(t, withCoordinator(coord))

	agentSlug := "claude-code"
	sourcePod := createImmutableResumeSource(t, orch, podSvc, db, &OrchestrateCreatePodRequest{
		OrganizationID: 1,
		UserID:         1,
		AgentSlug:      agentSlug,
	})

	result, err := createPodWithPlanSourceForTest(t, orch, context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID: 1,
		UserID:         1,
		SourcePodKey:   sourcePod.PodKey,
	})

	require.NoError(t, err)
	assert.NotNil(t, result.Pod)
	// Should inherit runner_id and agent_slug from source pod
	assert.Equal(t, int64(1), result.Pod.RunnerID)
	assert.Equal(t, agentSlug, result.Pod.AgentSlug)
	require.NotNil(t, result.Pod.ModelResourceID)
	assert.Equal(t, *sourcePod.ModelResourceID, *result.Pod.ModelResourceID)
}

func TestCreatePod_ResumeMode_SourcePodNotFound(t *testing.T) {
	orch, _, _ := setupOrchestrator(t)

	_, err := createPodWithPlanSourceForTest(t, orch, context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID: 1,
		UserID:         1,
		SourcePodKey:   "non-existent-pod",
	})

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrSourcePodNotFound))
}

func TestCreatePod_ResumeMode_AccessDenied(t *testing.T) {
	orch, podSvc, db := setupOrchestrator(t)

	agentSlug := "claude-code"
	sourcePod, err := podSvc.CreatePod(context.Background(), &CreatePodRequest{
		OrganizationID:  999, // Different org
		RunnerID:        2,
		AgentSlug:       agentSlug,
		ModelResourceID: testModelResourceID(),
		CreatedByID:     1,
	})
	require.NoError(t, err)
	db.Exec("UPDATE pods SET status = ? WHERE pod_key = ?", podDomain.StatusTerminated, sourcePod.PodKey)

	_, err = createPodWithPlanSourceForTest(t, orch, context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID: 1, // Different org from source pod
		UserID:         1,
		SourcePodKey:   sourcePod.PodKey,
	})

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrSourcePodAccessDenied))
}

func TestCreatePod_ResumeMode_NotTerminated(t *testing.T) {
	orch, podSvc, _ := setupOrchestrator(t)

	agentSlug := "claude-code"
	sourcePod, err := podSvc.CreatePod(context.Background(), &CreatePodRequest{
		OrganizationID:  1,
		RunnerID:        1,
		ClusterID:       19,
		AgentSlug:       agentSlug,
		ModelResourceID: testModelResourceID(),
		CreatedByID:     1,
	})
	require.NoError(t, err)
	// Pod is still "initializing" (default status)

	_, err = createPodWithPlanSourceForTest(t, orch, context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID: 1,
		UserID:         1,
		SourcePodKey:   sourcePod.PodKey,
	})

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrSourcePodNotTerminated))
}

func TestCreatePod_ResumeMode_AlreadyResumed(t *testing.T) {
	coord := &mockPodCoordinator{}
	orch, podSvc, db := setupOrchestrator(t, withCoordinator(coord))

	sourcePod := createImmutableResumeSource(t, orch, podSvc, db, &OrchestrateCreatePodRequest{
		OrganizationID: 1,
		UserID:         1,
		AgentSlug:      "claude-code",
	})

	// First resume should succeed
	_, err := createPodWithPlanSourceForTest(t, orch, context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID: 1,
		UserID:         1,
		SourcePodKey:   sourcePod.PodKey,
	})
	require.NoError(t, err)

	// Second resume from same source should fail
	_, err = createPodWithPlanSourceForTest(t, orch, context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID: 1,
		UserID:         1,
		SourcePodKey:   sourcePod.PodKey,
	})

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrSourcePodAlreadyResumed))
}

func TestCreatePod_ResumeMode_RunnerMismatch(t *testing.T) {
	orch, podSvc, db := setupOrchestrator(t)

	// Insert a second runner
	db.Exec("INSERT INTO runners (id, node_id, status, current_pods) VALUES (2, 'runner-002', 'online', 0)")

	agentSlug := "claude-code"
	sourcePod, err := podSvc.CreatePod(context.Background(), &CreatePodRequest{
		OrganizationID:  1,
		RunnerID:        1, // Source on runner 1
		AgentSlug:       agentSlug,
		ModelResourceID: testModelResourceID(),
		CreatedByID:     1,
		SessionID:       "session-1",
	})
	require.NoError(t, err)
	db.Exec("UPDATE pods SET status = ? WHERE pod_key = ?", podDomain.StatusTerminated, sourcePod.PodKey)

	_, err = createPodWithPlanSourceForTest(t, orch, context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID: 1,
		UserID:         1,
		RunnerID:       2, // Different runner
		SourcePodKey:   sourcePod.PodKey,
	})

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrResumeRunnerMismatch))
}

func TestCreatePod_ResumeMode_InheritRunnerID(t *testing.T) {
	coord := &mockPodCoordinator{}
	orch, podSvc, db := setupOrchestrator(t, withCoordinator(coord))

	sourcePod := createImmutableResumeSource(t, orch, podSvc, db, &OrchestrateCreatePodRequest{
		OrganizationID: 1,
		UserID:         1,
		AgentSlug:      "claude-code",
	})

	// RunnerID=0 -> should inherit from source pod
	result, err := createPodWithPlanSourceForTest(t, orch, context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID: 1,
		UserID:         1,
		RunnerID:       0,
		SourcePodKey:   sourcePod.PodKey,
	})

	require.NoError(t, err)
	assert.Equal(t, sourcePod.RunnerID, result.Pod.RunnerID)
	assert.Equal(t, sourcePod.ClusterID, result.Pod.ClusterID)
}

func TestCreatePod_ResumeMode_InheritConfig(t *testing.T) {
	coord := &mockPodCoordinator{}
	repoID := int64(10)
	repoSvc := &mockRepoService{repo: &gitprovider.Repository{
		ID:            repoID,
		HttpCloneURL:  "https://git.example.com/org/resume-config.git",
		DefaultBranch: "main",
	}}
	orch, podSvc, db := setupOrchestrator(t, withCoordinator(coord), withRepoSvc(repoSvc))

	agentSlug := "claude-code"
	ticketID := int64(20)
	branch := "feature-branch"
	sourcePod := createImmutableResumeSource(t, orch, podSvc, db, &OrchestrateCreatePodRequest{
		OrganizationID: 1,
		UserID:         1,
		AgentSlug:      agentSlug,
		RepositoryID:   &repoID,
		TicketID:       &ticketID,
		BranchName:     &branch,
	})

	result, err := createPodWithPlanSourceForTest(t, orch, context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID: 1,
		UserID:         1,
		SourcePodKey:   sourcePod.PodKey,
	})

	require.NoError(t, err)
	assert.Equal(t, agentSlug, result.Pod.AgentSlug)
	assert.Equal(t, &repoID, result.Pod.RepositoryID)
	assert.Equal(t, &ticketID, result.Pod.TicketID)
	assert.Equal(t, &branch, result.Pod.BranchName)
}

func TestCreatePod_ResumeMode_RejectsLegacySourceWithoutSnapshot(t *testing.T) {
	orch, podSvc, db := setupOrchestrator(t)
	sourcePod, err := podSvc.CreatePod(context.Background(), &CreatePodRequest{
		OrganizationID:  1,
		RunnerID:        1,
		AgentSlug:       "claude-code",
		ModelResourceID: testModelResourceID(),
		CreatedByID:     1,
	})
	require.NoError(t, err)
	require.NoError(
		t,
		db.Model(&podDomain.Pod{}).
			Where("pod_key = ?", sourcePod.PodKey).
			Update("status", podDomain.StatusTerminated).Error,
	)

	_, err = createPodWithPlanSourceForTest(t, orch, context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID: 1,
		UserID:         1,
		SourcePodKey:   sourcePod.PodKey,
	})

	require.ErrorIs(t, err, ErrWorkerSpecSnapshotUnavailable)
}

func TestCreatePod_ResumeMode_RejectsModelOverride(t *testing.T) {
	orch, podSvc, db := setupOrchestrator(t)
	sourcePod := createImmutableResumeSource(t, orch, podSvc, db, &OrchestrateCreatePodRequest{
		OrganizationID: 1,
		UserID:         1,
		AgentSlug:      "claude-code",
	})
	override := *sourcePod.ModelResourceID + 1

	_, err := createPodWithPlanSourceForTest(t, orch, context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID:  1,
		UserID:          1,
		SourcePodKey:    sourcePod.PodKey,
		ModelResourceID: &override,
	})

	require.ErrorIs(t, err, ErrWorkerSpecSnapshotMismatch)
}
