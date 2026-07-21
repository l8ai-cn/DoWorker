package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
	"github.com/l8ai-cn/agentcloud/marketplace/internal/domain/listing"
)

var (
	ErrListingReviewForbidden  = errors.New("listing review is forbidden")
	ErrListingPublishForbidden = errors.New("listing publish is forbidden")
	ErrInvalidListingTaxonomy  = errors.New("invalid listing taxonomy")
)

type ListingPublishingRepository interface {
	ResolveListingDraftReferences(context.Context, string, int64) (int64, int64, error)
	CreateListingDraft(
		context.Context,
		string,
		*listing.Listing,
		*listing.Version,
		[]ListingTaxonomyTag,
		[]string,
		string,
	) error
	GetListingForCommand(
		context.Context,
		string,
		string,
	) (*listing.Listing, *listing.Version, error)
	HasPublishedPrimarySpace(context.Context, int64) (bool, error)
	SaveListingCommand(
		context.Context,
		*listing.Listing,
		*listing.Version,
		int64,
	) error
}

type ListingReviewAuthorizer interface {
	CanReviewListing(context.Context, int64, int64) (bool, error)
	CanPublishListing(context.Context, int64, int64) (bool, error)
}

type CreateListingDraftCommand struct {
	MarketSlug           string
	CatalogItemVersionID int64
	Slug                 string
	Visibility           listing.Visibility
	AccessMode           listing.AccessMode
	DisplayName          string
	Tagline              string
	Description          string
	Outcomes             json.RawMessage
	UseCases             json.RawMessage
	TargetAudience       json.RawMessage
	Requirements         json.RawMessage
	TaxonomyTags         []ListingTaxonomyTag
	ReleaseNotes         string
	SpaceSlugs           []string
	PrimarySpaceSlug     string
	ActorUserID          int64
}

type ListingTaxonomyTag struct {
	Slug        string
	DisplayName string
	Kind        string
}

func normalizeListingTaxonomyTags(tags []ListingTaxonomyTag) ([]ListingTaxonomyTag, []string, error) {
	if len(tags) == 0 || len(tags) > 12 {
		return nil, nil, ErrInvalidListingTaxonomy
	}
	seen := make(map[string]struct{}, len(tags))
	normalized := make([]ListingTaxonomyTag, 0, len(tags))
	displayNames := make([]string, 0, len(tags))
	for _, tag := range tags {
		tag.Slug = strings.TrimSpace(tag.Slug)
		tag.DisplayName = strings.TrimSpace(tag.DisplayName)
		if slugkit.ValidateIdentifier("listing_taxonomy.slug", tag.Slug) != nil ||
			tag.DisplayName == "" || !validTaxonomyKind(tag.Kind) {
			return nil, nil, ErrInvalidListingTaxonomy
		}
		if _, exists := seen[tag.Slug]; exists {
			return nil, nil, ErrInvalidListingTaxonomy
		}
		seen[tag.Slug] = struct{}{}
		normalized = append(normalized, tag)
		displayNames = append(displayNames, tag.DisplayName)
	}
	return normalized, displayNames, nil
}

func validTaxonomyKind(kind string) bool {
	switch kind {
	case "scene", "industry", "audience", "capability", "integration", "readiness":
		return true
	default:
		return false
	}
}

type ListingCommand struct {
	MarketSlug       string
	ListingSlug      string
	ExpectedRevision int64
	ActorUserID      int64
}

type PublishListingCommand struct {
	MarketSlug       string
	ListingSlug      string
	ExpectedRevision int64
	ActorUserID      int64
	PublishedAt      time.Time
}

type ListingResult struct {
	ListingID        int64
	Slug             string
	Status           listing.Status
	Revision         int64
	CurrentVersionID int64
}
