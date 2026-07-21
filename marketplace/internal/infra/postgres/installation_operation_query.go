package postgres

import (
	"context"

	"github.com/l8ai-cn/agentcloud/marketplace/internal/service"
)

func (r *InstallationRepository) GetApplyResult(
	ctx context.Context,
	operationID string,
	actorUserID int64,
) (service.ApplyResult, error) {
	var row struct {
		InstallationID string
		OperationID    string
		Status         service.ApplyStatus
		Stage          string
		RuntimeRef     string
		ErrorCode      string
		ErrorMessage   string
	}
	result := r.db.WithContext(ctx).Raw(`
SELECT i.id::text AS installation_id, o.id::text AS operation_id,
  o.status, o.stage, COALESCE(i.runtime_ref, '') AS runtime_ref,
  COALESCE(o.error_code, '') AS error_code,
  COALESCE(o.error_message, '') AS error_message
FROM marketplace.marketplace_installation_operations o
JOIN marketplace.marketplace_installations i ON i.id = o.installation_id
WHERE o.id = ?::uuid
  AND i.installed_by_platform_user_id = ?
LIMIT 1
`, operationID, actorUserID).Scan(&row)
	if result.Error != nil {
		return service.ApplyResult{}, result.Error
	}
	if row.OperationID == "" {
		return service.ApplyResult{}, service.ErrOperationNotFound
	}
	return service.ApplyResult{
		InstallationID: row.InstallationID,
		OperationID:    row.OperationID,
		Status:         row.Status,
		Stage:          row.Stage,
		RuntimeRef:     row.RuntimeRef,
		ErrorCode:      row.ErrorCode,
		ErrorMessage:   row.ErrorMessage,
	}, nil
}
