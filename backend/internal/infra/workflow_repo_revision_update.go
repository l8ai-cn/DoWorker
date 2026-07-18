package infra

import (
	"context"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/workflow"
)

func (r *workflowRepo) UpdateForResourceRevision(
	ctx context.Context,
	id int64,
	resourceRevision int64,
	updates map[string]interface{},
) (bool, error) {
	updates["updated_at"] = time.Now()
	result := r.db.WithContext(ctx).
		Model(&workflow.Workflow{}).
		Where(
			"id = ? AND orchestration_resource_revision = ?",
			id,
			resourceRevision,
		).
		Updates(updates)
	return result.RowsAffected == 1, result.Error
}
