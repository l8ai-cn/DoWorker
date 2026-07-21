package postgres

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

func ensureExpertMarket(tx *gorm.DB) (expertCatalogReferences, error) {
	var refs expertCatalogReferences
	if err := tx.Raw(`
SELECT id AS marketplace_id, default_quota_plan_id AS quota_plan_id
FROM marketplace.marketplaces
WHERE slug = 'agent-cloud-market' AND status = 'published'
`).Scan(&refs).Error; err != nil {
		return refs, err
	}
	if refs.MarketplaceID == 0 || refs.QuotaPlanID == 0 {
		return refs, fmt.Errorf("published agent-cloud-market with quota plan is required")
	}
	if err := tx.Raw(`
INSERT INTO marketplace.marketplace_spaces
  (marketplace_id, slug, name, summary, description, status, sort_order,
   created_by_platform_user_id, published_at)
VALUES (?, 'expert-applications', '专家应用', '经过审核的可安装专家',
  '面向团队真实工作场景的专家应用。', 'published', 20, 1, NOW())
ON CONFLICT (marketplace_id, slug) DO UPDATE
SET status = 'published', updated_at = NOW()
WHERE marketplace.marketplace_spaces.status <> 'published'
RETURNING id
`, refs.MarketplaceID).Scan(&refs.SpaceID).Error; err != nil {
		return refs, err
	}
	if refs.SpaceID == 0 {
		if err := tx.Raw(`
SELECT id FROM marketplace.marketplace_spaces
WHERE marketplace_id = ? AND slug = 'expert-applications'
`, refs.MarketplaceID).Scan(&refs.SpaceID).Error; err != nil {
			return refs, err
		}
	}
	if refs.SpaceID == 0 {
		return refs, fmt.Errorf("expert applications space is required")
	}
	return refs, nil
}

func ensureExpertPublisher(
	tx *gorm.DB,
	release publishedExpertRelease,
) (int64, error) {
	var row struct{ ID int64 }
	if release.IsOperatorOwned {
		err := tx.Raw(`
SELECT id FROM marketplace.marketplace_publishers
WHERE slug = 'agent-cloud' AND publisher_type = 'platform'
`).Scan(&row).Error
		if err != nil {
			return 0, err
		}
		if row.ID == 0 {
			return 0, fmt.Errorf("agent-cloud platform publisher is required")
		}
		return row.ID, nil
	}
	err := tx.Raw(`
INSERT INTO marketplace.marketplace_publishers
  (slug, publisher_type, platform_org_id, display_name, summary,
   verification_status)
VALUES (?, 'organization', ?, ?, '平台审核的专家发布方', 'unverified')
ON CONFLICT (slug) DO UPDATE
SET display_name = EXCLUDED.display_name, updated_at = NOW()
WHERE marketplace.marketplace_publishers.display_name IS DISTINCT FROM EXCLUDED.display_name
RETURNING id
`, "publisher-"+release.PublisherSlug, release.PublisherOrganizationID,
		release.PublisherName).Scan(&row).Error
	if err == nil && row.ID == 0 {
		err = tx.Raw(`
SELECT id FROM marketplace.marketplace_publishers
WHERE slug = ?
`, "publisher-"+release.PublisherSlug).Scan(&row).Error
	}
	if err != nil {
		return 0, err
	}
	if row.ID == 0 {
		return 0, fmt.Errorf("expert publisher %s is required", release.PublisherSlug)
	}
	return row.ID, nil
}

func ensureExpertCatalogItem(
	tx *gorm.DB,
	publisherID int64,
	release publishedExpertRelease,
) (int64, error) {
	var row struct{ ID int64 }
	err := tx.Raw(`
INSERT INTO marketplace.marketplace_catalog_items
  (publisher_id, slug, resource_type, name, summary, platform_resource_type,
   platform_resource_id, status, created_by_platform_user_id)
VALUES (?, ?, 'application', ?, ?, 'expert', ?, 'active', ?)
ON CONFLICT (platform_resource_type, platform_resource_id) DO UPDATE
SET name = EXCLUDED.name, summary = EXCLUDED.summary, updated_at = NOW()
WHERE (marketplace.marketplace_catalog_items.name,
  marketplace.marketplace_catalog_items.summary)
  IS DISTINCT FROM (EXCLUDED.name, EXCLUDED.summary)
RETURNING id
`, publisherID, release.Slug, release.Name, release.Summary,
		release.ApplicationID, release.PublisherUserID).Scan(&row).Error
	if err == nil && row.ID == 0 {
		err = tx.Raw(`
SELECT id FROM marketplace.marketplace_catalog_items
WHERE platform_resource_type = 'expert' AND platform_resource_id = ?
`, release.ApplicationID).Scan(&row).Error
	}
	if err != nil {
		return 0, err
	}
	if row.ID == 0 {
		return 0, fmt.Errorf("expert catalog item %d is required", release.ApplicationID)
	}
	return row.ID, nil
}

func ensureExpertListing(
	tx *gorm.DB,
	refs expertCatalogReferences,
	release publishedExpertRelease,
) (int64, error) {
	var row struct{ ID int64 }
	if err := tx.Raw(`
INSERT INTO marketplace.marketplace_listings
  (marketplace_id, catalog_item_id, slug, status, visibility, access_mode)
VALUES (?, ?, ?, 'approved', 'public', 'direct')
ON CONFLICT (marketplace_id, catalog_item_id) DO UPDATE
SET updated_at = NOW()
WHERE marketplace.marketplace_listings.slug IS DISTINCT FROM EXCLUDED.slug
RETURNING id
`, refs.MarketplaceID, refs.CatalogItemID, release.Slug).Scan(&row).Error; err != nil {
		return 0, err
	}
	if row.ID == 0 {
		if err := tx.Raw(`
SELECT id FROM marketplace.marketplace_listings
WHERE marketplace_id = ? AND catalog_item_id = ?
`, refs.MarketplaceID, refs.CatalogItemID).Scan(&row).Error; err != nil {
			return 0, err
		}
	}
	if row.ID == 0 {
		return 0, fmt.Errorf("expert listing %d is required", refs.CatalogItemID)
	}
	if err := tx.Exec(`
INSERT INTO marketplace.marketplace_listing_spaces
  (marketplace_id, listing_id, space_id, is_primary, sort_order)
VALUES (?, ?, ?, TRUE, 20)
ON CONFLICT (listing_id, space_id) DO UPDATE SET is_primary = TRUE
`, refs.MarketplaceID, row.ID, refs.SpaceID).Error; err != nil {
		return 0, err
	}
	return row.ID, nil
}

func (s *ExpertCatalogSynchronizer) removeListing(
	ctx context.Context,
	applicationID int64,
) error {
	return s.db.WithContext(ctx).Exec(`
UPDATE marketplace.marketplace_listings l
SET status = 'removed', visibility = 'hidden', revision = l.revision + 1,
  updated_at = NOW()
FROM marketplace.marketplace_catalog_items ci
WHERE l.catalog_item_id = ci.id
  AND ci.platform_resource_type = 'expert'
  AND ci.platform_resource_id = ?
  AND l.status <> 'removed'
`, applicationID).Error
}
