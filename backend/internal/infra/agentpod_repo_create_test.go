package infra

import (
	"context"
	"testing"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	"github.com/anthropics/agentsmesh/backend/internal/testkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPodRepositoryCopiesRunnerCluster(t *testing.T) {
	db := testkit.SetupTestDB(t)
	require.NoError(t, db.Exec(`
INSERT INTO runners (id, organization_id, cluster_id, node_id)
VALUES (1, 77, 700, 'pod-placement-runner')
`).Error)
	pod := testPodForRunner(77, 1, "77-pod-placement-aabbccdd")

	err := (&podRepo{db: db}).Create(context.Background(), pod)

	require.NoError(t, err)
	assert.Equal(t, int64(700), pod.ClusterID)
	var stored agentpod.Pod
	require.NoError(t, db.First(&stored, pod.ID).Error)
	assert.Equal(t, int64(700), stored.ClusterID)
}

func TestPodRepositoryRejectsRunnerFromAnotherOrganization(t *testing.T) {
	db := testkit.SetupTestDB(t)
	require.NoError(t, db.Exec(`
INSERT INTO runners (id, organization_id, cluster_id, node_id)
VALUES (1, 78, 700, 'other-org-runner')
`).Error)
	pod := testPodForRunner(77, 1, "77-cross-org-aabbccdd")

	err := (&podRepo{db: db}).Create(context.Background(), pod)

	require.ErrorContains(t, err, "resolve runner cluster for pod")
	var count int64
	require.NoError(t, db.Model(&agentpod.Pod{}).Count(&count).Error)
	assert.Zero(t, count)
}

func testPodForRunner(organizationID, runnerID int64, podKey string) *agentpod.Pod {
	return &agentpod.Pod{
		OrganizationID:  organizationID,
		PodKey:          podKey,
		RunnerID:        runnerID,
		CreatedByID:     7,
		Status:          agentpod.StatusInitializing,
		AgentStatus:     agentpod.AgentStatusIdle,
		InteractionMode: agentpod.InteractionModeACP,
		AutomationLevel: agentpod.AutomationLevelAutonomous,
	}
}
