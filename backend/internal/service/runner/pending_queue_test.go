package runner

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	"github.com/anthropics/agentsmesh/backend/internal/infra/eventbus"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type memPendingRepo struct {
	mu   sync.Mutex
	rows []*agentpod.PendingCommand
	next int64
}

func (m *memPendingRepo) Enqueue(_ context.Context, cmd *agentpod.PendingCommand) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.enqueueLocked(cmd)
}

func (m *memPendingRepo) EnqueueWithinCapacity(
	_ context.Context,
	cmd *agentpod.PendingCommand,
	maxPerRunner int,
) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, r := range m.rows {
		if r.CommandID == cmd.CommandID {
			return agentpod.ErrDuplicateCommand
		}
	}
	count := 0
	for _, row := range m.rows {
		if row.RunnerID == cmd.RunnerID {
			count++
		}
	}
	if count >= maxPerRunner {
		return agentpod.ErrQueueFull
	}
	return m.enqueueLocked(cmd)
}

func (m *memPendingRepo) enqueueLocked(cmd *agentpod.PendingCommand) error {
	m.next++
	copy := *cmd
	copy.ID = m.next
	m.rows = append(m.rows, &copy)
	return nil
}

func (m *memPendingRepo) ExistsCommandID(_ context.Context, commandID string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, row := range m.rows {
		if row.CommandID == commandID {
			return true, nil
		}
	}
	return false, nil
}

func (m *memPendingRepo) CountByRunner(_ context.Context, runnerID int64) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	n := 0
	for _, r := range m.rows {
		if r.RunnerID == runnerID {
			n++
		}
	}
	return n, nil
}

func (m *memPendingRepo) ListByRunnerFIFO(_ context.Context, runnerID int64, limit int) ([]*agentpod.PendingCommand, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []*agentpod.PendingCommand
	for _, r := range m.rows {
		if r.RunnerID == runnerID {
			out = append(out, r)
		}
	}
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (m *memPendingRepo) Delete(_ context.Context, id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, r := range m.rows {
		if r.ID == id {
			m.rows = append(m.rows[:i], m.rows[i+1:]...)
			return nil
		}
	}
	return nil
}

func (m *memPendingRepo) DeleteByPodKey(_ context.Context, podKey string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var n int64
	kept := m.rows[:0]
	for _, r := range m.rows {
		if r.PodKey == podKey {
			n++
			continue
		}
		kept = append(kept, r)
	}
	m.rows = kept
	return n, nil
}

func (m *memPendingRepo) ListExpired(_ context.Context, now time.Time, limit int) ([]*agentpod.PendingCommand, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []*agentpod.PendingCommand
	for _, r := range m.rows {
		if !r.ExpiresAt.After(now) {
			out = append(out, r)
		}
	}
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (m *memPendingRepo) PositionByPodKey(_ context.Context, runnerID int64, podKey string) (int, error) {
	rows, _ := m.ListByRunnerFIFO(context.Background(), runnerID, 0)
	pos := 0
	for _, r := range rows {
		pos++
		if r.PodKey == podKey {
			return pos, nil
		}
	}
	return 0, nil
}

func (m *memPendingRepo) GetCreatePodByPodKey(_ context.Context, podKey string) (*agentpod.PendingCommand, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, r := range m.rows {
		if r.PodKey == podKey && r.CommandType == agentpod.CommandTypeCreatePod {
			copy := *r
			return &copy, nil
		}
	}
	return nil, nil
}

func (m *memPendingRepo) ListRunnerIDsWithPending(_ context.Context, limit int) ([]int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	seen := make(map[int64]struct{})
	var ids []int64
	for _, r := range m.rows {
		if _, ok := seen[r.RunnerID]; ok {
			continue
		}
		seen[r.RunnerID] = struct{}{}
		ids = append(ids, r.RunnerID)
		if limit > 0 && len(ids) >= limit {
			break
		}
	}
	return ids, nil
}

func TestEnqueueCreatePod_Success(t *testing.T) {
	repo := &memPendingRepo{}
	bus := eventbus.NewEventBus(nil, nil)
	q := NewPendingCommandQueue(repo, bus, 20, 30*time.Minute, true, newTestLogger())
	cmd := &runnerv1.CreatePodCommand{PodKey: "pd-1"}
	exp, err := q.EnqueueCreatePod(context.Background(), 1, 9, "pd-1", cmd, 0)
	require.NoError(t, err)
	assert.False(t, exp.IsZero())
	assert.Equal(t, 1, len(repo.rows))
}

func TestEnqueue_QueueFull(t *testing.T) {
	repo := &memPendingRepo{}
	q := NewPendingCommandQueue(repo, nil, 1, time.Minute, true, newTestLogger())
	cmd := &runnerv1.CreatePodCommand{PodKey: "a"}
	_, err := q.EnqueueCreatePod(context.Background(), 1, 1, "a", cmd, time.Minute)
	require.NoError(t, err)
	_, err = q.EnqueueCreatePod(context.Background(), 1, 1, "b", &runnerv1.CreatePodCommand{PodKey: "b"}, time.Minute)
	assert.ErrorIs(t, err, agentpod.ErrQueueFull)
}

func TestEnqueue_DuplicateCommandID(t *testing.T) {
	repo := &memPendingRepo{}
	q := NewPendingCommandQueue(repo, nil, 5, time.Minute, true, newTestLogger())
	cmd := &runnerv1.CreatePodCommand{PodKey: "same"}
	_, err := q.EnqueueCreatePod(context.Background(), 1, 1, "same", cmd, time.Minute)
	require.NoError(t, err)
	_, err = q.EnqueueCreatePod(context.Background(), 1, 1, "same", cmd, time.Minute)
	assert.ErrorIs(t, err, agentpod.ErrDuplicateCommand)
}

func TestEnqueue_DuplicateCommandWinsOverQueueFull(t *testing.T) {
	repo := &memPendingRepo{}
	q := NewPendingCommandQueue(repo, nil, 1, time.Minute, true, newTestLogger())
	require.NoError(t, q.EnqueueSendPrompt(
		context.Background(), 1, 1, "pod-1", "cmd-1", "first", time.Minute,
	))

	err := q.EnqueueSendPrompt(
		context.Background(), 1, 1, "pod-1", "cmd-1", "duplicate", time.Minute,
	)

	require.ErrorIs(t, err, agentpod.ErrDuplicateCommand)
}

func TestAllowsDurableCommandHonorsOfflineQueueGate(t *testing.T) {
	q := NewPendingCommandQueue(
		&memPendingRepo{},
		nil,
		5,
		time.Minute,
		false,
		newTestLogger(),
	)
	checker := &stubConnChecker{}
	q.SetConnectionChecker(checker)

	assert.False(t, q.AllowsDurableCommand(1))
	checker.connected = true
	assert.True(t, q.AllowsDurableCommand(1))
}

func TestCancelByPodKey_Idempotent(t *testing.T) {
	repo := &memPendingRepo{}
	q := NewPendingCommandQueue(repo, nil, 5, time.Minute, true, newTestLogger())
	_, err := q.EnqueueCreatePod(context.Background(), 1, 1, "pd-c", &runnerv1.CreatePodCommand{PodKey: "pd-c"}, time.Minute)
	require.NoError(t, err)
	require.NoError(t, q.EnqueueSendPrompt(context.Background(), 1, 1, "pd-c", "cmd-9", "hello", time.Minute))

	require.NoError(t, q.CancelByPodKey(context.Background(), "pd-c"))
	assert.Empty(t, repo.rows)
	require.NoError(t, q.CancelByPodKey(context.Background(), "pd-c"))
}

func TestQueuePosition_FIFO(t *testing.T) {
	repo := &memPendingRepo{}
	q := NewPendingCommandQueue(repo, nil, 5, time.Minute, true, newTestLogger())
	for _, key := range []string{"pd-1", "pd-2", "pd-3"} {
		_, err := q.EnqueueCreatePod(context.Background(), 1, 7, key, &runnerv1.CreatePodCommand{PodKey: key}, time.Minute)
		require.NoError(t, err)
	}
	pos, err := q.QueuePosition(context.Background(), 7, "pd-2")
	require.NoError(t, err)
	assert.Equal(t, 2, pos)
}

func TestEnqueue_TTLClamped(t *testing.T) {
	repo := &memPendingRepo{}
	q := NewPendingCommandQueue(repo, nil, 5, 30*time.Minute, true, newTestLogger())
	before := time.Now()
	_, err := q.EnqueueCreatePod(context.Background(), 1, 1, "pd", &runnerv1.CreatePodCommand{PodKey: "pd"}, 25*time.Hour)
	require.NoError(t, err)
	exp := repo.rows[0].ExpiresAt
	assert.True(t, exp.Before(before.Add(25*time.Hour)))
	assert.True(t, exp.Before(before.Add(maxQueueTTL+time.Minute)))
}
