package workflow

import (
	"context"
	"errors"
	"log/slog"
	"time"

	workflowDomain "github.com/anthropics/agentsmesh/backend/internal/domain/workflow"
)

type WorkflowService struct {
	repo workflowDomain.WorkflowRepository
}

func NewWorkflowService(repo workflowDomain.WorkflowRepository) *WorkflowService {
	return &WorkflowService{repo: repo}
}

func (s *WorkflowService) GetBySlug(ctx context.Context, orgID int64, slug string) (*workflowDomain.Workflow, error) {
	workflow, err := s.repo.GetBySlug(ctx, orgID, slug)
	if err != nil {
		if errors.Is(err, workflowDomain.ErrNotFound) {
			return nil, ErrWorkflowNotFound
		}
		return nil, err
	}
	return workflow, nil
}

func (s *WorkflowService) GetByID(ctx context.Context, id int64) (*workflowDomain.Workflow, error) {
	workflow, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, workflowDomain.ErrNotFound) {
			return nil, ErrWorkflowNotFound
		}
		return nil, err
	}
	return workflow, nil
}

func (s *WorkflowService) List(ctx context.Context, filter *ListWorkflowsFilter) ([]*workflowDomain.Workflow, int64, error) {
	return s.repo.List(ctx, filter)
}

func (s *WorkflowService) UpdateRunStats(ctx context.Context, workflowID int64, status string, lastRunAt time.Time) error {
	if err := s.repo.IncrementRunStats(ctx, workflowID, status, lastRunAt); err != nil {
		slog.ErrorContext(ctx, "failed to update workflow run stats", "workflow_id", workflowID, "status", status, "error", err)
		return err
	}
	return nil
}

func (s *WorkflowService) UpdateStats(ctx context.Context, workflowID int64, total, successful, failed int) error {
	return s.repo.Update(ctx, workflowID, map[string]interface{}{
		"total_runs":      total,
		"successful_runs": successful,
		"failed_runs":     failed,
	})
}

func (s *WorkflowService) ClearRuntimeState(ctx context.Context, workflowID int64) error {
	if err := s.repo.Update(ctx, workflowID, map[string]interface{}{
		"sandbox_path": nil,
		"last_pod_key": nil,
	}); err != nil {
		slog.ErrorContext(ctx, "failed to clear workflow runtime state", "workflow_id", workflowID, "error", err)
		return err
	}
	slog.InfoContext(ctx, "workflow runtime state cleared", "workflow_id", workflowID)
	return nil
}

func (s *WorkflowService) UpdateRuntimeState(ctx context.Context, workflowID int64, sandboxPath *string, lastPodKey *string) error {
	updates := map[string]interface{}{}
	if sandboxPath != nil {
		updates["sandbox_path"] = *sandboxPath
	}
	if lastPodKey != nil {
		updates["last_pod_key"] = *lastPodKey
	}
	if len(updates) == 0 {
		return nil
	}
	return s.repo.Update(ctx, workflowID, updates)
}

func (s *WorkflowService) UpdateNextRunAt(ctx context.Context, workflowID int64, nextRunAt *time.Time) error {
	return s.repo.Update(ctx, workflowID, map[string]interface{}{
		"next_run_at": nextRunAt,
	})
}

func (s *WorkflowService) GetDueCronWorkflows(ctx context.Context, orgIDs []int64) ([]*workflowDomain.Workflow, error) {
	return s.repo.GetDueCronWorkflows(ctx, orgIDs)
}

// ClaimCronWorkflow atomically claims a cron workflow with SKIP LOCKED and advances next_run_at.
func (s *WorkflowService) ClaimCronWorkflow(ctx context.Context, workflowID int64, nextRunAt *time.Time) (bool, error) {
	return s.repo.ClaimCronWorkflow(ctx, workflowID, nextRunAt)
}

func (s *WorkflowService) FindWorkflowsNeedingNextRun(ctx context.Context, orgIDs []int64) ([]*workflowDomain.Workflow, error) {
	return s.repo.FindWorkflowsNeedingNextRun(ctx, orgIDs)
}
