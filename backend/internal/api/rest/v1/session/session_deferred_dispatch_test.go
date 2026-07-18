package sessionapi

import (
	"net/http"
	"testing"
	"time"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	itemDomain "github.com/anthropics/agentsmesh/backend/internal/domain/conversationitem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestForkSessionTriggersDrainAfterCopiedItemsAndCommandPersist(t *testing.T) {
	deps, db, _ := setupSessionCreationCompensationTest(t)
	seedSourceSession(t, deps, &itemDomain.Item{
		ID: "item_source", SessionID: "conv_source", ItemType: "message",
		ResponseID: "resp_source", Status: "completed", Position: 1,
		Payload: []byte(`{"id":"item_source","type":"message"}`), CreatedAt: time.Now(),
	})
	orchestrator := deps.PodOrchestrator.(*fixedSessionPodOrchestrator)
	queue := deps.DispatchQueue.(*recordingSessionDispatchQueue)
	queue.beforeTrigger = func(runnerID int64) {
		assert.Equal(t, int64(3), runnerID)
		assert.Equal(t, int64(2), activeSessionCount(t, db))
		assert.Equal(t, int64(2), conversationItemCount(t, db))
		assert.Equal(t, int64(1), pendingCommandCount(t, db))
	}

	response := forkSessionCompensationRequest(deps)

	assert.Equal(t, http.StatusOK, response.Code)
	require.NotNil(t, orchestrator.request)
	assert.True(t, orchestrator.request.DeferRunnerDispatch)
	assert.Equal(t, 0, orchestrator.dispatches)
	assert.Equal(t, []int64{3}, queue.triggers)
}

func TestImportSessionTriggersDrainAfterImportedItemsAndCommandPersist(t *testing.T) {
	deps, db, _ := setupSessionCreationCompensationTest(t)
	orchestrator := deps.PodOrchestrator.(*fixedSessionPodOrchestrator)
	queue := deps.DispatchQueue.(*recordingSessionDispatchQueue)
	queue.beforeTrigger = func(runnerID int64) {
		assert.Equal(t, int64(3), runnerID)
		assert.Equal(t, int64(1), activeSessionCount(t, db))
		assert.Equal(t, int64(1), conversationItemCount(t, db))
		assert.Equal(t, int64(1), pendingCommandCount(t, db))
	}

	response := importSessionCompensationRequest(t, deps)

	assert.Equal(t, http.StatusOK, response.Code)
	require.NotNil(t, orchestrator.request)
	assert.True(t, orchestrator.request.DeferRunnerDispatch)
	assert.Equal(t, 0, orchestrator.dispatches)
	assert.Equal(t, []int64{3}, queue.triggers)
}

func TestCreateSessionTerminatesPodWhenDeferredCommandIsMissing(t *testing.T) {
	deps, db, lifecycle := setupSessionCreationCompensationTest(t)
	orchestrator := deps.PodOrchestrator.(*fixedSessionPodOrchestrator)
	orchestrator.result.DeferredCreateCommand = nil

	response := createSessionCompensationRequest(deps, `{"agent_id":"codex-cli"}`)

	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.Equal(t, int64(0), activeSessionCount(t, db))
	assert.Equal(t, int64(0), pendingCommandCount(t, db))
	assert.Equal(t, []string{"new-pod"}, lifecycle.terminated)
	assert.Empty(t, deps.DispatchQueue.(*recordingSessionDispatchQueue).triggers)
}

func TestCreateSessionRejectsOfflineCommandWhenQueueIsDisabled(t *testing.T) {
	deps, db, lifecycle := setupSessionCreationCompensationTest(t)
	queue := deps.DispatchQueue.(*recordingSessionDispatchQueue)
	queue.enabled = false

	response := createSessionCompensationRequest(deps, `{"agent_id":"codex-cli"}`)

	assert.Equal(t, http.StatusServiceUnavailable, response.Code)
	assert.JSONEq(t, `{
		"error":"runner unavailable",
		"code":"runner_unavailable"
	}`, response.Body.String())
	assert.Equal(t, []string{"new-pod"}, lifecycle.terminated)
	assert.Equal(t, int64(0), activeSessionCount(t, db))
	assert.Equal(t, int64(0), pendingCommandCount(t, db))
	assert.Empty(t, queue.triggers)
}

func TestCreateSessionAllowsCrashRecoveryWhenQueueDisabledButRunnerConnected(t *testing.T) {
	deps, db, _ := setupSessionCreationCompensationTest(t)
	queue := deps.DispatchQueue.(*recordingSessionDispatchQueue)
	queue.enabled = false
	queue.connected = true

	response := createSessionCompensationRequest(deps, `{"agent_id":"codex-cli"}`)

	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, int64(1), activeSessionCount(t, db))
	assert.Equal(t, int64(1), pendingCommandCount(t, db))
	assert.Equal(t, []int64{3}, queue.triggers)
}

func TestCreateSessionRejectsFullRunnerQueueAtomically(t *testing.T) {
	deps, db, lifecycle := setupSessionCreationCompensationTest(t)
	queue := deps.DispatchQueue.(*recordingSessionDispatchQueue)
	queue.maxPerRunner = 1
	require.NoError(t, db.Create(&podDomain.PendingCommand{
		OrganizationID: 21,
		RunnerID:       3,
		PodKey:         "existing-pod",
		CommandType:    podDomain.CommandTypeCreatePod,
		CommandID:      "existing-pod",
		Payload:        []byte{1},
		ExpiresAt:      time.Now().Add(time.Hour),
		CreatedAt:      time.Now(),
	}).Error)

	response := createSessionCompensationRequest(deps, `{"agent_id":"codex-cli"}`)

	assert.Equal(t, http.StatusTooManyRequests, response.Code)
	assert.JSONEq(t, `{
		"error":"runner queue full",
		"code":"runner_queue_full"
	}`, response.Body.String())
	assert.Equal(t, []string{"new-pod"}, lifecycle.terminated)
	assert.Equal(t, int64(0), activeSessionCount(t, db))
	assert.Equal(t, int64(1), pendingCommandCount(t, db))
	assert.Empty(t, queue.triggers)
}
