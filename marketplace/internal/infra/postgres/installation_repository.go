package postgres

import (
	"context"
	"encoding/json"

	"github.com/l8ai-cn/agentcloud/marketplace/internal/service"
	"gorm.io/gorm"
)

type InstallationRepository struct {
	db *gorm.DB
}

func NewInstallationRepository(db *gorm.DB) *InstallationRepository {
	return &InstallationRepository{db: db}
}

func (r *InstallationRepository) ResolveInstallSource(
	ctx context.Context,
	marketSlug string,
	listingSlug string,
	listingVersionID int64,
	targetOrganizationID int64,
) (service.InstallSource, error) {
	var source service.InstallSource
	result := r.db.WithContext(ctx).Raw(`
SELECT m.id AS marketplace_id, l.id AS listing_id, lv.id AS listing_version_id,
  l.access_mode, civ.content_digest, civ.permissions, civ.manifest,
  ci.platform_resource_type, COALESCE(ci.platform_resource_id, 0) AS platform_resource_id,
  COALESCE((civ.manifest->'source_release'->>'release_id')::bigint, 0)
    AS source_release_id,
  civ.manifest->'runtime_snapshot' AS runtime_snapshot,
  qp.id AS quota_plan_id,
  qp.charge_scope AS quota_charge_scope,
  CASE
    WHEN civ.manifest->>'installation_credits'
      ~ '^([0-9]+|[0-9]+\.[0-9]{1,6})$'
    THEN ((civ.manifest->>'installation_credits')::numeric * 1000000)::bigint
  END AS estimated_credits
FROM marketplace.marketplaces m
JOIN marketplace.marketplace_listings l ON l.marketplace_id = m.id
JOIN marketplace.marketplace_listing_versions lv
  ON lv.id = l.current_version_id AND lv.listing_id = l.id
JOIN marketplace.marketplace_catalog_item_versions civ
  ON civ.id = lv.catalog_item_version_id
JOIN marketplace.marketplace_catalog_items ci ON ci.id = civ.catalog_item_id
JOIN marketplace.marketplace_quota_plans qp
  ON qp.id = lv.quota_plan_id AND qp.marketplace_id = m.id
WHERE m.slug = ? AND m.status = 'published'
  AND l.slug = ? AND l.status = 'published'
  AND lv.id = ? AND lv.review_status = 'approved'
  AND civ.validation_status = 'passed'
  AND qp.status = 'active' AND qp.charge_scope = 'organization'
  AND civ.manifest ? 'installation_credits'
  AND civ.manifest ? 'runtime_snapshot'
  AND ci.resource_type = 'application'
  AND ci.platform_resource_type = 'expert'
LIMIT 1
`, marketSlug, listingSlug, listingVersionID).Scan(&source)
	if result.Error != nil {
		return service.InstallSource{}, result.Error
	}
	if source.ListingID == 0 || source.EstimatedCredits <= 0 {
		return service.InstallSource{}, service.ErrListingNotFound
	}
	accountID, err := r.resolveOrCreateOrganizationQuotaAccount(
		ctx,
		source.MarketplaceID,
		source.QuotaPlanID,
		targetOrganizationID,
	)
	if err != nil {
		return service.InstallSource{}, err
	}
	source.QuotaAccountID = accountID
	source.Permissions = cloneRawJSON(source.Permissions, `[]`)
	source.Manifest = cloneRawJSON(source.Manifest, `{}`)
	return source, nil
}

func cloneRawJSON(value json.RawMessage, empty string) json.RawMessage {
	if len(value) == 0 {
		return json.RawMessage(empty)
	}
	return append(json.RawMessage(nil), value...)
}
