package coordinator

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/anthropics/agentsmesh/agentfile/serialize"
	coordinatordom "github.com/anthropics/agentsmesh/backend/internal/domain/coordinator"
	ticketDomain "github.com/anthropics/agentsmesh/backend/internal/domain/ticket"
	agentpodSvc "github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
	ticketSvc "github.com/anthropics/agentsmesh/backend/internal/service/ticket"
)

// claimAndDispatch claims a single task on the platform, syncs it to a ticket,
// spawns a do-agent pod, and records the execution. Returns false (no error)
// when the task was already synced or is owned by another coordinator.
func (s *Service) claimAndDispatch(ctx context.Context, project *coordinatordom.Project, platform TaskPlatform, repo string, task ExternalTask) (bool, error) {
	if _, err := s.store.GetLinkByExternalID(ctx, project.OrganizationID, project.PlatformType, task.ExternalID); err == nil {
		return false, nil
	} else if !errors.Is(err, coordinatordom.ErrNotFound) {
		return false, fmt.Errorf("lookup external link: %w", err)
	}

	claimKey := fmt.Sprintf("project=%d task=%s", project.ID, task.ExternalID)
	claim, err := platform.TryClaim(ctx, repo, task, claimKey)
	if err != nil {
		return false, fmt.Errorf("claim: %w", err)
	}
	if !claim.Claimed {
		s.logger.Debug("task not claimed", "external_id", task.ExternalID, "reason", claim.Reason)
		return false, nil
	}

	ticket, err := s.syncTicket(ctx, project, task)
	if err != nil {
		return false, fmt.Errorf("sync ticket: %w", err)
	}

	pod, err := s.dispatch.CreatePod(ctx, &agentpodSvc.OrchestrateCreatePodRequest{
		OrganizationID: project.OrganizationID,
		UserID:         project.CreatedByID,
		AgentSlug:      project.AgentSlug,
		RepositoryID:   &project.RepositoryID,
		TicketID:       &ticket.ID,
		AgentfileLayer: ptr(buildAgentfileLayer(repo, task)),
		Cols:           120,
		Rows:           40,
	})
	if err != nil {
		return false, fmt.Errorf("dispatch pod: %w", err)
	}

	execution := &coordinatordom.Execution{
		OrganizationID: project.OrganizationID,
		ProjectID:      project.ID,
		TicketID:       ticket.ID,
		PodID:          &pod.Pod.ID,
		PodKey:         &pod.Pod.PodKey,
		Status:         coordinatordom.ExecutionStatusRunning,
		Stage:          "dispatched",
		ClaimMarker:    claim.Marker,
		ExternalID:     task.ExternalID,
		StartedAt:      &pod.Pod.CreatedAt,
	}
	if err := s.store.CreateExecution(ctx, execution); err != nil {
		return false, fmt.Errorf("create execution: %w", err)
	}
	s.logger.Info("dispatched coordinator task",
		"project_id", project.ID, "ticket_id", ticket.ID, "pod_key", pod.Pod.PodKey, "external_id", task.ExternalID)
	return true, nil
}

func (s *Service) syncTicket(ctx context.Context, project *coordinatordom.Project, task ExternalTask) (*ticketDomain.Ticket, error) {
	repoID := project.RepositoryID
	content := task.Description
	created, err := s.tickets.CreateTicket(ctx, &ticketSvc.CreateTicketRequest{
		OrganizationID: project.OrganizationID,
		RepositoryID:   &repoID,
		ReporterID:     project.CreatedByID,
		Title:          ticketTitle(task),
		Content:        &content,
		Status:         ticketDomain.TicketStatusInProgress,
		Labels:         task.Labels,
	})
	if err != nil {
		return nil, err
	}
	link := &coordinatordom.TicketExternalLink{
		OrganizationID: project.OrganizationID,
		TicketID:       created.ID,
		PlatformType:   project.PlatformType,
		ExternalID:     task.ExternalID,
		ExternalURL:    task.URL,
	}
	if err := s.store.CreateLink(ctx, link); err != nil {
		return nil, fmt.Errorf("create external link: %w", err)
	}
	return created, nil
}

func ticketTitle(task ExternalTask) string {
	title := strings.TrimSpace(task.Title)
	if title == "" {
		return task.ExternalID
	}
	return title
}

func buildAgentfileLayer(repo string, task ExternalTask) string {
	var lines []string
	prompt := strings.TrimSpace(task.Title + "\n\n" + task.Description)
	lines = append(lines, fmt.Sprintf("PROMPT %s", serialize.QuoteString(prompt)))
	if repo != "" {
		lines = append(lines, fmt.Sprintf("REPO %s", serialize.QuoteString(repo)))
	}
	return strings.Join(lines, "\n")
}

func ptr[T any](v T) *T { return &v }
