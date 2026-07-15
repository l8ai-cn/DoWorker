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
	AgentSlug         string
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
		AgentSlug:      row.AgentSlug,
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
