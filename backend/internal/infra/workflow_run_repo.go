package infra

import (
	"context"
	"time"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/workflow"
	"gorm.io/gorm"
)

type workflowRunRepo struct {
	db *gorm.DB
}

func NewWorkflowRunRepository(db *gorm.DB) workflow.WorkflowRunRepository {
	return &workflowRunRepo{db: db}
}

func (r *workflowRunRepo) Create(ctx context.Context, run *workflow.WorkflowRun) error {
	return r.db.WithContext(ctx).Create(run).Error
}

func (r *workflowRunRepo) GetByID(ctx context.Context, id int64) (*workflow.WorkflowRun, error) {
	var run workflow.WorkflowRun
	if err := r.db.WithContext(ctx).First(&run, id).Error; err != nil {
		if isNotFound(err) {
			return nil, workflow.ErrNotFound
		}
		return nil, err
	}
	return &run, nil
}

func (r *workflowRunRepo) List(ctx context.Context, filter *workflow.WorkflowRunListFilter) ([]*workflow.WorkflowRun, int64, error) {
	query := r.db.WithContext(ctx).Where("workflow_id = ?", filter.WorkflowID)

	if filter.Status != "" {
		query = query.Where(
			"(finished_at IS NOT NULL AND status = ?) OR (finished_at IS NULL)",
			filter.Status,
		)
	}

	var total int64
	if err := query.Model(&workflow.WorkflowRun{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	limit := filter.Limit
	if limit == 0 {
		limit = 20
	}

	var runs []*workflow.WorkflowRun
	if err := query.Order("created_at DESC").
		Limit(limit).
		Offset(filter.Offset).
		Find(&runs).Error; err != nil {
		return nil, 0, err
	}

	return runs, total, nil
}

func (r *workflowRunRepo) Update(ctx context.Context, runID int64, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return r.db.WithContext(ctx).
		Model(&workflow.WorkflowRun{}).
		Where("id = ?", runID).
		Updates(updates).Error
}

func (r *workflowRunRepo) BindPod(
	ctx context.Context,
	runID int64,
	podKey string,
	autopilotKey string,
) (bool, error) {
	updates := map[string]interface{}{
		"pod_key":    podKey,
		"updated_at": time.Now(),
	}
	if autopilotKey != "" {
		updates["autopilot_controller_key"] = autopilotKey
	}
	result := r.db.WithContext(ctx).
		Model(&workflow.WorkflowRun{}).
		Where("id = ? AND finished_at IS NULL AND pod_key IS NULL", runID).
		Updates(updates)
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected == 1, nil
}

// FinishRun atomically marks a run as finished with optimistic locking.
func (r *workflowRunRepo) FinishRun(ctx context.Context, runID int64, updates map[string]interface{}) (bool, error) {
	updates["updated_at"] = time.Now()
	result := r.db.WithContext(ctx).
		Model(&workflow.WorkflowRun{}).
		Where("id = ? AND finished_at IS NULL", runID).
		Updates(updates)
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *workflowRunRepo) GetMaxRunNumber(ctx context.Context, workflowID int64) (int, error) {
	var maxNumber int
	err := r.db.WithContext(ctx).
		Model(&workflow.WorkflowRun{}).
		Where("workflow_id = ?", workflowID).
		Select("COALESCE(MAX(run_number), 0)").
		Scan(&maxNumber).Error
	return maxNumber, err
}

func (r *workflowRunRepo) GetByAutopilotKey(ctx context.Context, autopilotKey string) (*workflow.WorkflowRun, error) {
	var run workflow.WorkflowRun
	if err := r.db.WithContext(ctx).
		Where("autopilot_controller_key = ? AND finished_at IS NULL", autopilotKey).
		First(&run).Error; err != nil {
		if isNotFound(err) {
			return nil, workflow.ErrNotFound
		}
		return nil, err
	}
	return &run, nil
}

func (r *workflowRunRepo) DeleteOldFinishedRuns(ctx context.Context, workflowID int64, keep int) (int64, error) {
	if keep <= 0 {
		return 0, nil
	}

	result := r.db.WithContext(ctx).Exec(`
		DELETE FROM workflow_runs
		WHERE workflow_id = ? AND finished_at IS NOT NULL
		  AND id NOT IN (
		    SELECT id FROM workflow_runs
		    WHERE workflow_id = ? AND finished_at IS NOT NULL
		    ORDER BY id DESC
		    LIMIT ?
		  )
	`, workflowID, workflowID, keep)

	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

var _ workflow.WorkflowRunRepository = (*workflowRunRepo)(nil)
