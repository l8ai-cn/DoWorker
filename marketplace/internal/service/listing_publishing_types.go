package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/anthropics/agentsmesh/marketplace/internal/domain/listing"
)

var (
	ErrListingReviewForbidden  = errors.New("listing review is forbidden")
	ErrListingPublishForbidden = errors.New("listing publish is forbidden")
)

type ListingPublishingRepository interface {
	ResolveListingDraftReferences(context.Context, string, int64) (int64, int64, error)
	CreateListingDraft(
		context.Context,
		string,
		*listing.Listing,
		*listing.Version,
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
	Tags                 []string
	ReleaseNotes         string
	SpaceSlugs           []string
	PrimarySpaceSlug     string
	ActorUserID          int64
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
