package workflow

import "context"

type WorkflowRunRepository interface {
	Create(ctx context.Context, run *WorkflowRun) error
	GetByID(ctx context.Context, id int64) (*WorkflowRun, error)
	List(ctx context.Context, filter *WorkflowRunListFilter) ([]*WorkflowRun, int64, error)
	Update(ctx context.Context, runID int64, updates map[string]interface{}) error
	BindPod(
		ctx context.Context,
		runID int64,
		podKey string,
		autopilotKey string,
	) (bool, error)
	GetMaxRunNumber(ctx context.Context, workflowID int64) (int, error)
	GetByAutopilotKey(ctx context.Context, autopilotKey string) (*WorkflowRun, error)

	// TriggerRunAtomic atomically creates a workflow run within a FOR UPDATE transaction.
	// Handles concurrency check (SSOT via Pod JOIN), run number generation, and record creation.
	TriggerRunAtomic(ctx context.Context, params *TriggerRunAtomicParams) (*TriggerRunAtomicResult, error)

	// FinishRun atomically marks a run as finished with optimistic locking.
	// Uses WHERE finished_at IS NULL to prevent double-processing from concurrent events.
	// Returns true if the row was updated (caller should proceed), false if already finished.
	FinishRun(ctx context.Context, runID int64, updates map[string]interface{}) (bool, error)

	// SSOT: cross-table queries (JOIN with pods/autopilot_controllers)
	CountActiveRuns(ctx context.Context, workflowID int64) (int64, error)
	GetActiveRunByPodKey(ctx context.Context, podKey string) (*WorkflowRun, error)
	GetTimedOutRuns(ctx context.Context, orgIDs []int64) ([]*WorkflowRun, error)
	// GetOrphanPendingRuns returns pending runs with no pod_key stuck for > 5 minutes.
	GetOrphanPendingRuns(ctx context.Context, orgIDs []int64) ([]*WorkflowRun, error)
	ComputeLoopStats(ctx context.Context, workflowID int64) (total, successful, failed int, err error)
	GetLatestPodKey(ctx context.Context, workflowID int64) *string

	// SSOT: batch status resolution helpers
	BatchGetPodStatuses(ctx context.Context, podKeys []string) ([]PodStatusInfo, error)
	BatchGetAutopilotPhases(ctx context.Context, autopilotKeys []string) (map[string]string, error)

	CountActiveRunsByWorkflowIDs(ctx context.Context, workflowIDs []int64) (map[int64]int64, error)

	GetAvgDuration(ctx context.Context, workflowID int64) (*float64, error)

	// DeleteOldFinishedRuns deletes finished runs exceeding the retention limit.
	// Keeps the most recent `keep` finished runs, deletes the rest.
	// Returns the number of rows deleted.
	DeleteOldFinishedRuns(ctx context.Context, workflowID int64, keep int) (int64, error)

	GetIdleWorkflowPods(ctx context.Context, orgIDs []int64) ([]*WorkflowRun, error)
}
