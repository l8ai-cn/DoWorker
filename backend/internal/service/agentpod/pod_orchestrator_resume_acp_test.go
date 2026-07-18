package agentpod

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
)

func TestCreatePod_ResumeMode_PreservesACPInteractionMode(t *testing.T) {
	coord := &mockPodCoordinator{}
	orch, _, db := setupOrchestrator(t,
		withCoordinator(coord),
		withAgentConfigProvider(newCodexTestProvider()),
	)

	source, err := createPodWithPlanSourceForTest(t, orch, context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID:  1,
		UserID:          1,
		RunnerID:        1,
		AgentSlug:       "codex-cli",
		ModelResourceID: testModelResourceID(),
	})
	require.NoError(t, err)
	require.Equal(t, podDomain.InteractionModeACP, source.Pod.InteractionMode)

	externalID := "codex-thread-1"
	require.NoError(t, db.Model(&podDomain.Pod{}).
		Where("pod_key = ?", source.Pod.PodKey).
		Updates(map[string]interface{}{
			"external_session_id": externalID,
			"status":              podDomain.StatusCompleted,
		}).Error)

	resumed, err := createPodWithPlanSourceForTest(t, orch, context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID: 1,
		UserID:         1,
		SourcePodKey:   source.Pod.PodKey,
	})
	require.NoError(t, err)

	assert.Equal(t, podDomain.InteractionModeACP, resumed.Pod.InteractionMode)
	assert.Equal(t, podDomain.InteractionModeACP, coord.lastCmd.InteractionMode)
	assert.Equal(t, []string{"app-server"}, coord.lastCmd.LaunchArgs)
	assert.Equal(t, externalID, coord.lastCmd.EnvVars["AGENTSMESH_RESUME_EXTERNAL_SESSION"])
}

func TestCreatePod_ResumeMode_InheritsAncestorExternalSessionID(t *testing.T) {
	coord := &mockPodCoordinator{}
	orch, podSvc, db := setupOrchestrator(t,
		withCoordinator(coord),
		withAgentConfigProvider(newCodexTestProvider()),
	)

	root, err := podSvc.CreatePod(context.Background(), &CreatePodRequest{
		OrganizationID:  1,
		RunnerID:        1,
		AgentSlug:       "codex-cli",
		ModelResourceID: testModelResourceID(),
		CreatedByID:     1,
		SessionID:       "platform-session",
		InteractionMode: podDomain.InteractionModeACP,
	})
	require.NoError(t, err)
	externalID := "codex-thread-root"
	require.NoError(t, db.Model(&podDomain.Pod{}).
		Where("pod_key = ?", root.PodKey).
		Updates(map[string]interface{}{
			"external_session_id": externalID,
			"status":              podDomain.StatusCompleted,
		}).Error)

	intermediate, err := podSvc.CreatePod(context.Background(), &CreatePodRequest{
		OrganizationID:  1,
		RunnerID:        1,
		AgentSlug:       "codex-cli",
		ModelResourceID: testModelResourceID(),
		CreatedByID:     1,
		SessionID:       "platform-session",
		SourcePodKey:    root.PodKey,
		InteractionMode: podDomain.InteractionModeACP,
	})
	require.NoError(t, err)
	require.NoError(t, db.Model(&podDomain.Pod{}).
		Where("pod_key = ?", intermediate.PodKey).
		Update("status", podDomain.StatusOrphaned).Error)

	_, err = createPodWithPlanSourceForTest(t, orch, context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID: 1,
		UserID:         1,
		SourcePodKey:   intermediate.PodKey,
	})
	require.NoError(t, err)

	assert.Equal(t, externalID, coord.lastCmd.EnvVars["AGENTSMESH_RESUME_EXTERNAL_SESSION"])
}
