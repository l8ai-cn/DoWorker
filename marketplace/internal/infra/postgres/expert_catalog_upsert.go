package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type expertCatalogReferences struct {
	MarketplaceID    int64
	SpaceID          int64
	QuotaPlanID      int64
	PublisherID      int64
	CatalogItemID    int64
	CatalogVersionID int64
	ListingID        int64
	ListingVersionID int64
}

func (s *ExpertCatalogSynchronizer) publishListing(
	ctx context.Context,
	release publishedExpertRelease,
	payload expertCatalogPayload,
) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		refs, err := ensureExpertCatalogReferences(tx, release, payload)
		if err != nil {
			return err
		}
		outcomes, _ := json.Marshal([]string(release.Outcomes))
		tags := []string(release.Tags)
		listingVersionID, err := ensureExpertListingVersion(
			tx, refs, release, outcomes, tags,
		)
		if err != nil {
			return err
		}
		refs.ListingVersionID = listingVersionID
		return tx.Exec(`
UPDATE marketplace.marketplace_listings
SET status = 'published', visibility = 'public', access_mode = 'direct',
  current_version_id = ?, submitted_by_platform_user_id = NULLIF(?, 0),
  published_at = ?, featured_rank = ?, revision = revision + 1, updated_at = NOW()
WHERE id = ?
  AND (status <> 'published' OR visibility <> 'public' OR access_mode <> 'direct'
    OR current_version_id IS DISTINCT FROM ?
    OR COALESCE(submitted_by_platform_user_id, 0) IS DISTINCT FROM ?
    OR published_at IS DISTINCT FROM ? OR featured_rank IS DISTINCT FROM ?)
`, refs.ListingVersionID, release.ReviewerUserID, release.PublishedAt,
			featuredRank(release.Featured), refs.ListingID,
			refs.ListingVersionID, release.ReviewerUserID, release.PublishedAt,
			featuredRank(release.Featured)).Error
	})
}

func ensureExpertCatalogReferences(
	tx *gorm.DB,
	release publishedExpertRelease,
	payload expertCatalogPayload,
) (expertCatalogReferences, error) {
	refs, err := ensureExpertMarket(tx)
	if err != nil {
		return refs, err
	}
	refs.PublisherID, err = ensureExpertPublisher(tx, release)
	if err != nil {
		return refs, err
	}
	refs.CatalogItemID, err = ensureExpertCatalogItem(tx, refs.PublisherID, release)
	if err != nil {
		return refs, err
	}
	refs.CatalogVersionID, err = ensureExpertCatalogVersion(
		tx, refs.CatalogItemID, release, payload,
	)
	if err != nil {
		return refs, err
	}
	refs.ListingID, err = ensureExpertListing(tx, refs, release)
	return refs, err
}

func ensureExpertCatalogVersion(
	tx *gorm.DB,
	catalogItemID int64,
	release publishedExpertRelease,
	payload expertCatalogPayload,
) (int64, error) {
	version := strconv.Itoa(release.Version) + ".0.0"
	if err := tx.Exec(`
INSERT INTO marketplace.marketplace_catalog_item_versions
  (catalog_item_id, version, source_revision, content_digest, manifest,
   permissions, compatibility, dependency_lock, validation_status,
   created_by_platform_user_id)
VALUES (?, ?, ?, ?, ?::jsonb, '["workspace.execute"]'::jsonb, ?::jsonb,
  ?::jsonb, 'passed', ?)
ON CONFLICT (catalog_item_id, version) DO NOTHING
`, catalogItemID, version, fmt.Sprintf("expert-release-%d", release.ReleaseID),
		payload.ContentDigest, string(payload.Manifest), string(payload.Compatibility),
		string(payload.DependencyLock), release.PublisherUserID).Error; err != nil {
		return 0, err
	}
	var row struct {
		ID            int64
		ContentDigest string
	}
	if err := tx.Raw(`
SELECT id, content_digest
FROM marketplace.marketplace_catalog_item_versions
WHERE catalog_item_id = ? AND version = ?
`, catalogItemID, version).Scan(&row).Error; err != nil {
		return 0, err
	}
	if row.ID == 0 || row.ContentDigest != payload.ContentDigest {
		return 0, fmt.Errorf("expert catalog version conflicts with release %d", release.ReleaseID)
	}
	if err := tx.Exec(`
UPDATE marketplace.marketplace_catalog_items
SET status = 'active', latest_version_id = ?, name = ?, summary = ?,
  revision = revision + 1, updated_at = NOW()
WHERE id = ?
  AND (status <> 'active' OR latest_version_id IS DISTINCT FROM ?
    OR name IS DISTINCT FROM ? OR summary IS DISTINCT FROM ?)
`, row.ID, release.Name, release.Summary, catalogItemID,
		row.ID, release.Name, release.Summary).Error; err != nil {
		return 0, err
	}
	return row.ID, nil
}

func ensureExpertListingVersion(
	tx *gorm.DB,
	refs expertCatalogReferences,
	release publishedExpertRelease,
	outcomes []byte,
	tags []string,
) (int64, error) {
	if err := tx.Exec(`
INSERT INTO marketplace.marketplace_listing_versions
  (listing_id, catalog_item_id, catalog_item_version_id, revision, display_name,
   tagline, description, outcomes, use_cases, target_audience, requirements,
   tags, quota_plan_id, release_notes, review_status)
VALUES (?, ?, ?, ?, ?, ?, ?, ?::jsonb, '[]'::jsonb,
  '["内容团队","品牌团队","视频创作者"]'::jsonb,
  '["目标组织已配置可用 Runner","目标组织已配置可用模型资源"]'::jsonb,
  ?, ?, ?, 'approved')
ON CONFLICT (listing_id, revision) DO NOTHING
`, refs.ListingID, refs.CatalogItemID, refs.CatalogVersionID, release.Version,
		release.Name, release.Summary, release.Description, string(outcomes),
		pq.Array(tags), refs.QuotaPlanID,
		fmt.Sprintf("专家市场审核发布版本 %d。", release.Version)).Error; err != nil {
		return 0, err
	}
	var row struct {
		ID                   int64
		CatalogItemVersionID int64
	}
	if err := tx.Raw(`
SELECT id, catalog_item_version_id
FROM marketplace.marketplace_listing_versions
WHERE listing_id = ? AND revision = ?
`, refs.ListingID, release.Version).Scan(&row).Error; err != nil {
		return 0, err
	}
	if row.ID == 0 || row.CatalogItemVersionID != refs.CatalogVersionID {
		return 0, fmt.Errorf(
			"expert listing version conflicts with release %d",
			release.ReleaseID,
		)
	}
	return row.ID, nil
}

func featuredRank(featured bool) int {
	if featured {
		return 100
	}
	return 0
}
