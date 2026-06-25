package coordinator

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	coordinatordom "github.com/anthropics/agentsmesh/backend/internal/domain/coordinator"
	ticketDomain "github.com/anthropics/agentsmesh/backend/internal/domain/ticket"
)

// HandlePodTerminated is the eventbus consumer entrypoint: when a coordinator
// pod reaches a terminal status, post feedback to the external platform and
// advance the execution + ticket. Pod is the SSOT for status; this never runs
// for non-coordinator pods (no execution row → early return).
func (s *Service) HandlePodTerminated(ctx context.Context, podKey, podStatus string) {
	execution, err := s.store.GetExecutionByPodKey(ctx, podKey)
	if err != nil {
		return
	}
	if coordinatordom.IsTerminalStatus(execution.Status) {
		return
	}

	final := mapPodStatus(podStatus)
	now := time.Now()
	updates := map[string]any{
		"status":      final,
		"stage":       "finished",
		"finished_at": now,
	}

	project, err := s.store.GetProject(ctx, execution.OrganizationID, execution.ProjectID)
	if err != nil {
		updates["error"] = fmt.Sprintf("load project: %v", err)
		_ = s.store.UpdateExecution(ctx, execution.ID, updates)
		return
	}

	if err := s.postFeedback(ctx, project, execution, final); err != nil {
		updates["feedback_status"] = coordinatordom.FeedbackStatusFailed
		updates["error"] = err.Error()
		if final == coordinatordom.ExecutionStatusSucceeded {
			updates["status"] = coordinatordom.ExecutionStatusFeedbackFailed
		}
	} else {
		updates["feedback_status"] = coordinatordom.FeedbackStatusPosted
	}

	_ = s.store.UpdateExecution(ctx, execution.ID, updates)
	s.advanceTicket(ctx, execution.TicketID, final)
}

func (s *Service) postFeedback(ctx context.Context, project *coordinatordom.Project, execution *coordinatordom.Execution, final string) error {
	platform, repo, err := s.platform.For(ctx, project)
	if err != nil {
		return fmt.Errorf("resolve platform: %w", err)
	}
	// CNB keys off the issue number; Linear keys off the opaque external id and
	// leaves Number zero. Number is best-effort, not required.
	number, _ := parseIssueNumber(execution.ExternalID)
	body := feedbackBody(final, execution)
	task := ExternalTask{ExternalID: execution.ExternalID, Number: number}
	return platform.PostFeedback(ctx, repo, task, body)
}

func (s *Service) advanceTicket(ctx context.Context, ticketID int64, final string) {
	status := ticketDomain.TicketStatusInReview
	if final != coordinatordom.ExecutionStatusSucceeded {
		status = ticketDomain.TicketStatusTodo
	}
	if err := s.tickets.UpdateStatus(ctx, ticketID, status); err != nil {
		s.logger.Warn("failed to advance ticket status", "ticket_id", ticketID, "error", err)
	}
}

func mapPodStatus(podStatus string) string {
	switch podStatus {
	case "completed":
		return coordinatordom.ExecutionStatusSucceeded
	case "cancelled", "terminated":
		return coordinatordom.ExecutionStatusCancelled
	default:
		return coordinatordom.ExecutionStatusFailed
	}
}

func feedbackBody(final string, execution *coordinatordom.Execution) string {
	verb := "completed successfully"
	switch final {
	case coordinatordom.ExecutionStatusFailed:
		verb = "failed"
	case coordinatordom.ExecutionStatusCancelled:
		verb = "was cancelled"
	}
	body := fmt.Sprintf("AgentsMesh Coordinator: do-agent run %s.", verb)
	if strings.TrimSpace(execution.Summary) != "" {
		body += "\n\n" + execution.Summary
	}
	return body
}

func parseIssueNumber(externalID string) (int, bool) {
	_, rest, ok := strings.Cut(externalID, ":")
	if !ok {
		return 0, false
	}
	n, err := strconv.Atoi(strings.TrimSpace(rest))
	if err != nil {
		return 0, false
	}
	return n, true
}
