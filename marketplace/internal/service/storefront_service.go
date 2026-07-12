package service

import (
	"context"
	"net"
	"strings"
)

type StorefrontService struct {
	repository StorefrontRepository
}

func NewStorefrontService(repository StorefrontRepository) *StorefrontService {
	return &StorefrontService{repository: repository}
}

func (s *StorefrontService) GetMarket(
	ctx context.Context,
	marketSlug string,
	requestHost string,
) (MarketView, error) {
	return s.repository.ResolveMarket(ctx, marketSlug, normalizeHost(requestHost))
}

func (s *StorefrontService) ListListings(
	ctx context.Context,
	marketSlug string,
	requestHost string,
	limit int,
) ([]ListingSummary, error) {
	market, err := s.GetMarket(ctx, marketSlug, requestHost)
	if err != nil {
		return nil, err
	}
	if market.Status == "suspended" {
		return nil, ErrMarketSuspended
	}
	return s.repository.ListPublishedListings(ctx, market.MarketplaceID, limit)
}

func (s *StorefrontService) GetListing(
	ctx context.Context,
	marketSlug string,
	requestHost string,
	listingSlug string,
) (ListingDetail, error) {
	market, err := s.GetMarket(ctx, marketSlug, requestHost)
	if err != nil {
		return ListingDetail{}, err
	}
	return s.repository.GetPublishedListing(ctx, market.MarketplaceID, listingSlug)
}

func normalizeHost(requestHost string) string {
	host := strings.TrimSpace(strings.ToLower(requestHost))
	if parsed, _, err := net.SplitHostPort(host); err == nil {
		host = parsed
	}
	return strings.TrimSuffix(host, ".")
}
