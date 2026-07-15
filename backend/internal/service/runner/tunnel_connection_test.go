package runner

import (
	"context"
	"testing"
	"time"

	runnerdomain "github.com/anthropics/agentsmesh/backend/internal/domain/runner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateTunnelConnection(t *testing.T) {
	db := setupTestDB(t)
	service := newTestService(db)
	ctx := context.Background()
	oldError := "old_error"
	runner := &runnerdomain.Runner{
		OrganizationID:    1,
		NodeID:            "tunnel-runner",
		Status:            runnerdomain.RunnerStatusOnline,
		TunnelState:       tunnelStateDisconnected,
		TunnelLastError:   &oldError,
		MaxConcurrentPods: 1,
	}
	require.NoError(t, db.Create(runner).Error)
	service.activeRunners.Store(runner.ID, &ActiveRunner{
		Runner:   runner,
		LastPing: time.Now().Add(-time.Minute),
		PodCount: 0,
	})

	require.NoError(t, service.UpdateTunnelConnection(ctx, runner.ID, true, ""))

	var connected runnerdomain.Runner
	require.NoError(t, db.First(&connected, runner.ID).Error)
	assert.Equal(t, tunnelStateConnected, connected.TunnelState)
	assert.NotNil(t, connected.TunnelLastSeenAt)
	assert.Nil(t, connected.TunnelLastError)
	activeConnected, err := service.GetRunner(ctx, runner.ID)
	require.NoError(t, err)
	assert.Equal(t, tunnelStateConnected, activeConnected.TunnelState)
	assert.NotNil(t, activeConnected.TunnelLastSeenAt)
	assert.Nil(t, activeConnected.TunnelLastError)

	require.NoError(t, service.UpdateTunnelConnection(ctx, runner.ID, false, "gateway_unavailable"))

	var disconnected runnerdomain.Runner
	require.NoError(t, db.First(&disconnected, runner.ID).Error)
	assert.Equal(t, tunnelStateDisconnected, disconnected.TunnelState)
	assert.NotNil(t, disconnected.TunnelLastSeenAt)
	require.NotNil(t, disconnected.TunnelLastError)
	assert.Equal(t, "gateway_unavailable", *disconnected.TunnelLastError)
	activeDisconnected, err := service.GetRunner(ctx, runner.ID)
	require.NoError(t, err)
	assert.Equal(t, tunnelStateDisconnected, activeDisconnected.TunnelState)
	require.NotNil(t, activeDisconnected.TunnelLastError)
	assert.Equal(t, "gateway_unavailable", *activeDisconnected.TunnelLastError)
}
