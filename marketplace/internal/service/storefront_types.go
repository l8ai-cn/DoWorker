package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

var (
	ErrMarketNotFound  = errors.New("marketplace not found")
	ErrMarketSuspended = errors.New("marketplace suspended")
	ErrListingNotFound = errors.New("listing not found")
)

type MarketView struct {
	MarketplaceID int64
	Slug          string
	Name          string
	Summary       string
	Status        string
	DefaultLocale string
}

type PublisherView struct {
	Slug        string
	DisplayName string
	Verified    bool
}

type SpaceView struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type ListingSummary struct {
	ListingID        int64
	ListingVersionID int64
	Slug             string
	ResourceType     string
	DisplayName      string
	Tagline          string
	Publisher        PublisherView
	Spaces           []SpaceView
	EstimatedCredits int64
	PublishedAt      time.Time
}

type ListingDetail struct {
	ListingSummary
	Description    string
	Outcomes       json.RawMessage
	UseCases       json.RawMessage
	TargetAudience json.RawMessage
	Requirements   json.RawMessage
	Permissions    json.RawMessage
	Version        string
	ReleaseNotes   string
}

type StorefrontRepository interface {
	ResolveMarket(ctx context.Context, marketSlug, host string) (MarketView, error)
	ListPublishedListings(ctx context.Context, marketplaceID int64, limit int) ([]ListingSummary, error)
	GetPublishedListing(ctx context.Context, marketplaceID int64, listingSlug string) (ListingDetail, error)
}
