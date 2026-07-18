package mesh

import (
	"context"
	"errors"
	"strings"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	meshDomain "github.com/anthropics/agentsmesh/backend/internal/domain/mesh"
	podService "github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
)

var (
	ErrPodCreatorNotConfigured    = errors.New("pod creator is not configured")
	ErrWorkerSpecSnapshotRequired = errors.New("worker spec snapshot is required")
)

type PodCreator interface {
	CreatePod(context.Context, *podService.OrchestrateCreatePodRequest) (*podService.OrchestrateCreatePodResult, error)
}

func (s *Service) SetPodCreator(creator PodCreator) {
	s.podCreator = creator
}

func (s *Service) CreatePodForTicket(ctx context.Context, req *meshDomain.CreatePodForTicketRequest) (*podDomain.Pod, error) {
	if s.podCreator == nil {
		return nil, ErrPodCreatorNotConfigured
	}
	if req.WorkerSpecSnapshotID <= 0 {
		return nil, ErrWorkerSpecSnapshotRequired
	}
	snapshotID := req.WorkerSpecSnapshotID
	ticketID := req.TicketID
	var promptOverride *string
	if prompt := strings.TrimSpace(req.Prompt); prompt != "" {
		promptOverride = &prompt
	}
	result, err := s.podCreator.CreatePod(ctx, &podService.OrchestrateCreatePodRequest{
		OrganizationID:           req.OrganizationID,
		UserID:                   req.CreatedByID,
		TicketID:                 &ticketID,
		WorkerSpecSnapshotID:     &snapshotID,
		WorkerSpecPromptOverride: promptOverride,
	})
	if err != nil {
		return nil, err
	}
	return result.Pod, nil
}
