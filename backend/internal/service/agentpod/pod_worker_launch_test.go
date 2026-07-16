package agentpod

import (
	"context"
	"testing"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	"github.com/anthropics/agentsmesh/backend/internal/infra"
	"github.com/anthropics/agentsmesh/backend/internal/testkit"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPodServiceReusesOrchestrationWorkerLaunchPod(t *testing.T) {
	db := testkit.SetupTestDB(t)
	service := NewPodService(infra.NewPodRepository(db))
	launchID := int64(71)
	snapshotID := int64(91)
	request := workerLaunchPodRequest(launchID, snapshotID)

	first, err := service.CreatePod(context.Background(), request)
	require.NoError(t, err)
	second, err := service.CreatePod(
		context.Background(),
		workerLaunchPodRequest(launchID, snapshotID),
	)
	require.NoError(t, err)

	assert.Equal(t, first.ID, second.ID)
	assert.Equal(t, first.PodKey, second.PodKey)
	var count int64
	require.NoError(t, db.Model(&podDomain.Pod{}).Count(&count).Error)
	assert.Equal(t, int64(1), count)
}

func TestPodServiceRejectsWorkerLaunchSnapshotMismatch(t *testing.T) {
	db := testkit.SetupTestDB(t)
	service := NewPodService(infra.NewPodRepository(db))
	launchID := int64(71)

	_, err := service.CreatePod(
		context.Background(),
		workerLaunchPodRequest(launchID, 91),
	)
	require.NoError(t, err)
	_, err = service.CreatePod(
		context.Background(),
		workerLaunchPodRequest(launchID, 92),
	)

	assert.ErrorIs(t, err, ErrWorkerLaunchPodMismatch)
}

func TestDispatchCreatedPodReturnsDeferredCreateCommand(t *testing.T) {
	pod := &podDomain.Pod{
		PodKey: "7-standalone-12345678",
	}
	command := &runnerv1.CreatePodCommand{PodKey: pod.PodKey}

	result, err := (&PodOrchestrator{}).dispatchCreatedPod(
		context.Background(),
		&OrchestrateCreatePodRequest{DeferRunnerDispatch: true},
		pod,
		command,
		"11111111-1111-4111-8111-111111111111",
		false,
	)

	require.NoError(t, err)
	assert.Same(t, command, result.DeferredCreateCommand)
	assert.Same(t, pod, result.Pod)
}

func TestCreatePodRejectsWorkerLaunchWithoutDeferredDispatch(t *testing.T) {
	launchID := int64(71)

	_, err := (&PodOrchestrator{}).CreatePod(
		context.Background(),
		&OrchestrateCreatePodRequest{
			OrganizationID:              42,
			UserID:                      7,
			OrchestrationWorkerLaunchID: &launchID,
		},
	)

	assert.ErrorIs(t, err, ErrWorkerLaunchRequiresDeferredDispatch)
}

func workerLaunchPodRequest(
	launchID int64,
	snapshotID int64,
) *CreatePodRequest {
	return &CreatePodRequest{
		OrganizationID:              42,
		RunnerID:                    12,
		ClusterID:                   13,
		AgentSlug:                   "codex",
		CreatedByID:                 7,
		Prompt:                      "Review authorization",
		Alias:                       stringPointer("reviewer-42"),
		SessionID:                   "11111111-1111-4111-8111-111111111111",
		InteractionMode:             podDomain.InteractionModeACP,
		AutomationLevel:             podDomain.AutomationLevelAutonomous,
		InitialStatus:               podDomain.StatusQueued,
		WorkerSpecSnapshotID:        &snapshotID,
		OrchestrationWorkerLaunchID: &launchID,
	}
}

func stringPointer(value string) *string {
	return &value
}
