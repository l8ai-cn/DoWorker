package postgres

import (
	"context"

	"github.com/l8ai-cn/agentcloud/marketplace/internal/service"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (r *InstallationRepository) resolveOrCreateOrganizationQuotaAccount(
	ctx context.Context,
	marketplaceID int64,
	quotaPlanID int64,
	organizationID int64,
) (string, error) {
	var accountID string
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var created struct {
			ID string
		}
		if err := tx.Raw(`
INSERT INTO marketplace.marketplace_quota_accounts
  (id, marketplace_id, subject_type, subject_ref, quota_plan_id, status,
   period_start, period_end)
SELECT ?::uuid, qp.marketplace_id, 'organization', ?, qp.id, 'active', NOW(),
  CASE WHEN qp.period = 'monthly'
    THEN NOW() + INTERVAL '1 month'
    ELSE NOW() + INTERVAL '100 years'
  END
FROM marketplace.marketplace_quota_plans qp
WHERE qp.marketplace_id = ? AND qp.id = ? AND qp.status = 'active'
  AND qp.charge_scope = 'organization'
ON CONFLICT (marketplace_id, subject_type, subject_ref, quota_plan_id)
DO NOTHING
RETURNING id::text
`, uuid.NewString(), organizationID, marketplaceID, quotaPlanID).Scan(&created).Error; err != nil {
			return err
		}
		var account struct {
			ID string
		}
		if err := tx.Raw(`
SELECT qa.id::text
FROM marketplace.marketplace_quota_accounts qa
WHERE qa.marketplace_id = ? AND qa.subject_type = 'organization'
  AND qa.subject_ref = ? AND qa.quota_plan_id = ?
  AND qa.status = 'active'
  AND NOW() >= qa.period_start AND NOW() < qa.period_end
LIMIT 1
`, marketplaceID, organizationID, quotaPlanID).Scan(&account).Error; err != nil {
			return err
		}
		if account.ID == "" {
			return service.ErrQuotaAccountNotFound
		}
		if created.ID != "" {
			if err := tx.Exec(`
INSERT INTO marketplace.marketplace_quota_ledger_entries
  (id, marketplace_id, quota_account_id, entry_type, available_delta,
   period_start, reason)
SELECT gen_random_uuid(), qa.marketplace_id, qa.id, 'grant',
  qp.grant_credits, qa.period_start, 'automatic_organization_grant'
FROM marketplace.marketplace_quota_accounts qa
JOIN marketplace.marketplace_quota_plans qp ON qp.id = qa.quota_plan_id
WHERE qa.id = ?::uuid
`, created.ID).Error; err != nil {
				return err
			}
		}
		accountID = account.ID
		return nil
	})
	return accountID, err
}
