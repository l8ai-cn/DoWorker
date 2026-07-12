package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStorefrontNormalizesHostPort(t *testing.T) {
	repository := &storefrontRepositoryStub{
		resolve: func(_ context.Context, slug, host string) (MarketView, error) {
			require.Equal(t, "commerce-market", slug)
			require.Equal(t, "market.example.com", host)
			return MarketView{MarketplaceID: 42, Status: "published"}, nil
		},
	}
	storefront := NewStorefrontService(repository)

	_, err := storefront.GetMarket(context.Background(), "commerce-market", "market.example.com:443")
	require.NoError(t, err)
}

func TestStorefrontBlocksListingBrowseWhenMarketSuspended(t *testing.T) {
	repository := &storefrontRepositoryStub{
		resolve: func(context.Context, string, string) (MarketView, error) {
			return MarketView{MarketplaceID: 42, Status: "suspended"}, nil
		},
	}
	storefront := NewStorefrontService(repository)

	_, err := storefront.ListListings(context.Background(), "commerce-market", "market.example.com", 20)
	require.ErrorIs(t, err, ErrMarketSuspended)
}

func TestStorefrontKeepsListingDetailReadableWhenMarketSuspended(t *testing.T) {
	repository := &storefrontRepositoryStub{
		resolve: func(context.Context, string, string) (MarketView, error) {
			return MarketView{MarketplaceID: 42, Status: "suspended"}, nil
		},
		detail: ListingDetail{
			ListingSummary: ListingSummary{ListingID: 108, Slug: "listing-optimizer"},
		},
	}
	storefront := NewStorefrontService(repository)

	item, err := storefront.GetListing(
		context.Background(),
		"commerce-market",
		"market.example.com",
		"listing-optimizer",
	)
	require.NoError(t, err)
	require.Equal(t, int64(108), item.ListingID)
}

type storefrontRepositoryStub struct {
	resolve func(context.Context, string, string) (MarketView, error)
	detail  ListingDetail
}

func (s *storefrontRepositoryStub) ResolveMarket(
	ctx context.Context,
	slug string,
	host string,
) (MarketView, error) {
	return s.resolve(ctx, slug, host)
}

func (s *storefrontRepositoryStub) ListPublishedListings(
	context.Context,
	int64,
	int,
) ([]ListingSummary, error) {
	return nil, nil
}

func (s *storefrontRepositoryStub) GetPublishedListing(
	context.Context,
	int64,
	string,
) (ListingDetail, error) {
	return s.detail, nil
}
