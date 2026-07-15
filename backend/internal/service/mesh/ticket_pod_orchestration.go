package mesh

import (
	"context"
	"errors"
	"strings"

	"github.com/anthropics/agentsmesh/agentfile"
	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	meshDomain "github.com/anthropics/agentsmesh/backend/internal/domain/mesh"
	podService "github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
)

var ErrPodCreatorNotConfigured = errors.New("pod creator is not configured")

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
	result, err := s.podCreator.CreatePod(ctx, &podService.OrchestrateCreatePodRequest{
		OrganizationID:  req.OrganizationID,
		UserID:          req.CreatedByID,
		RunnerID:        req.RunnerID,
		AgentSlug:       meshDomain.LegacyTicketPodAgentSlug,
		TicketID:        &req.TicketID,
		AgentfileLayer:  legacyTicketPodAgentfile(req),
		AutomationLevel: podDomain.AutomationLevelAutonomous,
	})
	if err != nil {
		return nil, err
	}
	return result.Pod, nil
}

func legacyTicketPodAgentfile(req *meshDomain.CreatePodForTicketRequest) *string {
	model := req.Model
	if model == "" {
		model = meshDomain.LegacyTicketPodModel
	}
	permissionMode := req.PermissionMode
	if permissionMode == "" {
		permissionMode = meshDomain.LegacyTicketPodPermissionMode
	}
	lines := []string{
		"MODE pty",
		"CONFIG model = " + agentfile.FormatStringLiteral(model),
		"CONFIG permission_mode = " + agentfile.FormatStringLiteral(permissionMode),
	}
	if req.Prompt != "" {
		lines = append(lines, "PROMPT "+agentfile.FormatStringLiteral(req.Prompt))
	}
	layer := strings.Join(lines, "\n")
	return &layer
}
