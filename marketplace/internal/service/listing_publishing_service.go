package service

import (
	"context"

	"github.com/anthropics/agentsmesh/marketplace/internal/domain/listing"
)

type ListingPublishingService struct {
	repository ListingPublishingRepository
	authorizer ListingReviewAuthorizer
}

func NewListingPublishingService(
	repository ListingPublishingRepository,
	authorizer ListingReviewAuthorizer,
) *ListingPublishingService {
	return &ListingPublishingService{
		repository: repository,
		authorizer: authorizer,
	}
}

func (s *ListingPublishingService) CreateDraft(
	ctx context.Context,
	command CreateListingDraftCommand,
) (ListingResult, error) {
	marketplaceID, catalogItemID, err := s.repository.ResolveListingDraftReferences(
		ctx,
		command.MarketSlug,
		command.CatalogItemVersionID,
	)
	if err != nil {
		return ListingResult{}, err
	}
	tags, displayNames, err := normalizeListingTaxonomyTags(command.TaxonomyTags)
	if err != nil {
		return ListingResult{}, err
	}
	item, err := listing.New(marketplaceID, catalogItemID, command.Slug)
	if err != nil {
		return ListingResult{}, err
	}
	if err := item.SetVisibility(command.Visibility); err != nil {
		return ListingResult{}, err
	}
	if err := item.SetAccessMode(command.AccessMode); err != nil {
		return ListingResult{}, err
	}
	version, err := listing.NewVersion(
		0,
		command.CatalogItemVersionID,
		1,
		command.DisplayName,
		command.Tagline,
		command.Description,
		command.Outcomes,
		command.UseCases,
		command.TargetAudience,
		command.Requirements,
		displayNames,
		command.ReleaseNotes,
	)
	if err != nil {
		return ListingResult{}, err
	}
	if err := s.repository.CreateListingDraft(
		ctx,
		command.MarketSlug,
		item,
		version,
		tags,
		command.SpaceSlugs,
		command.PrimarySpaceSlug,
	); err != nil {
		return ListingResult{}, err
	}
	return mapListingResult(item, item.Revision()), nil
}

func (s *ListingPublishingService) Submit(
	ctx context.Context,
	command ListingCommand,
) (ListingResult, error) {
	item, version, err := s.loadExpectedRevision(ctx, command)
	if err != nil {
		return ListingResult{}, err
	}
	if err := item.Submit(command.ActorUserID); err != nil {
		return ListingResult{}, err
	}
	if err := version.Submit(); err != nil {
		return ListingResult{}, err
	}
	if err := item.AdvanceRevision(command.ExpectedRevision); err != nil {
		return ListingResult{}, ErrRevisionConflict
	}
	if err := s.save(ctx, item, version, command.ExpectedRevision); err != nil {
		return ListingResult{}, err
	}
	return mapListingResult(item, item.Revision()), nil
}

func (s *ListingPublishingService) Approve(
	ctx context.Context,
	command ListingCommand,
) (ListingResult, error) {
	item, version, err := s.loadExpectedRevision(ctx, command)
	if err != nil {
		return ListingResult{}, err
	}
	allowed, err := s.authorizer.CanReviewListing(ctx, item.ID, command.ActorUserID)
	if err != nil {
		return ListingResult{}, err
	}
	if !allowed {
		return ListingResult{}, ErrListingReviewForbidden
	}
	if err := item.Approve(command.ActorUserID); err != nil {
		return ListingResult{}, err
	}
	if err := version.Approve(); err != nil {
		return ListingResult{}, err
	}
	if err := item.AdvanceRevision(command.ExpectedRevision); err != nil {
		return ListingResult{}, ErrRevisionConflict
	}
	if err := s.save(ctx, item, version, command.ExpectedRevision); err != nil {
		return ListingResult{}, err
	}
	return mapListingResult(item, item.Revision()), nil
}

func (s *ListingPublishingService) Publish(
	ctx context.Context,
	command PublishListingCommand,
) (ListingResult, error) {
	item, version, err := s.repository.GetListingForCommand(
		ctx,
		command.MarketSlug,
		command.ListingSlug,
	)
	if err != nil {
		return ListingResult{}, err
	}
	if item.Revision() != command.ExpectedRevision {
		return ListingResult{}, ErrRevisionConflict
	}
	allowed, err := s.authorizer.CanPublishListing(ctx, item.ID, command.ActorUserID)
	if err != nil {
		return ListingResult{}, err
	}
	if !allowed {
		return ListingResult{}, ErrListingPublishForbidden
	}
	hasPrimarySpace, err := s.repository.HasPublishedPrimarySpace(ctx, item.ID)
	if err != nil {
		return ListingResult{}, err
	}
	if err := item.Publish(version.ID(), hasPrimarySpace, command.PublishedAt); err != nil {
		return ListingResult{}, err
	}
	if err := item.AdvanceRevision(command.ExpectedRevision); err != nil {
		return ListingResult{}, ErrRevisionConflict
	}
	if err := s.save(ctx, item, version, command.ExpectedRevision); err != nil {
		return ListingResult{}, err
	}
	return mapListingResult(item, item.Revision()), nil
}
