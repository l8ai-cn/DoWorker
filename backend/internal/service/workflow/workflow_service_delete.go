package workflow

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	workflowDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/workflow"
)

var ErrHasActiveRuns = errors.New("workflow has active runs")

// Delete is atomic — returns ErrHasActiveRuns when active runs exist (no orphan rows).
func (s *WorkflowService) Delete(ctx context.Context, orgID int64, slug string) error {
	workflow, err := s.GetBySlug(ctx, orgID, slug)
	if err != nil {
		return err
	}
	if workflow.IsResourceManaged() {
		return ErrWorkflowManagedByResourceApply
	}

	affected, err := s.repo.Delete(ctx, orgID, slug)
	if err != nil {
		if errors.Is(err, workflowDomain.ErrHasActiveRuns) {
			return ErrHasActiveRuns
		}
		slog.ErrorContext(ctx, "failed to delete workflow", "slug", slug, "org_id", orgID, "error", err)
		return err
	}
	if affected == 0 {
		return ErrWorkflowNotFound
	}
	slog.InfoContext(ctx, "workflow deleted", "slug", slug, "org_id", orgID)
	return nil
}

var validStatuses = map[string]bool{
	workflowDomain.StatusEnabled:  true,
	workflowDomain.StatusDisabled: true,
}

func (s *WorkflowService) SetStatus(ctx context.Context, orgID int64, slug string, status string) (*workflowDomain.Workflow, error) {
	if !validStatuses[status] {
		return nil, fmt.Errorf("%w: status must be 'enabled' or 'disabled'", ErrInvalidEnumValue)
	}

	workflow, err := s.GetBySlug(ctx, orgID, slug)
	if err != nil {
		return nil, err
	}

	updates := map[string]interface{}{
		"status": status,
	}

	if status == workflowDomain.StatusEnabled && workflow.HasCron() {
		schedule, err := cronParser.Parse(*workflow.CronExpression)
		if err == nil {
			next := schedule.Next(time.Now())
			updates["next_run_at"] = next
		}
	}

	if err := s.repo.Update(ctx, workflow.ID, updates); err != nil {
		slog.ErrorContext(ctx, "failed to set workflow status", "workflow_id", workflow.ID, "slug", slug, "status", status, "error", err)
		return nil, err
	}

	slog.InfoContext(ctx, "workflow status changed", "workflow_id", workflow.ID, "slug", slug, "org_id", orgID, "status", status)
	return s.GetBySlug(ctx, orgID, slug)
}
