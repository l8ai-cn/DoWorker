package coordinator

import (
	"context"
	"errors"
	"fmt"
	"strings"

	coordinatordom "github.com/l8ai-cn/agentcloud/backend/internal/domain/coordinator"
	ticketDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/ticket"
	agentpodSvc "github.com/l8ai-cn/agentcloud/backend/internal/service/agentpod"
	ticketSvc "github.com/l8ai-cn/agentcloud/backend/internal/service/ticket"
)

var ErrCoordinatorWorkerSpecSnapshotRequired = errors.New(
	"coordinator: worker spec snapshot is required",
)

func (s *Service) claimAndDispatch(ctx context.Context, project *coordinatordom.Project, platform TaskPlatform, repo string, task ExternalTask) (bool, error) {
	if s.podTerminator == nil {
		return false, errors.New("coordinator: pod terminator unavailable")
	}
	snapshotID, err := coordinatorSnapshotID(project)
	if err != nil {
		return false, err
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

	var execution *coordinatordom.Execution
	created := false
	createdTicketID := int64(0)
	err = s.store.WithinProjectDispatch(ctx, project.ID, func(store coordinatordom.Repository) error {
		if _, activeErr := store.GetActiveExecutionByProjectAndExternalID(
			ctx,
			project.ID,
			task.ExternalID,
		); activeErr == nil {
			return nil
		} else if !errors.Is(activeErr, coordinatordom.ErrNotFound) {
			return fmt.Errorf("lookup active execution: %w", activeErr)
		}
		active, countErr := store.CountActiveExecutions(ctx, project.ID)
		if countErr != nil {
			return fmt.Errorf("count active executions: %w", countErr)
		}
		maxConcurrent := project.MaxConcurrent
		if maxConcurrent <= 0 {
			maxConcurrent = 1
		}
		if active >= int64(maxConcurrent) {
			return nil
		}

		ticket, ticketCreated, ticketErr := s.loadOrCreateTicket(ctx, store, project, task)
		if ticketErr != nil {
			return ticketErr
		}
		if ticketCreated {
			createdTicketID = ticket.ID
		}
		execution = &coordinatordom.Execution{
			OrganizationID: project.OrganizationID,
			ProjectID:      project.ID,
			TicketID:       ticket.ID,
			Status:         coordinatordom.ExecutionStatusClaimed,
			Stage:          "claimed",
			ClaimMarker:    claim.Marker,
			ExternalID:     task.ExternalID,
		}
		if createErr := store.CreateExecution(ctx, execution); createErr != nil {
			return createErr
		}
		created = true
		return nil
	})
	if err != nil {
		cleanupErr := s.deleteCreatedTicket(ctx, createdTicketID)
		return false, fmt.Errorf("create claimed execution: %w", errors.Join(err, cleanupErr))
	}
	if !created {
		return false, nil
	}

	pod, err := s.dispatch.CreatePod(ctx, &agentpodSvc.OrchestrateCreatePodRequest{
		OrganizationID:           project.OrganizationID,
		UserID:                   project.CreatedByID,
		WorkerSpecSnapshotID:     snapshotID,
		TicketID:                 &execution.TicketID,
		WorkerSpecPromptOverride: ptr(buildTaskPrompt(repo, task)),
		Cols:                     120,
		Rows:                     40,
	})
	if err != nil {
		return false, s.failClaimedExecution(ctx, execution, "dispatch_failed", fmt.Errorf("dispatch pod: %w", err))
	}

	if err := s.attachExecutionPod(ctx, execution, pod); err != nil {
		return false, err
	}
	s.logger.Info("dispatched coordinator task",
		"project_id", project.ID, "ticket_id", execution.TicketID, "pod_key", pod.Pod.PodKey, "external_id", task.ExternalID)
	return true, nil
}

func (s *Service) loadOrCreateTicket(
	ctx context.Context,
	store coordinatordom.Repository,
	project *coordinatordom.Project,
	task ExternalTask,
) (*ticketDomain.Ticket, bool, error) {
	link, err := store.GetLinkByExternalID(
		ctx,
		project.OrganizationID,
		project.PlatformType,
		task.ExternalID,
	)
	if err == nil {
		ticket, loadErr := s.tickets.GetTicket(ctx, link.TicketID)
		if loadErr != nil {
			return nil, false, fmt.Errorf("load linked ticket: %w", loadErr)
		}
		return ticket, false, nil
	}
	if !errors.Is(err, coordinatordom.ErrNotFound) {
		return nil, false, fmt.Errorf("lookup external link: %w", err)
	}

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
		return nil, false, err
	}
	createdLink := &coordinatordom.TicketExternalLink{
		OrganizationID: project.OrganizationID,
		TicketID:       created.ID,
		PlatformType:   project.PlatformType,
		ExternalID:     task.ExternalID,
		ExternalURL:    task.URL,
	}
	if err := store.CreateLink(ctx, createdLink); err != nil {
		return nil, false, fmt.Errorf("create external link: %w", err)
	}
	return created, true, nil
}

func (s *Service) deleteCreatedTicket(ctx context.Context, ticketID int64) error {
	if ticketID == 0 {
		return nil
	}
	cleanupCtx, cancel := coordinatorCompensationContext(ctx)
	defer cancel()
	return s.tickets.DeleteTicket(cleanupCtx, ticketID)
}

func ticketTitle(task ExternalTask) string {
	title := strings.TrimSpace(task.Title)
	if title == "" {
		return task.ExternalID
	}
	return title
}

func ptr[T any](v T) *T { return &v }
