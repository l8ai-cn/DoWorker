package infra

import (
	"context"
	"testing"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	"github.com/anthropics/agentsmesh/backend/internal/testkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupPendingCommandDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testkit.SetupTestDB(t)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS pending_runner_commands (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		organization_id INTEGER NOT NULL,
		runner_id INTEGER NOT NULL,
		pod_key TEXT NOT NULL,
		command_type TEXT NOT NULL,
		command_id TEXT NOT NULL,
		payload BLOB NOT NULL,
		expires_at DATETIME NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`).Error)
	require.NoError(t, db.Exec(
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_pending_cmds_command ON pending_runner_commands (command_id)`,
	).Error)
	return db
}

func newPendingCmd(runnerID int64, podKey, commandID string) *agentpod.PendingCommand {
	return &agentpod.PendingCommand{
		OrganizationID: 1,
		RunnerID:       runnerID,
		PodKey:         podKey,
		CommandType:    agentpod.CommandTypeCreatePod,
		CommandID:      commandID,
		Payload:        append([]byte(agentpod.PendingPayloadPrefix), 0x1),
		ExpiresAt:      time.Now().Add(time.Hour),
	}
}

func TestPendingCommandRepo_Enqueue_DuplicateReturnsSentinel(t *testing.T) {
	repo := NewPendingCommandRepository(setupPendingCommandDB(t))
	ctx := context.Background()

	require.NoError(t, repo.Enqueue(ctx, newPendingCmd(1, "pd-1", "cmd-1")))
	exists, err := repo.ExistsCommandID(ctx, "cmd-1")
	require.NoError(t, err)
	assert.True(t, exists)
	exists, err = repo.ExistsCommandID(ctx, "missing")
	require.NoError(t, err)
	assert.False(t, exists)

	err = repo.Enqueue(ctx, newPendingCmd(1, "pd-1", "cmd-1"))
	assert.ErrorIs(t, err, agentpod.ErrDuplicateCommand)
}

func TestPendingCommandRepo_EnqueueWithinCapacity(t *testing.T) {
	repo := NewPendingCommandRepository(setupPendingCommandDB(t))
	ctx := context.Background()

	require.NoError(t, repo.EnqueueWithinCapacity(
		ctx,
		newPendingCmd(1, "pd-1", "cmd-1"),
		1,
	))
	err := repo.EnqueueWithinCapacity(
		ctx,
		newPendingCmd(1, "pd-2", "cmd-2"),
		1,
	)
	assert.ErrorIs(t, err, agentpod.ErrQueueFull)
	err = repo.EnqueueWithinCapacity(
		ctx,
		newPendingCmd(1, "pd-1", "cmd-1"),
		1,
	)
	assert.ErrorIs(t, err, agentpod.ErrDuplicateCommand)
}

func TestPendingCommandRepo_FIFOAndPosition(t *testing.T) {
	repo := NewPendingCommandRepository(setupPendingCommandDB(t))
	ctx := context.Background()

	for _, id := range []string{"c1", "c2", "c3"} {
		require.NoError(t, repo.Enqueue(ctx, newPendingCmd(7, "pd-"+id, id)))
	}
	rows, err := repo.ListByRunnerFIFO(ctx, 7, 10)
	require.NoError(t, err)
	require.Len(t, rows, 3)
	assert.Equal(t, "c1", rows[0].CommandID)
	assert.Equal(t, "c3", rows[2].CommandID)

	pos, err := repo.PositionByPodKey(ctx, 7, "pd-c2")
	require.NoError(t, err)
	assert.Equal(t, 2, pos)

	ids, err := repo.ListRunnerIDsWithPending(ctx, 10)
	require.NoError(t, err)
	assert.Equal(t, []int64{7}, ids)
}

func TestPendingCommandRepo_DeleteByPodKey(t *testing.T) {
	repo := NewPendingCommandRepository(setupPendingCommandDB(t))
	ctx := context.Background()

	require.NoError(t, repo.Enqueue(ctx, newPendingCmd(5, "pd-x", "x-create")))
	require.NoError(t, repo.Enqueue(ctx, newPendingCmd(5, "pd-y", "y-create")))

	n, err := repo.DeleteByPodKey(ctx, "pd-x")
	require.NoError(t, err)
	assert.Equal(t, int64(1), n)

	count, err := repo.CountByRunner(ctx, 5)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}
