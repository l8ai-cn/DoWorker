package agentpod

import (
	"context"
	"testing"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/workerdependency"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
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

func TestArtifactWorkerAgentRejectsMissingAdapterID(t *testing.T) {
	_, err := artifactWorkerAgent(&workerdependency.Document{
		Worker: workerdependency.Worker{
			WorkerType:      slugkit.MustNewForTest("claude-code"),
			AgentfileSource: "AGENT claude\nEXECUTABLE claude\nMODE pty\n",
		},
	})
	require.ErrorIs(t, err, ErrMissingAgentAdapter)
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
