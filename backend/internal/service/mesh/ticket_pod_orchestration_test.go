package mesh

import (
	"context"
	"testing"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	meshDomain "github.com/anthropics/agentsmesh/backend/internal/domain/mesh"
	podService "github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
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
		OrganizationID: 1,
		TicketID:       2,
		RunnerID:       3,
		CreatedByID:    4,
		Prompt:         "repair failing test",
	})

	require.NoError(t, err)
	assert.Equal(t, "ticket-pod", pod.PodKey)
	require.NotNil(t, creator.request)
	assert.Equal(t, int64(3), creator.request.RunnerID)
	assert.Equal(t, meshDomain.LegacyTicketPodAgentSlug, creator.request.AgentSlug)
	assert.Equal(t, `MODE pty
CONFIG model = "opus"
CONFIG permission_mode = "bypassPermissions"
PROMPT "repair failing test"`, *creator.request.AgentfileLayer)
}
