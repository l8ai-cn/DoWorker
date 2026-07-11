package public

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/anthropics/agentsmesh/marketplace/internal/service"
)

type publisherResponse struct {
	Slug        string `json:"slug"`
	DisplayName string `json:"display_name"`
	Verified    bool   `json:"verified"`
}

type listingSummaryResponse struct {
	ListingID    string              `json:"listing_id"`
	Slug         string              `json:"slug"`
	ResourceType string              `json:"resource_type"`
	DisplayName  string              `json:"display_name"`
	Tagline      string              `json:"tagline"`
	Publisher    publisherResponse   `json:"publisher"`
	Spaces       []service.SpaceView `json:"spaces"`
	PublishedAt  time.Time           `json:"published_at"`
}

type listingDetailResponse struct {
	listingSummaryResponse
	Description    string          `json:"description"`
	Outcomes       json.RawMessage `json:"outcomes"`
	UseCases       json.RawMessage `json:"use_cases"`
	TargetAudience json.RawMessage `json:"target_audience"`
	Requirements   json.RawMessage `json:"requirements"`
	Permissions    json.RawMessage `json:"permissions"`
	Version        string          `json:"version"`
	ReleaseNotes   string          `json:"release_notes"`
}

func mapListingSummary(item service.ListingSummary) listingSummaryResponse {
	return listingSummaryResponse{
		ListingID:    strconv.FormatInt(item.ListingID, 10),
		Slug:         item.Slug,
		ResourceType: item.ResourceType,
		DisplayName:  item.DisplayName,
		Tagline:      item.Tagline,
		Publisher: publisherResponse{
			Slug:        item.Publisher.Slug,
			DisplayName: item.Publisher.DisplayName,
			Verified:    item.Publisher.Verified,
		},
		Spaces:      item.Spaces,
		PublishedAt: item.PublishedAt.UTC(),
	}
}

func mapListingDetail(item service.ListingDetail) listingDetailResponse {
	return listingDetailResponse{
		listingSummaryResponse: mapListingSummary(item.ListingSummary),
		Description:            item.Description,
		Outcomes:               item.Outcomes,
		UseCases:               item.UseCases,
		TargetAudience:         item.TargetAudience,
		Requirements:           item.Requirements,
		Permissions:            item.Permissions,
		Version:                item.Version,
		ReleaseNotes:           item.ReleaseNotes,
	}
}
