package runner

import (
	"context"
	"testing"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	runnerDomain "github.com/anthropics/agentsmesh/backend/internal/domain/runner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPodCoordinator_MarkStaleAsDisconnected_Notifies(t *testing.T) {
	pc, _, _, db := setupPodEventHandlerDeps(t)

	r := &runnerDomain.Runner{OrganizationID: 1, NodeID: "stale-node", Status: "online"}
	require.NoError(t, db.Create(r).Error)

	staleAt := time.Now().Add(-2 * time.Hour)
	db.Exec(`INSERT INTO pods (pod_key, runner_id, status, agent_status, last_activity) VALUES (?, ?, ?, ?, ?)`,
		"stale-run-1", r.ID, agentpod.StatusRunning, "idle", staleAt)

	recorder := &statusChangeRecorder{}
	pc.SetStatusChangeCallback(recorder.callback)

	threshold := time.Now().Add(-30 * time.Minute)
	n, err := pc.MarkStaleAsDisconnected(context.Background(), threshold)
	require.NoError(t, err)
	assert.Equal(t, int64(1), n)

	var pod agentpod.Pod
	require.NoError(t, db.Where("pod_key = ?", "stale-run-1").First(&pod).Error)
	assert.Equal(t, agentpod.StatusDisconnected, pod.Status)

	calls := recorder.getCalls()
	require.Len(t, calls, 1)
	assert.Equal(t, "stale-run-1", calls[0].PodKey)
	assert.Equal(t, agentpod.StatusDisconnected, calls[0].Status)
}

func TestPodCoordinator_CleanupStaleTerminal_Notifies(t *testing.T) {
	pc, _, _, db := setupPodEventHandlerDeps(t)

	r := &runnerDomain.Runner{OrganizationID: 1, NodeID: "stale-term-node", Status: "online"}
	require.NoError(t, db.Create(r).Error)

	staleAt := time.Now().Add(-48 * time.Hour)
	db.Exec(`INSERT INTO pods (pod_key, runner_id, status, agent_status, last_activity) VALUES (?, ?, ?, ?, ?)`,
		"stale-disc-1", r.ID, agentpod.StatusDisconnected, "idle", staleAt)

	recorder := &statusChangeRecorder{}
	pc.SetStatusChangeCallback(recorder.callback)

	threshold := time.Now().Add(-24 * time.Hour)
	n, err := pc.CleanupStaleTerminal(context.Background(), threshold)
	require.NoError(t, err)
	assert.Equal(t, int64(1), n)

	var pod agentpod.Pod
	require.NoError(t, db.Where("pod_key = ?", "stale-disc-1").First(&pod).Error)
	assert.Equal(t, agentpod.StatusTerminated, pod.Status)

	calls := recorder.getCalls()
	require.Len(t, calls, 1)
	assert.Equal(t, "stale-disc-1", calls[0].PodKey)
	assert.Equal(t, agentpod.StatusTerminated, calls[0].Status)
}
