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

type quotaResponse struct {
	Mode                  string `json:"mode"`
	EstimatedCreditsMicro string `json:"estimated_credits_micro"`
}

type listingSummaryResponse struct {
	ListingID        string                    `json:"listing_id"`
	ListingVersionID string                    `json:"listing_version_id"`
	Slug             string                    `json:"slug"`
	ResourceType     string                    `json:"resource_type"`
	DisplayName      string                    `json:"display_name"`
	Tagline          string                    `json:"tagline"`
	Publisher        publisherResponse         `json:"publisher"`
	Spaces           []service.SpaceView       `json:"spaces"`
	Tags             []service.TaxonomyTagView `json:"tags"`
	Quota            *quotaResponse            `json:"quota,omitempty"`
	PublishedAt      time.Time                 `json:"published_at"`
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
	response := listingSummaryResponse{
		ListingID:        strconv.FormatInt(item.ListingID, 10),
		ListingVersionID: strconv.FormatInt(item.ListingVersionID, 10),
		Slug:             item.Slug,
		ResourceType:     item.ResourceType,
		DisplayName:      item.DisplayName,
		Tagline:          item.Tagline,
		Publisher: publisherResponse{
			Slug:        item.Publisher.Slug,
			DisplayName: item.Publisher.DisplayName,
			Verified:    item.Publisher.Verified,
		},
		Spaces:      nonNilSpaces(item.Spaces),
		Tags:        nonNilTags(item.Tags),
		PublishedAt: item.PublishedAt.UTC(),
	}
	if item.EstimatedCredits > 0 {
		response.Quota = &quotaResponse{
			Mode:                  "per_install",
			EstimatedCreditsMicro: strconv.FormatInt(item.EstimatedCredits, 10),
		}
	}
	return response
}

func nonNilSpaces(spaces []service.SpaceView) []service.SpaceView {
	if spaces == nil {
		return []service.SpaceView{}
	}
	return spaces
}

func nonNilTags(tags []service.TaxonomyTagView) []service.TaxonomyTagView {
	if tags == nil {
		return []service.TaxonomyTagView{}
	}
	return tags
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
