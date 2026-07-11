package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/anthropics/agentsmesh/marketplace/internal/service"
)

type listingRow struct {
	ListingID         int64
	Slug              string
	ResourceType      string
	DisplayName       string
	Tagline           string
	PublisherSlug     string
	PublisherName     string
	PublisherVerified bool
	SpacesJSON        []byte
	PublishedAt       time.Time
	Description       string
	Outcomes          []byte
	UseCases          []byte
	TargetAudience    []byte
	Requirements      []byte
	Permissions       []byte
	Version           string
	ReleaseNotes      string
}

func (r *StorefrontRepository) ListPublishedListings(
	ctx context.Context,
	marketplaceID int64,
	limit int,
) ([]service.ListingSummary, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, err := r.queryPublishedListings(ctx, marketplaceID, "", limit)
	if err != nil {
		return nil, err
	}
	items := make([]service.ListingSummary, 0, len(rows))
	for _, row := range rows {
		item, mapErr := mapListingSummary(row)
		if mapErr != nil {
			return nil, mapErr
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *StorefrontRepository) GetPublishedListing(
	ctx context.Context,
	marketplaceID int64,
	listingSlug string,
) (service.ListingDetail, error) {
	rows, err := r.queryPublishedListings(ctx, marketplaceID, listingSlug, 1)
	if err != nil {
		return service.ListingDetail{}, err
	}
	if len(rows) == 0 {
		return service.ListingDetail{}, service.ErrListingNotFound
	}
	summary, err := mapListingSummary(rows[0])
	if err != nil {
		return service.ListingDetail{}, err
	}
	row := rows[0]
	return service.ListingDetail{
		ListingSummary: summary,
		Description:    row.Description,
		Outcomes:       json.RawMessage(row.Outcomes),
		UseCases:       json.RawMessage(row.UseCases),
		TargetAudience: json.RawMessage(row.TargetAudience),
		Requirements:   json.RawMessage(row.Requirements),
		Permissions:    json.RawMessage(row.Permissions),
		Version:        row.Version,
		ReleaseNotes:   row.ReleaseNotes,
	}, nil
}

func (r *StorefrontRepository) queryPublishedListings(
	ctx context.Context,
	marketplaceID int64,
	listingSlug string,
	limit int,
) ([]listingRow, error) {
	query := r.db.WithContext(ctx).Raw(storefrontListingQuery, marketplaceID, listingSlug, listingSlug, limit)
	var rows []listingRow
	if err := query.Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func mapListingSummary(row listingRow) (service.ListingSummary, error) {
	var spaces []service.SpaceView
	if err := json.Unmarshal(row.SpacesJSON, &spaces); err != nil {
		return service.ListingSummary{}, err
	}
	return service.ListingSummary{
		ListingID:    row.ListingID,
		Slug:         row.Slug,
		ResourceType: row.ResourceType,
		DisplayName:  row.DisplayName,
		Tagline:      row.Tagline,
		Publisher: service.PublisherView{
			Slug:        row.PublisherSlug,
			DisplayName: row.PublisherName,
			Verified:    row.PublisherVerified,
		},
		Spaces:      spaces,
		PublishedAt: row.PublishedAt,
	}, nil
}

const storefrontListingQuery = `
SELECT l.id AS listing_id, l.slug, ci.resource_type, lv.display_name, lv.tagline,
  p.slug AS publisher_slug, p.display_name AS publisher_name,
  p.verification_status = 'verified' AS publisher_verified,
  COALESCE(jsonb_agg(DISTINCT jsonb_build_object('slug', s.slug, 'name', s.name))
    FILTER (WHERE s.id IS NOT NULL), '[]') AS spaces_json,
  l.published_at, lv.description, lv.outcomes, lv.use_cases, lv.target_audience,
  lv.requirements, civ.permissions, civ.version, lv.release_notes
FROM marketplace.marketplace_listings l
JOIN marketplace.marketplace_listing_versions lv ON lv.id = l.current_version_id
JOIN marketplace.marketplace_catalog_items ci ON ci.id = l.catalog_item_id
JOIN marketplace.marketplace_catalog_item_versions civ
  ON civ.id = lv.catalog_item_version_id AND civ.catalog_item_id = ci.id
JOIN marketplace.marketplace_publishers p ON p.id = ci.publisher_id
JOIN marketplace.marketplace_listing_spaces ls ON ls.listing_id = l.id
JOIN marketplace.marketplace_spaces s ON s.id = ls.space_id
WHERE l.marketplace_id = ? AND l.status = 'published' AND l.visibility = 'public'
  AND (? = '' OR l.slug = ?)
GROUP BY l.id, ci.resource_type, lv.id, p.id, civ.id
ORDER BY l.published_at DESC, l.id DESC
LIMIT ?`
