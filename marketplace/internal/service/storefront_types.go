package service

import (
	"context"
	"encoding/base64"
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

type TaxonomyTagView struct {
	Slug        string `json:"slug"`
	DisplayName string `json:"display_name"`
	Kind        string `json:"kind"`
}

type ListingCursor struct {
	Sort         string    `json:"s"`
	FeaturedRank int       `json:"f"`
	Relevance    int       `json:"r"`
	PublishedAt  time.Time `json:"p"`
	ListingID    int64     `json:"i"`
}

type ListingQuery struct {
	Q           string
	Scene       string
	Industry    string
	Audience    string
	Type        string
	Integration string
	Readiness   string
	Space       string
	Sort        string
	Cursor      *ListingCursor
}

type listingQueryKey struct{}

func WithListingQuery(ctx context.Context, query ListingQuery) context.Context {
	return context.WithValue(ctx, listingQueryKey{}, query)
}

func ListingQueryFromContext(ctx context.Context) ListingQuery {
	query, _ := ctx.Value(listingQueryKey{}).(ListingQuery)
	return query
}

func EncodeListingCursor(cursor ListingCursor) (string, error) {
	value, err := json.Marshal(cursor)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}

func DecodeListingCursor(value string) (ListingCursor, error) {
	raw, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return ListingCursor{}, err
	}
	var cursor ListingCursor
	if err := json.Unmarshal(raw, &cursor); err != nil {
		return ListingCursor{}, err
	}
	if cursor.Sort == "" || cursor.PublishedAt.IsZero() || cursor.ListingID <= 0 {
		return ListingCursor{}, errors.New("invalid listing cursor")
	}
	return cursor, nil
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
	Tags             []TaxonomyTagView
	EstimatedCredits int64
	PublishedAt      time.Time
	PageCursor       ListingCursor
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
