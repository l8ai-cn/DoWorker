package service

import (
	"context"

	"github.com/anthropics/agentsmesh/marketplace/internal/domain/listing"
)

func (s *ListingPublishingService) loadExpectedRevision(
	ctx context.Context,
	command ListingCommand,
) (*listing.Listing, *listing.Version, error) {
	item, version, err := s.repository.GetListingForCommand(
		ctx,
		command.MarketSlug,
		command.ListingSlug,
	)
	if err != nil {
		return nil, nil, err
	}
	if item.Revision() != command.ExpectedRevision {
		return nil, nil, ErrRevisionConflict
	}
	return item, version, nil
}

func (s *ListingPublishingService) save(
	ctx context.Context,
	item *listing.Listing,
	version *listing.Version,
	expectedRevision int64,
) error {
	return s.repository.SaveListingCommand(ctx, item, version, expectedRevision)
}

func mapListingResult(item *listing.Listing, revision int64) ListingResult {
	return ListingResult{
		ListingID:        item.ID,
		Slug:             item.Slug().String(),
		Status:           item.Status(),
		Revision:         revision,
		CurrentVersionID: item.CurrentVersionID(),
	}
}
