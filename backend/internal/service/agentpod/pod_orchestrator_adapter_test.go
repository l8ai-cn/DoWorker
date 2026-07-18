package agentpod

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreatePodSendsExplicitAgentAdapterID(t *testing.T) {
	coordinator := &mockPodCoordinator{}
	orchestrator, _, _ := setupOrchestrator(t, withCoordinator(coordinator))

	_, err := createPodWithPlanSourceForTest(t, orchestrator, context.Background(), adapterTestCreateRequest())

	require.NoError(t, err)
	require.NotNil(t, coordinator.lastCmd)
	assert.Equal(t, "claude-stream-json", coordinator.lastCmd.AdapterId)
}

func TestCreatePodRejectsMissingAgentAdapterID(t *testing.T) {
	coordinator := &mockPodCoordinator{}
	provider := newTestProvider()
	provider.agentDef.AdapterID = ""
	orchestrator, _, _ := setupOrchestrator(t,
		withCoordinator(coordinator),
		withAgentConfigProvider(provider),
	)

	_, err := createPodWithPlanSourceForTest(t, orchestrator, context.Background(), adapterTestCreateRequest())

	require.ErrorIs(t, err, ErrMissingAgentAdapter)
	assert.False(t, coordinator.createPodCalled)
}

func adapterTestCreateRequest() *OrchestrateCreatePodRequest {
	return &OrchestrateCreatePodRequest{
		OrganizationID:  1,
		UserID:          1,
		RunnerID:        1,
		AgentSlug:       "claude-code",
		ModelResourceID: testModelResourceID(),
	}
}
