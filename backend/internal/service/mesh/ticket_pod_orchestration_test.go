package mesh

import (
	"context"
	"testing"

	podDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agentpod"
	meshDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/mesh"
	podService "github.com/l8ai-cn/agentcloud/backend/internal/service/agentpod"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type recordingPodCreator struct {
	request *podService.OrchestrateCreatePodRequest
}

func (c *recordingPodCreator) CreatePod(_ context.Context, req *podService.OrchestrateCreatePodRequest) (*podService.OrchestrateCreatePodResult, error) {
	c.request = req
	return &podService.OrchestrateCreatePodResult{Pod: &podDomain.Pod{PodKey: "ticket-pod"}}, nil
}

func TestCreatePodForTicketDelegatesToPodOrchestrator(t *testing.T) {
	repo, _ := setupTestRepo(t)
	creator := &recordingPodCreator{}
	service := NewService(repo, nil, nil, nil)
	service.SetPodCreator(creator)

	pod, err := service.CreatePodForTicket(context.Background(), &meshDomain.CreatePodForTicketRequest{
		OrganizationID:       1,
		TicketID:             2,
		CreatedByID:          4,
		WorkerSpecSnapshotID: 91,
		Prompt:               "repair failing test",
	})

	require.NoError(t, err)
	assert.Equal(t, "ticket-pod", pod.PodKey)
	require.NotNil(t, creator.request)
	require.NotNil(t, creator.request.WorkerSpecSnapshotID)
	assert.Equal(t, int64(91), *creator.request.WorkerSpecSnapshotID)
	assert.Equal(t, int64(2), *creator.request.TicketID)
	assert.Equal(t, "repair failing test", *creator.request.WorkerSpecPromptOverride)
	assert.Zero(t, creator.request.RunnerID)
	assert.Empty(t, creator.request.AgentSlug)
	assert.Nil(t, creator.request.AgentfileLayer)
}
