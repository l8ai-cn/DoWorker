package agentpod

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreatePodSetupScriptModeTextDoesNotDisableAutonomousCodexACP(t *testing.T) {
	coord := &mockPodCoordinator{}
	orch, podSvc, _ := setupOrchestrator(t,
		withCoordinator(coord),
		withAgentConfigProvider(newCodexTestProvider()),
	)

	result, err := orch.CreatePod(context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID: 1,
		UserID:         1,
		RunnerID:       1,
		AgentSlug:      "codex-cli",
		AgentfileLayer: ptrStr("SETUP <<SCRIPT\nMODE pty\nSCRIPT"),
	})

	require.NoError(t, err)
	dbPod, err := podSvc.GetPod(context.Background(), result.Pod.PodKey)
	require.NoError(t, err)
	assert.Equal(t, "acp", dbPod.InteractionMode)
	assert.Equal(t, "acp", coord.lastCmd.InteractionMode)
	assert.Equal(t, []string{"app-server"}, coord.lastCmd.LaunchArgs)
}
