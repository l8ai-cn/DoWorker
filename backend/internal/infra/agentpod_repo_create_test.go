package infra

import (
	"context"
	"testing"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/agentpod"
	"github.com/l8ai-cn/agentcloud/backend/internal/testkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPodRepositoryCopiesRunnerCluster(t *testing.T) {
	db := testkit.SetupTestDB(t)
	require.NoError(t, db.Exec(`
INSERT INTO runners (id, organization_id, cluster_id, node_id)
VALUES (1, 77, 700, 'pod-placement-runner')
`).Error)
	pod := &agentpod.Pod{
		OrganizationID:  77,
		PodKey:          "77-pod-placement-aabbccdd",
		RunnerID:        1,
		CreatedByID:     7,
		Status:          agentpod.StatusInitializing,
		AgentStatus:     agentpod.AgentStatusIdle,
		InteractionMode: agentpod.InteractionModeACP,
		AutomationLevel: agentpod.AutomationLevelAutonomous,
	}

	err := (&podRepo{db: db}).Create(context.Background(), pod)

	require.NoError(t, err)
	assert.Equal(t, int64(700), pod.ClusterID)
	var stored agentpod.Pod
	require.NoError(t, db.First(&stored, pod.ID).Error)
	assert.Equal(t, int64(700), stored.ClusterID)
}
