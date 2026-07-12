package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/anthropics/agentsmesh/marketplace/internal/service"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type applyRow struct {
	MarketplaceID        int64
	InstallationID       string
	OperationID          string
	ListingVersionID     int64
	TargetOrganizationID int64
	ActorUserID          int64
	QuotaAccountID       string
	Status               string
	IdempotencyKey       string
	Plan                 []byte
	Configuration        []byte
}

type storedPlan struct {
	PlanID               string          `json:"plan_id"`
	PlanDigest           string          `json:"plan_digest"`
	ExpiresAt            time.Time       `json:"expires_at"`
	EstimatedCredits     int64           `json:"estimated_credits_micro"`
	PlatformResourceType string          `json:"platform_resource_type"`
	RuntimeSnapshot      json.RawMessage `json:"runtime_snapshot"`
}

func (r *InstallationRepository) BeginApply(
	ctx context.Context,
	command service.ApplyInstallationCommand,
) (service.ApplyExecution, bool, error) {
	var execution service.ApplyExecution
	var existing bool
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		row, err := lockApplyRow(tx, command.OperationID)
		if err != nil {
			return err
		}
		if row.ActorUserID != command.ActorUserID {
			return service.ErrOperationNotFound
		}
		if row.Status != "planned" {
			if row.IdempotencyKey != command.IdempotencyKey {
				return service.ErrPlanMismatch
			}
			if row.Status == "running" {
				var plan storedPlan
				if json.Unmarshal(row.Plan, &plan) != nil {
					return service.ErrPlanMismatch
				}
				execution = applyExecution(row, plan)
				return nil
			}
			existing = true
			return nil
		}
		var plan storedPlan
		if json.Unmarshal(row.Plan, &plan) != nil ||
			plan.PlanID != command.PlanID ||
			plan.PlanDigest != command.PlanDigest {
			return service.ErrPlanMismatch
		}
		if !time.Now().UTC().Before(plan.ExpiresAt) {
			return service.ErrPlanExpired
		}
		if err := reserveInstallationQuota(tx, row, command, plan.EstimatedCredits); err != nil {
			return err
		}
		if err := tx.Exec(`
UPDATE marketplace.marketplace_installation_operations
SET idempotency_key = ?::uuid, status = 'running', stage = 'runtime',
  started_at = NOW()
WHERE id = ?::uuid AND status = 'planned'
`, command.IdempotencyKey, command.OperationID).Error; err != nil {
			return err
		}
		if err := tx.Exec(`
UPDATE marketplace.marketplace_installations
SET status = 'installing', updated_at = NOW() WHERE id = ?::uuid
`, row.InstallationID).Error; err != nil {
			return err
		}
		execution = applyExecution(row, plan)
		return nil
	})
	return execution, existing, err
}

func applyExecution(row applyRow, plan storedPlan) service.ApplyExecution {
	return service.ApplyExecution{
		InstallationID: row.InstallationID, OperationID: row.OperationID,
		ListingVersionID:     row.ListingVersionID,
		TargetOrganizationID: row.TargetOrganizationID,
		PlatformResourceType: plan.PlatformResourceType,
		RuntimeSnapshot:      plan.RuntimeSnapshot,
		ActorUserID:          row.ActorUserID,
		Configuration:        row.Configuration,
		ReservedCredits:      plan.EstimatedCredits,
	}
}

func lockApplyRow(tx *gorm.DB, operationID string) (applyRow, error) {
	var row applyRow
	result := tx.Raw(`
SELECT i.marketplace_id, i.id::text AS installation_id,
  o.id::text AS operation_id, i.listing_version_id,
  i.target_platform_org_id AS target_organization_id,
  i.installed_by_platform_user_id AS actor_user_id,
  i.quota_account_id::text, o.status, o.idempotency_key::text,
  o.plan, i.config_snapshot AS configuration
FROM marketplace.marketplace_installation_operations o
JOIN marketplace.marketplace_installations i ON i.id = o.installation_id
WHERE o.id = ?::uuid FOR UPDATE OF o, i
`, operationID).Scan(&row)
	if result.Error != nil {
		return applyRow{}, result.Error
	}
	if row.OperationID == "" {
		return applyRow{}, service.ErrOperationNotFound
	}
	return row, nil
}

func reserveInstallationQuota(
	tx *gorm.DB,
	row applyRow,
	command service.ApplyInstallationCommand,
	credits int64,
) error {
	if credits <= 0 {
		return service.ErrInvalidInstallationRequest
	}
	var account struct{ ID string }
	if err := tx.Raw(`
SELECT id::text
FROM marketplace.marketplace_quota_accounts
WHERE id = ?::uuid AND status = 'active'
  AND NOW() >= period_start AND NOW() < period_end
FOR UPDATE
`, row.QuotaAccountID).Scan(&account).Error; err != nil {
		return err
	}
	if account.ID == "" {
		return service.ErrQuotaAccountNotFound
	}
	var balance struct{ Available int64 }
	if err := tx.Raw(`
SELECT COALESCE(SUM(available_delta) * 1000000, 0)::bigint AS available
FROM marketplace.marketplace_quota_ledger_entries
WHERE quota_account_id = ?::uuid
`, row.QuotaAccountID).Scan(&balance).Error; err != nil {
		return err
	}
	if balance.Available < credits {
		return service.ErrQuotaInsufficient
	}
	reservationID := uuid.NewString()
	if err := tx.Exec(`
INSERT INTO marketplace.marketplace_quota_reservations
  (id, marketplace_id, quota_account_id, reservation_type, subject_ref,
   idempotency_key, reserved_credits, status, expires_at)
VALUES (?::uuid, ?, ?::uuid, 'installation', ?, ?::uuid,
  (?::numeric / 1000000), 'held', NOW() + INTERVAL '15 minutes')
`, reservationID, row.MarketplaceID, row.QuotaAccountID, row.OperationID,
		command.IdempotencyKey, credits).Error; err != nil {
		return err
	}
	return tx.Exec(`
INSERT INTO marketplace.marketplace_quota_ledger_entries
  (id, marketplace_id, quota_account_id, entry_type, available_delta,
   reserved_delta, reservation_id, operation_id, reason,
   created_by_platform_user_id)
VALUES (gen_random_uuid(), ?, ?::uuid, 'reserve',
  -(?::numeric / 1000000), (?::numeric / 1000000), ?::uuid, ?::uuid,
  'installation_apply', ?)
`, row.MarketplaceID, row.QuotaAccountID, credits, credits,
		reservationID, row.OperationID, command.ActorUserID).Error
}
