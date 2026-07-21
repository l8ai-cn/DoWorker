package postgres

import (
	"context"

	"github.com/l8ai-cn/agentcloud/marketplace/internal/domain/listing"
	"github.com/l8ai-cn/agentcloud/marketplace/internal/service"
	"gorm.io/gorm"
)

func (r *ListingConsoleRepository) SaveListingCommand(
	ctx context.Context,
	item *listing.Listing,
	version *listing.Version,
	expectedRevision int64,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		marketSlug, err := lockListingMarketByID(tx, item.MarketplaceID)
		if err != nil {
			return err
		}
		if err := lockPassedCatalogVersion(
			tx,
			item.CatalogItemID,
			version.CatalogItemVersionID(),
		); err != nil {
			return err
		}
		if err := lockListingRevision(
			tx,
			item.ID,
			marketSlug,
			expectedRevision,
		); err != nil {
			return err
		}
		if err := lockListingVersion(tx, item.ID, version.ID()); err != nil {
			return err
		}
		if item.Status() == listing.StatusPublished {
			if err := lockPublishedPrimarySpace(tx, item.ID); err != nil {
				return err
			}
		}
		if err := tx.Exec(`
UPDATE marketplace.marketplace_listing_versions
SET review_status = ?
WHERE id = ? AND listing_id = ?
`, version.ReviewStatus(), version.ID(), item.ID).Error; err != nil {
			return err
		}
		result := tx.Exec(`
UPDATE marketplace.marketplace_listings
SET status = ?, visibility = ?, access_mode = ?, current_version_id = NULLIF(?, 0),
  submitted_by_platform_user_id = NULLIF(?, 0), published_at = ?,
  revision = ?, updated_at = NOW()
WHERE id = ? AND revision = ?
`, item.Status(), item.Visibility(), item.AccessMode(), item.CurrentVersionID(),
			item.SubmittedBy(), item.PublishedAt(), item.Revision(), item.ID, expectedRevision)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return service.ErrRevisionConflict
		}
		return nil
	})
}

func lockListingMarketByID(tx *gorm.DB, marketplaceID int64) (string, error) {
	var row struct {
		ID     int64
		Slug   string
		Status string
	}
	if err := tx.Raw(`
SELECT id, slug, status
FROM marketplace.marketplaces
WHERE id = ?
FOR UPDATE
`, marketplaceID).Scan(&row).Error; err != nil {
		return "", err
	}
	if row.ID == 0 {
		return "", service.ErrMarketNotFound
	}
	if row.Status != "configuring" && row.Status != "review" && row.Status != "published" {
		return "", service.ErrMarketNotConfigurable
	}
	return row.Slug, nil
}

func lockListingRevision(
	tx *gorm.DB,
	listingID int64,
	marketSlug string,
	expectedRevision int64,
) error {
	var row struct {
		ID       int64
		Revision int64
	}
	if err := tx.Raw(`
SELECT l.id, l.revision
FROM marketplace.marketplace_listings l
JOIN marketplace.marketplaces m ON m.id = l.marketplace_id
WHERE l.id = ? AND m.slug = ?
FOR UPDATE OF l
`, listingID, marketSlug).Scan(&row).Error; err != nil {
		return err
	}
	if row.ID == 0 {
		return service.ErrListingNotFound
	}
	if row.Revision != expectedRevision {
		return service.ErrRevisionConflict
	}
	return nil
}

func lockListingVersion(tx *gorm.DB, listingID, versionID int64) error {
	var row struct{ ID int64 }
	if err := tx.Raw(`
SELECT id
FROM marketplace.marketplace_listing_versions
WHERE id = ? AND listing_id = ?
FOR UPDATE
`, versionID, listingID).Scan(&row).Error; err != nil {
		return err
	}
	if row.ID == 0 {
		return service.ErrListingNotFound
	}
	return nil
}

func lockPublishedPrimarySpace(tx *gorm.DB, listingID int64) error {
	var row struct{ ID int64 }
	if err := tx.Raw(`
SELECT s.id
FROM marketplace.marketplace_listing_spaces ls
JOIN marketplace.marketplace_spaces s ON s.id = ls.space_id
WHERE ls.listing_id = ? AND ls.is_primary AND s.status = 'published'
FOR UPDATE OF s
`, listingID).Scan(&row).Error; err != nil {
		return err
	}
	if row.ID == 0 {
		return listing.ErrPrimarySpaceRequired
	}
	return nil
}
