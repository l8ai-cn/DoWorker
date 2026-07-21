package postgres

import (
	"context"

	"github.com/l8ai-cn/agentcloud/marketplace/internal/service"
	"gorm.io/gorm"
)

func (r *InstallationRepository) CreateDirectPlan(
	ctx context.Context,
	record service.InstallationPlanRecord,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var entitlement struct{ ID string }
		if err := tx.Raw(`
INSERT INTO marketplace.marketplace_entitlements
  (id, marketplace_id, listing_id, subject_type, subject_platform_id,
   target_platform_org_id, status, source, starts_at,
   granted_by_platform_user_id)
VALUES (?::uuid, ?, ?, 'user', ?, ?, 'active', 'direct', NOW(), ?)
ON CONFLICT (
  marketplace_id, listing_id, subject_type, subject_platform_id, target_platform_org_id
) WHERE source = 'direct' AND status = 'active'
DO UPDATE SET updated_at = NOW()
RETURNING id::text
`, record.EntitlementID, record.MarketplaceID, record.ListingID,
			record.ActorUserID, record.TargetOrganizationID,
			record.ActorUserID).Scan(&entitlement).Error; err != nil {
			return err
		}
		if err := tx.Exec(`
INSERT INTO marketplace.marketplace_installations
  (id, marketplace_id, listing_id, listing_version_id, entitlement_id,
   target_platform_org_id, quota_charge_scope, quota_account_id,
   installed_by_platform_user_id, status, config_snapshot, plan_digest)
VALUES (?::uuid, ?, ?, ?, ?::uuid, ?, 'organization', ?::uuid, ?,
  'planning', ?::jsonb, ?)
`, record.InstallationID, record.MarketplaceID, record.ListingID,
			record.ListingVersionID, entitlement.ID, record.TargetOrganizationID,
			record.QuotaAccountID, record.ActorUserID,
			string(record.Configuration), record.PlanDigest).Error; err != nil {
			return err
		}
		if err := tx.Exec(`
INSERT INTO marketplace.marketplace_installation_operations
  (id, marketplace_id, installation_id, operation_type, idempotency_key,
   status, stage, plan)
VALUES (?::uuid, ?, ?::uuid, 'install', ?::uuid, 'planned',
  'entitlement', ?::jsonb)
`, record.OperationID, record.MarketplaceID, record.InstallationID,
			record.OperationID, string(record.Plan)).Error; err != nil {
			return err
		}
		if err := tx.Exec(`
UPDATE marketplace.marketplace_installations
SET current_operation_id = ?::uuid, updated_at = NOW()
WHERE id = ?::uuid
`, record.OperationID, record.InstallationID).Error; err != nil {
			return err
		}
		return writeAudit(tx, record.MarketplaceID, record.ActorUserID,
			"installation.plan_created", "installation", record.InstallationID,
			string(record.Plan))
	})
}

func writeAudit(
	tx *gorm.DB,
	marketplaceID int64,
	actorUserID int64,
	action string,
	targetType string,
	targetRef string,
	newData string,
) error {
	return tx.Exec(`
INSERT INTO marketplace.marketplace_audit_events
  (id, marketplace_id, actor_platform_user_id, action, target_type,
   target_ref, new_data)
VALUES (gen_random_uuid(), ?, NULLIF(?, 0), ?, ?, ?, ?::jsonb)
`, marketplaceID, actorUserID, action, targetType, targetRef, newData).Error
}
