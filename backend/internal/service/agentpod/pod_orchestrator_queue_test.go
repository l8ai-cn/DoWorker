package agentpod

import (
	"context"
	"errors"
	"testing"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func queueTestRequest() *OrchestrateCreatePodRequest {
	return &OrchestrateCreatePodRequest{
		OrganizationID:  1,
		UserID:          1,
		RunnerID:        9,
		AgentSlug:       "claude-code",
		ModelResourceID: testModelResourceID(),
		Cols:            80, Rows: 24,
	}
}

func TestOrchestrate_QueueIfUnavailable_PodCreatedAsQueued(t *testing.T) {
	coord := &mockPodCoordinator{queueErr: podDomain.ErrPodQueued}
	orch, podSvc, _ := setupOrchestrator(t, withCoordinator(coord))

	req := queueTestRequest()
	req.QueueIfUnavailable = true
	result, err := orch.CreatePod(context.Background(), req)
	require.NoError(t, err)
	require.True(t, result.Queued)

	pod, err := podSvc.GetPod(context.Background(), result.Pod.PodKey)
	require.NoError(t, err)
	assert.Equal(t, podDomain.StatusQueued, pod.Status)
	assert.Nil(t, pod.ErrorCode)
	assert.True(t, coord.lastQueueOpts.Queue)
}

func TestOrchestrate_Default_BehaviorUnchanged(t *testing.T) {
	coord := &mockPodCoordinator{}
	orch, podSvc, _ := setupOrchestrator(t, withCoordinator(coord))

	result, err := orch.CreatePod(context.Background(), queueTestRequest())
	require.NoError(t, err)
	assert.False(t, result.Queued)
	assert.True(t, coord.createPodCalled)
	assert.False(t, coord.lastQueueOpts.Queue)

	pod, err := podSvc.GetPod(context.Background(), result.Pod.PodKey)
	require.NoError(t, err)
	assert.Equal(t, podDomain.StatusInitializing, pod.Status)
}

func TestOrchestrate_QueueFull_MarksPodError(t *testing.T) {
	coord := &mockPodCoordinator{queueErr: podDomain.ErrQueueFull}
	orch, podSvc, db := setupOrchestrator(t, withCoordinator(coord))

	req := queueTestRequest()
	req.QueueIfUnavailable = true
	_, err := orch.CreatePod(context.Background(), req)
	require.ErrorIs(t, err, podDomain.ErrQueueFull)

	var podKey string
	require.NoError(t, db.Raw(`SELECT pod_key FROM pods ORDER BY id DESC LIMIT 1`).Scan(&podKey).Error)
	pod, err := podSvc.GetPod(context.Background(), podKey)
	require.NoError(t, err)
	assert.Equal(t, podDomain.StatusError, pod.Status)
	require.NotNil(t, pod.ErrorCode)
	assert.Equal(t, "QUEUE_FULL", *pod.ErrorCode)
}

func TestOrchestrate_DispatchFailure_MarksPodError(t *testing.T) {
	coord := &mockPodCoordinator{queueErr: errors.New("runner not connected")}
	orch, podSvc, db := setupOrchestrator(t, withCoordinator(coord))

	req := queueTestRequest()
	req.QueueIfUnavailable = true
	_, err := orch.CreatePod(context.Background(), req)
	require.ErrorIs(t, err, ErrRunnerDispatchFailed)

	var podKey string
	require.NoError(t, db.Raw(`SELECT pod_key FROM pods ORDER BY id DESC LIMIT 1`).Scan(&podKey).Error)
	pod, err := podSvc.GetPod(context.Background(), podKey)
	require.NoError(t, err)
	assert.Equal(t, podDomain.StatusError, pod.Status)
	require.NotNil(t, pod.ErrorCode)
	assert.Equal(t, "RUNNER_UNREACHABLE", *pod.ErrorCode)
}

func TestInitialPodStatusQueuesDeferredWorkerLaunchWhenRunnerReady(t *testing.T) {
	launchID := int64(71)
	orchestrator := &PodOrchestrator{
		podCoordinator: &readyPodCoordinator{},
	}

	status := orchestrator.initialPodStatus(&OrchestrateCreatePodRequest{
		RunnerID:                    9,
		QueueIfUnavailable:          true,
		DeferRunnerDispatch:         true,
		OrchestrationWorkerLaunchID: &launchID,
	})

	assert.Equal(t, podDomain.StatusQueued, status)
}

type readyPodCoordinator struct {
	mockPodCoordinator
}

func (*readyPodCoordinator) ShouldDispatchNow(context.Context, int64) bool {
	return true
}
