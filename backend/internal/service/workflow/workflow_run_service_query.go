package workflow

import (
	"context"
	"log/slog"

	workflowDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/workflow"
)

func (s *WorkflowRunService) GetTimedOutRuns(ctx context.Context, orgIDs []int64) ([]*workflowDomain.WorkflowRun, error) {
	return s.repo.GetTimedOutRuns(ctx, orgIDs)
}

// GetOrphanPendingRuns returns pending runs with no pod_key stuck for > 5 minutes.
func (s *WorkflowRunService) GetOrphanPendingRuns(ctx context.Context, orgIDs []int64) ([]*workflowDomain.WorkflowRun, error) {
	return s.repo.GetOrphanPendingRuns(ctx, orgIDs)
}

func (s *WorkflowRunService) GetIdleWorkflowPods(ctx context.Context, orgIDs []int64) ([]*workflowDomain.WorkflowRun, error) {
	return s.repo.GetIdleWorkflowPods(ctx, orgIDs)
}

// ComputeLoopStats computes run statistics from Pod status (SSOT).
func (s *WorkflowRunService) ComputeLoopStats(ctx context.Context, workflowID int64) (total int, successful int, failed int, err error) {
	return s.repo.ComputeLoopStats(ctx, workflowID)
}

func (s *WorkflowRunService) GetLatestPodKey(ctx context.Context, workflowID int64) *string {
	return s.repo.GetLatestPodKey(ctx, workflowID)
}

func (s *WorkflowRunService) CountActiveRunsByWorkflowIDs(ctx context.Context, workflowIDs []int64) (map[int64]int64, error) {
	return s.repo.CountActiveRunsByWorkflowIDs(ctx, workflowIDs)
}

func (s *WorkflowRunService) GetAvgDuration(ctx context.Context, workflowID int64) (*float64, error) {
	return s.repo.GetAvgDuration(ctx, workflowID)
}

func (s *WorkflowRunService) DeleteOldFinishedRuns(ctx context.Context, workflowID int64, keep int) (int64, error) {
	deleted, err := s.repo.DeleteOldFinishedRuns(ctx, workflowID, keep)
	if err != nil {
		slog.ErrorContext(ctx, "failed to delete old finished runs", "workflow_id", workflowID, "keep", keep, "error", err)
		return 0, err
	}
	if deleted > 0 {
		slog.InfoContext(ctx, "old finished runs deleted", "workflow_id", workflowID, "deleted", deleted, "keep", keep)
	}
	return deleted, nil
}

func (s *WorkflowRunService) GetAutopilotPhase(ctx context.Context, autopilotKey string) string {
	phases, err := s.repo.BatchGetAutopilotPhases(ctx, []string{autopilotKey})
	if err != nil || phases == nil {
		return ""
	}
	return phases[autopilotKey]
}
