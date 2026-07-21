package infra

import (
	"context"
	"time"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/workflow"
	"gorm.io/gorm"
)

func (r *workflowRepo) IncrementRunStats(
	ctx context.Context,
	workflowID int64,
	status string,
	lastRunAt time.Time,
) error {
	updates := map[string]interface{}{
		"total_runs":  gorm.Expr("total_runs + 1"),
		"last_run_at": lastRunAt,
		"updated_at":  time.Now(),
	}
	switch status {
	case workflow.RunStatusCompleted:
		updates["successful_runs"] = gorm.Expr("successful_runs + 1")
	case workflow.RunStatusFailed,
		workflow.RunStatusTimeout,
		workflow.RunStatusCancelled:
		updates["failed_runs"] = gorm.Expr("failed_runs + 1")
	}
	return r.db.WithContext(ctx).
		Model(&workflow.Workflow{}).
		Where("id = ?", workflowID).
		Updates(updates).Error
}
