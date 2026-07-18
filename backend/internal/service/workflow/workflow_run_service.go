package workflow

import (
	"context"
	"errors"
	"log/slog"

	workflowDomain "github.com/anthropics/agentsmesh/backend/internal/domain/workflow"
)

var (
	ErrRunNotFound  = errors.New("workflow run not found")
	ErrRunStartLost = errors.New("workflow run is no longer startable")
)

// WorkflowRunService — read methods resolve effective status from Pod (SSOT).
// run.Status is authoritative only when pod_key is NULL.
type WorkflowRunService struct {
	repo workflowDomain.WorkflowRunRepository
}

func NewWorkflowRunService(repo workflowDomain.WorkflowRunRepository) *WorkflowRunService {
	return &WorkflowRunService{repo: repo}
}

type ListWorkflowRunsFilter struct {
	WorkflowID int64
	Status     string
	Limit      int
	Offset     int
}

func (s *WorkflowRunService) Create(ctx context.Context, run *workflowDomain.WorkflowRun) error {
	if err := s.repo.Create(ctx, run); err != nil {
		slog.ErrorContext(ctx, "failed to create workflow run", "workflow_id", run.WorkflowID, "run_number", run.RunNumber, "error", err)
		return err
	}
	slog.InfoContext(ctx, "workflow run created", "run_id", run.ID, "workflow_id", run.WorkflowID, "run_number", run.RunNumber)
	return nil
}

func (s *WorkflowRunService) GetByID(ctx context.Context, id int64) (*workflowDomain.WorkflowRun, error) {
	run, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, workflowDomain.ErrNotFound) {
			return nil, ErrRunNotFound
		}
		return nil, err
	}
	s.resolveRunStatus(ctx, run)
	return run, nil
}

func (s *WorkflowRunService) ListWorkflowRuns(ctx context.Context, filter *ListWorkflowRunsFilter) ([]*workflowDomain.WorkflowRun, int64, error) {
	runs, total, err := s.repo.List(ctx, &workflowDomain.WorkflowRunListFilter{
		WorkflowID: filter.WorkflowID,
		Status:     filter.Status,
		Limit:      filter.Limit,
		Offset:     filter.Offset,
	})
	if err != nil {
		return nil, 0, err
	}

	s.resolveRunStatuses(ctx, runs)

	if filter.Status != "" {
		filtered := make([]*workflowDomain.WorkflowRun, 0, len(runs))
		for _, run := range runs {
			if run.Status == filter.Status {
				filtered = append(filtered, run)
			}
		}
		removed := int64(len(runs) - len(filtered))
		runs = filtered
		total -= removed
	}

	return runs, total, nil
}

func (s *WorkflowRunService) TriggerRunAtomic(ctx context.Context, params *workflowDomain.TriggerRunAtomicParams) (*workflowDomain.TriggerRunAtomicResult, error) {
	return s.repo.TriggerRunAtomic(ctx, params)
}

func (s *WorkflowRunService) GetNextRunNumber(ctx context.Context, workflowID int64) (int, error) {
	maxNumber, err := s.repo.GetMaxRunNumber(ctx, workflowID)
	if err != nil {
		return 0, err
	}
	return maxNumber + 1, nil
}

func (s *WorkflowRunService) CountActiveRuns(ctx context.Context, workflowID int64) (int64, error) {
	return s.repo.CountActiveRuns(ctx, workflowID)
}

func (s *WorkflowRunService) UpdateStatus(ctx context.Context, runID int64, updates map[string]interface{}) error {
	return s.repo.Update(ctx, runID, updates)
}

func (s *WorkflowRunService) BindPod(
	ctx context.Context,
	runID int64,
	podKey string,
	autopilotKey string,
) (bool, error) {
	return s.repo.BindPod(ctx, runID, podKey, autopilotKey)
}

func (s *WorkflowRunService) FinishRun(ctx context.Context, runID int64, updates map[string]interface{}) (bool, error) {
	updated, err := s.repo.FinishRun(ctx, runID, updates)
	if err != nil {
		slog.ErrorContext(ctx, "failed to finish workflow run", "run_id", runID, "error", err)
		return false, err
	}
	if updated {
		slog.InfoContext(ctx, "workflow run finished", "run_id", runID, "status", updates["status"])
	}
	return updated, nil
}

func (s *WorkflowRunService) GetActiveRunByPodKey(ctx context.Context, podKey string) (*workflowDomain.WorkflowRun, error) {
	run, err := s.repo.GetActiveRunByPodKey(ctx, podKey)
	if err != nil {
		if errors.Is(err, workflowDomain.ErrNotFound) {
			return nil, ErrRunNotFound
		}
		return nil, err
	}
	s.resolveRunStatus(ctx, run)
	return run, nil
}

func (s *WorkflowRunService) FindActiveRunByPodKey(ctx context.Context, podKey string) (*workflowDomain.WorkflowRun, error) {
	run, err := s.repo.GetActiveRunByPodKey(ctx, podKey)
	if err != nil {
		if errors.Is(err, workflowDomain.ErrNotFound) {
			return nil, ErrRunNotFound
		}
		return nil, err
	}
	return run, nil
}

func (s *WorkflowRunService) GetActiveRunByAutopilotKey(ctx context.Context, autopilotKey string) (*workflowDomain.WorkflowRun, error) {
	run, err := s.repo.GetByAutopilotKey(ctx, autopilotKey)
	if err != nil {
		if errors.Is(err, workflowDomain.ErrNotFound) {
			return nil, ErrRunNotFound
		}
		return nil, err
	}
	s.resolveRunStatus(ctx, run)
	return run, nil
}

func (s *WorkflowRunService) FindActiveRunByAutopilotKey(ctx context.Context, autopilotKey string) (*workflowDomain.WorkflowRun, error) {
	run, err := s.repo.GetByAutopilotKey(ctx, autopilotKey)
	if err != nil {
		if errors.Is(err, workflowDomain.ErrNotFound) {
			return nil, ErrRunNotFound
		}
		return nil, err
	}
	return run, nil
}
