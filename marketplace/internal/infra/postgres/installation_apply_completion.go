package postgres

import (
	"context"
	"encoding/json"

	"github.com/l8ai-cn/agentcloud/marketplace/internal/service"
	"gorm.io/gorm"
)

type reservationRow struct {
	MarketplaceID   int64
	QuotaAccountID  string
	ReservationID   string
	ReservedCredits string
	Status          string
}

func (r *InstallationRepository) CompleteApply(
	ctx context.Context,
	execution service.ApplyExecution,
	runtimeResult service.RuntimeInstallResult,
) (service.ApplyResult, error) {
	result := service.ApplyResult{
		InstallationID: execution.InstallationID,
		OperationID:    execution.OperationID,
		Status:         service.ApplySucceeded,
		Stage:          "settle",
		RuntimeRef:     runtimeResult.RuntimeRef,
	}
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		reservation, err := lockHeldReservation(tx, execution.OperationID)
		if err != nil {
			return err
		}
		if err := tx.Exec(`
INSERT INTO marketplace.marketplace_quota_ledger_entries
  (id, marketplace_id, quota_account_id, entry_type, reserved_delta,
   consumed_delta, reservation_id, operation_id, reason)
VALUES (gen_random_uuid(), ?, ?::uuid, 'debit', -?::numeric, ?::numeric,
  ?::uuid, ?::uuid, 'installation_succeeded')
`, reservation.MarketplaceID, reservation.QuotaAccountID,
			reservation.ReservedCredits, reservation.ReservedCredits,
			reservation.ReservationID, execution.OperationID).Error; err != nil {
			return err
		}
		if err := tx.Exec(`
UPDATE marketplace.marketplace_quota_reservations
SET status = 'settled', updated_at = NOW()
WHERE id = ?::uuid AND status = 'held'
`, reservation.ReservationID).Error; err != nil {
			return err
		}
		payload, err := json.Marshal(map[string]any{
			"runtime_ref": runtimeResult.RuntimeRef,
			"result":      runtimeResult.Result,
		})
		if err != nil {
			return err
		}
		if err := tx.Exec(`
UPDATE marketplace.marketplace_installation_operations
SET status = 'succeeded', stage = 'settle', result = ?::jsonb,
  completed_at = NOW()
WHERE id = ?::uuid AND status = 'running'
`, string(payload), execution.OperationID).Error; err != nil {
			return err
		}
		if err := tx.Exec(`
UPDATE marketplace.marketplace_installations
SET status = 'active', runtime_ref = ?, last_verified_at = NOW(),
  updated_at = NOW()
WHERE id = ?::uuid AND status = 'installing'
`, runtimeResult.RuntimeRef, execution.InstallationID).Error; err != nil {
			return err
		}
		return writeAudit(tx, reservation.MarketplaceID, 0,
			"installation.succeeded", "installation",
			execution.InstallationID, string(payload))
	})
	return result, err
}

func lockHeldReservation(tx *gorm.DB, operationID string) (reservationRow, error) {
	var row reservationRow
	result := tx.Raw(`
SELECT qr.marketplace_id, qr.quota_account_id::text,
  qr.id::text AS reservation_id, qr.reserved_credits::text,
  qr.status
FROM marketplace.marketplace_quota_reservations qr
WHERE qr.subject_ref = ? AND qr.reservation_type = 'installation'
FOR UPDATE
`, operationID).Scan(&row)
	if result.Error != nil {
		return reservationRow{}, result.Error
	}
	if row.ReservationID == "" || row.Status != "held" {
		return reservationRow{}, service.ErrPlanMismatch
	}
	return row, nil
}
