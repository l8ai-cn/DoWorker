package main

import (
	"context"
	"testing"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	"github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
	workerplanner "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationworker"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestOrchestrationWorkerPodLauncherMaterializesDeferredSnapshotPod(
	t *testing.T,
) {
	orchestrator := &workerPodOrchestratorStub{
		result: &agentpod.OrchestrateCreatePodResult{
			Pod: &podDomain.Pod{
				ID: 73, PodKey: "7-standalone-12345678", RunnerID: 11,
			},
			DeferredCreateCommand: &runnerv1.CreatePodCommand{
				PodKey: "7-standalone-12345678",
			},
		},
	}
	launcher := newOrchestrationWorkerPodLauncher(orchestrator)
	prompt := "Review authorization"

	launch, err := launcher.MaterializeWorkerPod(
		context.Background(),
		workerplanner.WorkerLaunchClaim{
			LaunchID:             71,
			OrganizationID:       42,
			ActorID:              7,
			WorkerSpecSnapshotID: 91,
			Prompt:               &prompt,
			Alias:                "reviewer-42",
		},
	)
	require.NoError(t, err)
	require.NotNil(t, orchestrator.request)
	assert.Equal(t, int64(42), orchestrator.request.OrganizationID)
	assert.Equal(t, int64(7), orchestrator.request.UserID)
	assert.Equal(t, int64(91), *orchestrator.request.WorkerSpecSnapshotID)
	assert.Equal(t, int64(71), *orchestrator.request.OrchestrationWorkerLaunchID)
	assert.Equal(t, prompt, *orchestrator.request.WorkerSpecPromptOverride)
	assert.Equal(t, "reviewer-42", *orchestrator.request.Alias)
	assert.True(t, orchestrator.request.DeferRunnerDispatch)
	assert.True(t, orchestrator.request.QueueIfUnavailable)
	assert.Equal(t, int64(73), launch.PodID)
	assert.Equal(t, int64(11), launch.RunnerID)

	var message runnerv1.ServerMessage
	require.NoError(t, proto.Unmarshal(launch.CommandPayload, &message))
	require.NotNil(t, message.GetCreatePod())
	assert.Equal(t, launch.PodKey, message.GetCreatePod().GetPodKey())
}

func TestOrchestrationWorkerDispatchNotifierTriggersQueue(t *testing.T) {
	queue := &workerDispatchQueueStub{}

	newOrchestrationWorkerDispatchNotifier(queue).
		TriggerWorkerDispatch(11)

	assert.Equal(t, int64(11), queue.runnerID)
}

type workerPodOrchestratorStub struct {
	request *agentpod.OrchestrateCreatePodRequest
	result  *agentpod.OrchestrateCreatePodResult
	err     error
}

func (stub *workerPodOrchestratorStub) CreatePod(
	_ context.Context,
	request *agentpod.OrchestrateCreatePodRequest,
) (*agentpod.OrchestrateCreatePodResult, error) {
	stub.request = request
	return stub.result, stub.err
}

type workerDispatchQueueStub struct {
	runnerID int64
}

func (stub *workerDispatchQueueStub) TriggerDrain(runnerID int64) {
	stub.runnerID = runnerID
}
