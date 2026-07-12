package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/anthropics/agentsmesh/marketplace/internal/service"
)

type listingRow struct {
	ListingID         int64
	ListingVersionID  int64
	Slug              string
	ResourceType      string
	DisplayName       string
	Tagline           string
	PublisherSlug     string
	PublisherName     string
	PublisherVerified bool
	SpacesJSON        []byte
	TagsJSON          []byte
	EstimatedCredits  int64
	PublishedAt       time.Time
	FeaturedRank      int
	Relevance         int
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
	query := service.ListingQueryFromContext(ctx)
	if query.Sort == "" {
		query.Sort = "featured"
	}
	rows, err := r.queryPublishedListings(ctx, marketplaceID, "", query, limit)
	if err != nil {
		return nil, err
	}
	items := make([]service.ListingSummary, 0, len(rows))
	for _, row := range rows {
		item, mapErr := mapListingSummary(row)
		if mapErr != nil {
			return nil, mapErr
		}
		item.PageCursor.Sort = query.Sort
		items = append(items, item)
	}
	return items, nil
}

func (r *StorefrontRepository) GetPublishedListing(
	ctx context.Context,
	marketplaceID int64,
	listingSlug string,
) (service.ListingDetail, error) {
	rows, err := r.queryPublishedListings(
		ctx,
		marketplaceID,
		listingSlug,
		service.ListingQuery{Sort: "latest"},
		1,
	)
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
	query service.ListingQuery,
	limit int,
) ([]listingRow, error) {
	args := listingQueryArgs(marketplaceID, listingSlug, query, limit)
	queryResult := r.db.WithContext(ctx).Raw(storefrontListingQuery, args...)
	var rows []listingRow
	if err := queryResult.Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func mapListingSummary(row listingRow) (service.ListingSummary, error) {
	var spaces []service.SpaceView
	if err := json.Unmarshal(row.SpacesJSON, &spaces); err != nil {
		return service.ListingSummary{}, err
	}
	var tags []service.TaxonomyTagView
	if err := json.Unmarshal(row.TagsJSON, &tags); err != nil {
		return service.ListingSummary{}, err
	}
	return service.ListingSummary{
		ListingID:        row.ListingID,
		ListingVersionID: row.ListingVersionID,
		Slug:             row.Slug,
		ResourceType:     row.ResourceType,
		DisplayName:      row.DisplayName,
		Tagline:          row.Tagline,
		Publisher: service.PublisherView{
			Slug:        row.PublisherSlug,
			DisplayName: row.PublisherName,
			Verified:    row.PublisherVerified,
		},
		Spaces:           spaces,
		Tags:             tags,
		EstimatedCredits: row.EstimatedCredits,
		PublishedAt:      row.PublishedAt,
		PageCursor: service.ListingCursor{
			FeaturedRank: row.FeaturedRank,
			Relevance:    row.Relevance,
			PublishedAt:  row.PublishedAt,
			ListingID:    row.ListingID,
		},
	}, nil
}

func listingQueryArgs(
	marketplaceID int64,
	listingSlug string,
	query service.ListingQuery,
	limit int,
) []any {
	var cursorPublishedAt any
	cursorFeaturedRank, cursorRelevance, cursorListingID := 0, 0, int64(0)
	if query.Cursor != nil {
		cursorPublishedAt = query.Cursor.PublishedAt
		cursorFeaturedRank = query.Cursor.FeaturedRank
		cursorRelevance = query.Cursor.Relevance
		cursorListingID = query.Cursor.ListingID
	}
	return []any{
		sql.Named("marketplace_id", marketplaceID),
		sql.Named("listing_slug", listingSlug),
		sql.Named("q", query.Q),
		sql.Named("scene", query.Scene),
		sql.Named("industry", query.Industry),
		sql.Named("audience", query.Audience),
		sql.Named("type", query.Type),
		sql.Named("capability", query.Capability),
		sql.Named("integration", query.Integration),
		sql.Named("readiness", query.Readiness),
		sql.Named("space", query.Space),
		sql.Named("sort", query.Sort),
		sql.Named("cursor_published_at", cursorPublishedAt),
		sql.Named("cursor_featured_rank", cursorFeaturedRank),
		sql.Named("cursor_relevance", cursorRelevance),
		sql.Named("cursor_listing_id", cursorListingID),
		sql.Named("limit", limit),
	}
}

const storefrontListingQuery = `
SELECT l.id AS listing_id, lv.id AS listing_version_id, l.slug,
  ci.resource_type, lv.display_name, lv.tagline,
  p.slug AS publisher_slug, p.display_name AS publisher_name,
  p.verification_status = 'verified' AS publisher_verified,
  spaces.spaces_json, tags.tags_json,
  CASE
    WHEN civ.manifest->>'installation_credits'
      ~ '^([0-9]+|[0-9]+\.[0-9]{1,6})$'
    THEN ((civ.manifest->>'installation_credits')::numeric * 1000000)::bigint
    ELSE 0
  END AS estimated_credits, l.featured_rank, relevance.relevance_score AS relevance,
  l.published_at, lv.description, lv.outcomes, lv.use_cases, lv.target_audience,
  lv.requirements, civ.permissions, civ.version, lv.release_notes
FROM marketplace.marketplace_listings l
JOIN marketplace.marketplace_listing_versions lv ON lv.id = l.current_version_id
JOIN marketplace.marketplace_catalog_items ci ON ci.id = l.catalog_item_id
JOIN marketplace.marketplace_catalog_item_versions civ
  ON civ.id = lv.catalog_item_version_id AND civ.catalog_item_id = ci.id
JOIN marketplace.marketplace_publishers p ON p.id = ci.publisher_id
JOIN LATERAL (
  SELECT COALESCE(jsonb_agg(jsonb_build_object('slug', s.slug, 'name', s.name)
    ORDER BY ls.sort_order, s.slug), '[]'::jsonb) AS spaces_json
  FROM marketplace.marketplace_listing_spaces ls
  JOIN marketplace.marketplace_spaces s ON s.id = ls.space_id AND s.status = 'published'
  WHERE ls.listing_id = l.id
) spaces ON true
JOIN LATERAL (
  SELECT COALESCE(jsonb_agg(jsonb_build_object(
    'slug', t.slug, 'display_name', t.display_name, 'kind', t.kind
  ) ORDER BY t.kind, t.slug), '[]'::jsonb) AS tags_json
  FROM marketplace.marketplace_listing_version_tags lvt
  JOIN marketplace.marketplace_taxonomy_tags t ON t.id = lvt.taxonomy_tag_id
  WHERE lvt.listing_version_id = lv.id
) tags ON true
JOIN LATERAL (
  SELECT CASE
    WHEN @q = '' THEN 0
    WHEN LOWER(lv.display_name) = LOWER(@q) THEN 3
    WHEN LOWER(lv.display_name) LIKE '%' || LOWER(@q) || '%' THEN 2
    ELSE 1
  END AS relevance_score
) relevance ON true
WHERE l.marketplace_id = @marketplace_id AND l.status = 'published' AND l.visibility = 'public'
  AND lv.review_status = 'approved' AND civ.validation_status = 'passed'
  AND (@listing_slug = '' OR l.slug = @listing_slug)
  AND (@q = '' OR LOWER(lv.display_name) LIKE '%' || LOWER(@q) || '%'
    OR LOWER(lv.tagline) LIKE '%' || LOWER(@q) || '%'
    OR LOWER(p.display_name) LIKE '%' || LOWER(@q) || '%'
    OR EXISTS (
      SELECT 1 FROM marketplace.marketplace_listing_version_tags lvt
      JOIN marketplace.marketplace_taxonomy_tags t ON t.id = lvt.taxonomy_tag_id
      WHERE lvt.listing_version_id = lv.id AND LOWER(t.display_name) LIKE '%' || LOWER(@q) || '%'
    ))
  AND (@scene = '' OR EXISTS (
    SELECT 1 FROM marketplace.marketplace_listing_version_tags lvt
    JOIN marketplace.marketplace_taxonomy_tags t ON t.id = lvt.taxonomy_tag_id
    WHERE lvt.listing_version_id = lv.id AND t.kind = 'scene' AND t.slug = @scene
  ))
  AND (@industry = '' OR EXISTS (
    SELECT 1 FROM marketplace.marketplace_listing_version_tags lvt
    JOIN marketplace.marketplace_taxonomy_tags t ON t.id = lvt.taxonomy_tag_id
    WHERE lvt.listing_version_id = lv.id AND t.kind = 'industry' AND t.slug = @industry
  ))
  AND (@audience = '' OR EXISTS (
    SELECT 1 FROM marketplace.marketplace_listing_version_tags lvt
    JOIN marketplace.marketplace_taxonomy_tags t ON t.id = lvt.taxonomy_tag_id
    WHERE lvt.listing_version_id = lv.id AND t.kind = 'audience' AND t.slug = @audience
  ))
  AND (@type = '' OR ci.resource_type = @type)
  AND (@capability = '' OR EXISTS (
    SELECT 1 FROM marketplace.marketplace_listing_version_tags lvt
    JOIN marketplace.marketplace_taxonomy_tags t ON t.id = lvt.taxonomy_tag_id
    WHERE lvt.listing_version_id = lv.id AND t.kind = 'capability' AND t.slug = @capability
  ))
  AND (@integration = '' OR EXISTS (
    SELECT 1 FROM marketplace.marketplace_listing_version_tags lvt
    JOIN marketplace.marketplace_taxonomy_tags t ON t.id = lvt.taxonomy_tag_id
    WHERE lvt.listing_version_id = lv.id AND t.kind = 'integration' AND t.slug = @integration
  ))
  AND (@readiness = '' OR EXISTS (
    SELECT 1 FROM marketplace.marketplace_listing_version_tags lvt
    JOIN marketplace.marketplace_taxonomy_tags t ON t.id = lvt.taxonomy_tag_id
    WHERE lvt.listing_version_id = lv.id AND t.kind = 'readiness' AND t.slug = @readiness
  ))
  AND (@space = '' OR EXISTS (
    SELECT 1 FROM marketplace.marketplace_listing_spaces ls
    JOIN marketplace.marketplace_spaces s ON s.id = ls.space_id AND s.status = 'published'
    WHERE ls.listing_id = l.id AND s.slug = @space
  ))
  AND (CAST(@cursor_published_at AS timestamptz) IS NULL
    OR (@sort = 'featured' AND (l.featured_rank, l.published_at, l.id) <
      (@cursor_featured_rank, CAST(@cursor_published_at AS timestamptz), @cursor_listing_id))
    OR (@sort = 'relevance' AND (relevance.relevance_score, l.published_at, l.id) <
      (@cursor_relevance, CAST(@cursor_published_at AS timestamptz), @cursor_listing_id))
    OR (@sort = 'latest' AND (l.published_at, l.id) <
      (CAST(@cursor_published_at AS timestamptz), @cursor_listing_id)))
ORDER BY
  CASE WHEN @sort = 'featured' THEN l.featured_rank ELSE 0 END DESC,
  CASE WHEN @sort = 'relevance' THEN relevance.relevance_score ELSE 0 END DESC,
  l.published_at DESC, l.id DESC
LIMIT @limit`
