package runner

import (
	"context"
	"testing"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/agentpod"
	runnerv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/runner/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreatePodOrQueue_OnlineDispatchesDirectly(t *testing.T) {
	logger := newTestLogger()
	db, cm, tr, hb, podStore, runnerRepo := setupPodCoordinatorDeps(t)
	r := createOnlineRunner(t, db, runnerRepo, 0, 5)
	stream := newMockRunnerStreamWithTesting(t)
	rc := cm.AddConnection(r.ID, "node", "org", stream)
	rc.SetInitialized(true, []string{"codex-cli"})

	pc := NewPodCoordinator(podStore, runnerRepo, cm, tr, hb, logger)
	pc.SetConnectionChecker(cm)
	mockSender := &MockCommandSender{}
	pc.SetCommandSender(mockSender)
	repo := &memPendingRepo{}
	pc.SetPendingQueue(NewPendingCommandQueue(repo, nil, 5, 0, true, testPendingEncryptor(), logger))

	cmd := &runnerv1.CreatePodCommand{PodKey: "pd-online"}
	err := pc.CreatePodOrQueue(context.Background(), r.ID, cmd, agentpod.CreatePodQueueOpts{Queue: true, OrgID: 1})
	require.NoError(t, err)
	assert.Equal(t, 1, mockSender.CreatePodCalls)
	assert.Empty(t, repo.rows)
}

func TestCreatePodOrQueue_OfflineQueueFalse_FailsAsToday(t *testing.T) {
	logger := newTestLogger()
	db, cm, tr, hb, podStore, runnerRepo := setupPodCoordinatorDeps(t)
	r := createOnlineRunner(t, db, runnerRepo, 0, 5)
	pc := NewPodCoordinator(podStore, runnerRepo, cm, tr, hb, logger)
	pc.SetConnectionChecker(cm)
	pc.SetCommandSender(&MockCommandSender{})

	cmd := &runnerv1.CreatePodCommand{PodKey: "pd-off"}
	err := pc.CreatePodOrQueue(context.Background(), r.ID, cmd, agentpod.CreatePodQueueOpts{Queue: false, OrgID: 1})
	assert.ErrorIs(t, err, ErrRunnerNotConnected)
}

func TestCreatePodOrQueue_OfflineQueueTrue_ReturnsErrPodQueued(t *testing.T) {
	logger := newTestLogger()
	db, cm, tr, hb, podStore, runnerRepo := setupPodCoordinatorDeps(t)
	r := createOnlineRunner(t, db, runnerRepo, 0, 5)
	pc := NewPodCoordinator(podStore, runnerRepo, cm, tr, hb, logger)
	pc.SetConnectionChecker(cm)
	repo := &memPendingRepo{}
	pc.SetPendingQueue(NewPendingCommandQueue(repo, nil, 5, 0, true, testPendingEncryptor(), logger))

	cmd := &runnerv1.CreatePodCommand{PodKey: "pd-q"}
	err := pc.CreatePodOrQueue(context.Background(), r.ID, cmd, agentpod.CreatePodQueueOpts{Queue: true, OrgID: 1})
	assert.ErrorIs(t, err, agentpod.ErrPodQueued)
	assert.Len(t, repo.rows, 1)
}

func TestCreatePodOrQueue_BusyQueueTrue_Enqueues(t *testing.T) {
	logger := newTestLogger()
	db, cm, tr, hb, podStore, runnerRepo := setupPodCoordinatorDeps(t)
	r := createOnlineRunner(t, db, runnerRepo, 5, 5)
	stream := newMockRunnerStreamWithTesting(t)
	rc := cm.AddConnection(r.ID, "node", "org", stream)
	rc.SetInitialized(true, []string{"codex-cli"})

	pc := NewPodCoordinator(podStore, runnerRepo, cm, tr, hb, logger)
	pc.SetConnectionChecker(cm)
	pc.SetCommandSender(&MockCommandSender{})
	repo := &memPendingRepo{}
	pc.SetPendingQueue(NewPendingCommandQueue(repo, nil, 5, 0, true, testPendingEncryptor(), logger))

	cmd := &runnerv1.CreatePodCommand{PodKey: "pd-busy"}
	err := pc.CreatePodOrQueue(context.Background(), r.ID, cmd, agentpod.CreatePodQueueOpts{Queue: true, OrgID: 1})
	assert.ErrorIs(t, err, agentpod.ErrPodQueued)
}
