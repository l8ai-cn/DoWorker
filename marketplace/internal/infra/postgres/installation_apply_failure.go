package postgres

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/anthropics/agentsmesh/marketplace/internal/service"
	"gorm.io/gorm"
)

func (r *InstallationRepository) FailApply(
	ctx context.Context,
	execution service.ApplyExecution,
	cause error,
) (service.ApplyResult, error) {
	result := service.ApplyResult{
		InstallationID: execution.InstallationID,
		OperationID:    execution.OperationID,
		Status:         service.ApplyFailed,
		Stage:          "runtime",
		ErrorCode:      "RUNTIME_INSTALL_FAILED",
		ErrorMessage:   "运行时安装失败",
	}
	switch {
	case errors.Is(cause, service.ErrTargetOrganizationForbidden):
		result.ErrorCode = "TARGET_ORGANIZATION_FORBIDDEN"
		result.ErrorMessage = "你已无权在这个组织中启用应用"
	case errors.Is(cause, service.ErrRuntimeAuthorizationFailed):
		result.ErrorCode = "RUNTIME_AUTHORIZATION_FAILED"
		result.ErrorMessage = "组织权限校验暂时失败"
	}
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		reservation, err := lockHeldReservation(tx, execution.OperationID)
		if err != nil {
			return err
		}
		if err := tx.Exec(`
INSERT INTO marketplace.marketplace_quota_ledger_entries
  (id, marketplace_id, quota_account_id, entry_type, available_delta,
   reserved_delta, reservation_id, operation_id, reason)
VALUES (gen_random_uuid(), ?, ?::uuid, 'release', ?::numeric, -?::numeric,
  ?::uuid, ?::uuid, 'runtime_install_failed')
`, reservation.MarketplaceID, reservation.QuotaAccountID,
			reservation.ReservedCredits, reservation.ReservedCredits,
			reservation.ReservationID, execution.OperationID).Error; err != nil {
			return err
		}
		if err := tx.Exec(`
UPDATE marketplace.marketplace_quota_reservations
SET status = 'released', updated_at = NOW()
WHERE id = ?::uuid AND status = 'held'
`, reservation.ReservationID).Error; err != nil {
			return err
		}
		if err := tx.Exec(`
UPDATE marketplace.marketplace_installation_operations
SET status = 'failed', stage = 'runtime', error_code = ?,
  error_message = ?, completed_at = NOW()
WHERE id = ?::uuid AND status = 'running'
`, result.ErrorCode, result.ErrorMessage, execution.OperationID).Error; err != nil {
			return err
		}
		if err := tx.Exec(`
UPDATE marketplace.marketplace_installations
SET status = 'failed', updated_at = NOW()
WHERE id = ?::uuid AND status = 'installing'
`, execution.InstallationID).Error; err != nil {
			return err
		}
		payload, err := json.Marshal(map[string]string{"code": result.ErrorCode})
		if err != nil {
			return err
		}
		return writeAudit(tx, reservation.MarketplaceID, 0,
			"installation.failed", "installation",
			execution.InstallationID, string(payload))
	})
	return result, err
}
