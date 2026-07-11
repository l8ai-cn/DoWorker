package postgres

import (
	"context"

	"github.com/anthropics/agentsmesh/marketplace/internal/domain/listing"
	"github.com/anthropics/agentsmesh/marketplace/internal/service"
	"gorm.io/gorm"
)

type ListingConsoleRepository struct {
	db *gorm.DB
}

func NewListingConsoleRepository(db *gorm.DB) *ListingConsoleRepository {
	return &ListingConsoleRepository{db: db}
}

func (r *ListingConsoleRepository) ResolveListingDraftReferences(
	ctx context.Context,
	marketSlug string,
	catalogItemVersionID int64,
) (int64, int64, error) {
	var row struct {
		MarketplaceID int64
		CatalogItemID int64
	}
	result := r.db.WithContext(ctx).Raw(`
SELECT m.id AS marketplace_id, civ.catalog_item_id
FROM marketplace.marketplaces m
JOIN marketplace.marketplace_catalog_item_versions civ ON civ.id = ?
JOIN marketplace.marketplace_catalog_items ci ON ci.id = civ.catalog_item_id
WHERE m.slug = ? AND m.status IN ('configuring', 'review', 'published')
  AND civ.validation_status = 'passed' AND ci.status = 'active'
LIMIT 1
`, catalogItemVersionID, marketSlug).Scan(&row)
	if result.Error != nil {
		return 0, 0, result.Error
	}
	if row.MarketplaceID == 0 {
		return 0, 0, service.ErrCatalogVersionNotPassed
	}
	return row.MarketplaceID, row.CatalogItemID, nil
}

func (r *ListingConsoleRepository) GetListingForCommand(
	ctx context.Context,
	marketSlug string,
	listingSlug string,
) (*listing.Listing, *listing.Version, error) {
	item, err := r.loadListing(ctx, marketSlug, listingSlug)
	if err != nil {
		return nil, nil, err
	}
	version, err := r.loadLatestListingVersion(ctx, item.ID)
	if err != nil {
		return nil, nil, err
	}
	return item, version, nil
}

func (r *ListingConsoleRepository) HasPublishedPrimarySpace(
	ctx context.Context,
	listingID int64,
) (bool, error) {
	var row struct{ Found bool }
	err := r.db.WithContext(ctx).Raw(`
SELECT EXISTS (
  SELECT 1
  FROM marketplace.marketplace_listing_spaces ls
  JOIN marketplace.marketplace_spaces s ON s.id = ls.space_id
  WHERE ls.listing_id = ? AND ls.is_primary AND s.status = 'published'
) AS found
`, listingID).Scan(&row).Error
	return row.Found, err
}
