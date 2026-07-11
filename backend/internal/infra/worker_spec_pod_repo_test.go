package infra

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPodRepositoryCreatesWorkerSpecSnapshotAtomically(t *testing.T) {
	db := workerSpecSnapshotDBForContract(t)
	repo := &podRepo{db: db}
	pod := workerSpecPodForRepoTest("7-standalone-aabbccdd")
	revision := workerSpecRevisionForRepoTest()

	err := repo.CreateWithConfigAndWorkerSpec(
		context.Background(),
		pod,
		revision,
		workerSpecSnapshotForContract(t, 77),
	)

	require.NoError(t, err)
	require.NotNil(t, pod.WorkerSpecSnapshotID)
	assert.Positive(t, *pod.WorkerSpecSnapshotID)
	assert.Positive(t, pod.ID)
	assert.Positive(t, revision.ID)
	assert.Equal(t, pod.ID, revision.PodID)
	require.NotNil(t, pod.ActiveConfigRevisionID)
	assert.Equal(t, revision.ID, *pod.ActiveConfigRevisionID)

	var snapshotCount int64
	require.NoError(t, db.Table("worker_spec_snapshots").Count(&snapshotCount).Error)
	assert.Equal(t, int64(1), snapshotCount)
	var stored agentpod.Pod
	require.NoError(t, db.First(&stored, pod.ID).Error)
	assert.Equal(t, pod.WorkerSpecSnapshotID, stored.WorkerSpecSnapshotID)
}

func TestPodRepositoryRollsBackSnapshotWhenConfigRevisionFails(t *testing.T) {
	db := workerSpecSnapshotDBForContract(t)
	require.NoError(t, db.Exec(`
CREATE TRIGGER reject_worker_spec_revision
BEFORE INSERT ON pod_config_revisions
BEGIN
	SELECT RAISE(ABORT, 'forced revision failure');
END
`).Error)
	repo := &podRepo{db: db}

	err := repo.CreateWithConfigAndWorkerSpec(
		context.Background(),
		workerSpecPodForRepoTest("7-standalone-eeff0011"),
		workerSpecRevisionForRepoTest(),
		workerSpecSnapshotForContract(t, 77),
	)

	require.Error(t, err)
	for _, table := range []string{"worker_spec_snapshots", "pods", "pod_config_revisions"} {
		var count int64
		require.NoError(t, db.Table(table).Count(&count).Error)
		assert.Zero(t, count, table)
	}
}

func workerSpecPodForRepoTest(key string) *agentpod.Pod {
	return &agentpod.Pod{
		OrganizationID:  77,
		PodKey:          key,
		RunnerID:        1,
		AgentSlug:       "codex-cli",
		CreatedByID:     7,
		Status:          agentpod.StatusInitializing,
		AgentStatus:     agentpod.AgentStatusIdle,
		InteractionMode: agentpod.InteractionModeACP,
		AutomationLevel: agentpod.AutomationLevelAutonomous,
	}
}

func workerSpecRevisionForRepoTest() *agentpod.PodConfigRevision {
	return &agentpod.PodConfigRevision{
		Revision:       1,
		AgentfileLayer: "MODE acp\n",
		Status:         agentpod.ConfigRevisionStatusActive,
		ConfigSummary:  json.RawMessage(`{}`),
		CreatedByID:    7,
	}
}
