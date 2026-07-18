package agentpod

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreatePodSetupHeredocModeTextKeepsAutonomousCodexACP(t *testing.T) {
	coord := &mockPodCoordinator{}
	orch, podSvc, _ := setupOrchestrator(t,
		withCoordinator(coord),
		withAgentConfigProvider(newCodexTestProvider()),
	)

	result, err := createPodWithPlanSourceForTest(t, orch, context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID:  1,
		UserID:          1,
		RunnerID:        1,
		AgentSlug:       "codex-cli",
		ModelResourceID: testModelResourceID(),
		AgentfileLayer:  ptrStr("SETUP <<SCRIPT\nMODE pty\nSCRIPT"),
	})

	require.NoError(t, err)
	dbPod, err := podSvc.GetPod(context.Background(), result.Pod.PodKey)
	require.NoError(t, err)
	assert.Equal(t, "acp", dbPod.InteractionMode)
	assert.Equal(t, "acp", coord.lastCmd.InteractionMode)
	assert.Equal(t, []string{"app-server"}, coord.lastCmd.LaunchArgs)
}
