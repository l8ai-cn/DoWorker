package postgres

import (
	"context"
	"time"

	"github.com/l8ai-cn/agentcloud/marketplace/internal/domain/listing"
	"github.com/l8ai-cn/agentcloud/marketplace/internal/service"
	"github.com/lib/pq"
)

func (r *ListingConsoleRepository) loadListing(
	ctx context.Context,
	marketSlug string,
	listingSlug string,
) (*listing.Listing, error) {
	var row struct {
		ID               int64
		MarketplaceID    int64
		CatalogItemID    int64
		Slug             string
		Status           listing.Status
		Visibility       listing.Visibility
		AccessMode       listing.AccessMode
		CurrentVersionID int64
		SubmittedBy      int64
		PublishedAt      *time.Time
		HasPrimarySpace  bool
		Revision         int64
	}
	result := r.db.WithContext(ctx).Raw(`
SELECT l.id, l.marketplace_id, l.catalog_item_id, l.slug, l.status, l.visibility,
  l.access_mode, COALESCE(l.current_version_id, 0) AS current_version_id,
  COALESCE(l.submitted_by_platform_user_id, 0) AS submitted_by,
  l.published_at, l.revision,
  EXISTS (
    SELECT 1
    FROM marketplace.marketplace_listing_spaces ls
    JOIN marketplace.marketplace_spaces s ON s.id = ls.space_id
    WHERE ls.listing_id = l.id AND ls.is_primary AND s.status = 'published'
  ) AS has_primary_space
FROM marketplace.marketplace_listings l
JOIN marketplace.marketplaces m ON m.id = l.marketplace_id
WHERE m.slug = ? AND l.slug = ?
LIMIT 1
`, marketSlug, listingSlug).Scan(&row)
	if result.Error != nil {
		return nil, result.Error
	}
	if row.ID == 0 {
		return nil, service.ErrListingNotFound
	}
	return listing.Restore(listing.State{
		ID:               row.ID,
		MarketplaceID:    row.MarketplaceID,
		CatalogItemID:    row.CatalogItemID,
		Slug:             row.Slug,
		Status:           row.Status,
		Visibility:       row.Visibility,
		AccessMode:       row.AccessMode,
		CurrentVersionID: row.CurrentVersionID,
		SubmittedBy:      row.SubmittedBy,
		PublishedAt:      row.PublishedAt,
		HasPrimarySpace:  row.HasPrimarySpace,
		Revision:         row.Revision,
	})
}

func (r *ListingConsoleRepository) loadLatestListingVersion(
	ctx context.Context,
	listingID int64,
) (*listing.Version, error) {
	var row listingVersionRow
	result := r.db.WithContext(ctx).Raw(`
SELECT id, listing_id, catalog_item_version_id, revision, display_name, tagline,
  description, outcomes, use_cases, target_audience, requirements, tags,
  release_notes, review_status
FROM marketplace.marketplace_listing_versions
WHERE listing_id = ?
ORDER BY revision DESC
LIMIT 1
`, listingID).Scan(&row)
	if result.Error != nil {
		return nil, result.Error
	}
	if row.ID == 0 {
		return nil, service.ErrListingNotFound
	}
	return listing.RestoreVersion(row.state())
}

type listingVersionRow struct {
	ID                   int64
	ListingID            int64
	CatalogItemVersionID int64
	Revision             int
	DisplayName          string
	Tagline              string
	Description          string
	Outcomes             []byte
	UseCases             []byte
	TargetAudience       []byte
	Requirements         []byte
	Tags                 pq.StringArray `gorm:"type:text[]"`
	ReleaseNotes         string
	ReviewStatus         listing.ReviewStatus
}

func (r listingVersionRow) state() listing.VersionState {
	return listing.VersionState{
		ID:                   r.ID,
		ListingID:            r.ListingID,
		CatalogItemVersionID: r.CatalogItemVersionID,
		Revision:             r.Revision,
		DisplayName:          r.DisplayName,
		Tagline:              r.Tagline,
		Description:          r.Description,
		Outcomes:             r.Outcomes,
		UseCases:             r.UseCases,
		TargetAudience:       r.TargetAudience,
		Requirements:         r.Requirements,
		Tags:                 []string(r.Tags),
		ReleaseNotes:         r.ReleaseNotes,
		ReviewStatus:         r.ReviewStatus,
	}
}
