package agentpod

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	podDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agentpod"
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
	assert.Equal(t, externalID, coord.lastCmd.EnvVars["AGENTCLOUD_RESUME_EXTERNAL_SESSION"])
}

func TestCreatePod_ResumeMode_InheritsAncestorExternalSessionID(t *testing.T) {
	coord := &mockPodCoordinator{}
	orch, podSvc, db := setupOrchestrator(t,
		withCoordinator(coord),
		withAgentConfigProvider(newCodexTestProvider()),
	)

	root := createImmutableResumeSource(t, orch, podSvc, db, &OrchestrateCreatePodRequest{
		OrganizationID: 1,
		UserID:         1,
		AgentSlug:      "codex-cli",
	})
	externalID := "codex-thread-root"
	root = updateResumeSource(t, podSvc, db, root.PodKey, map[string]interface{}{
		"external_session_id": externalID,
		"session_id":          "platform-session",
		"status":              podDomain.StatusCompleted,
	})

	intermediateResult, err := createPodWithPlanSourceForTest(
		t,
		orch,
		context.Background(),
		&OrchestrateCreatePodRequest{
			OrganizationID: 1,
			UserID:         1,
			SourcePodKey:   root.PodKey,
		},
	)
	require.NoError(t, err)
	intermediate := updateResumeSource(
		t,
		podSvc,
		db,
		intermediateResult.Pod.PodKey,
		map[string]interface{}{"status": podDomain.StatusOrphaned},
	)

	_, err = createPodWithPlanSourceForTest(t, orch, context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID: 1,
		UserID:         1,
		SourcePodKey:   intermediate.PodKey,
	})
	require.NoError(t, err)

	assert.Equal(t, externalID, coord.lastCmd.EnvVars["AGENTCLOUD_RESUME_EXTERNAL_SESSION"])
}
