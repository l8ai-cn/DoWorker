package service

import (
	"context"
	"testing"
	"time"

	"github.com/l8ai-cn/agentcloud/marketplace/internal/domain/listing"
	"github.com/stretchr/testify/require"
)

func TestListingPublishingLifecycle(t *testing.T) {
	repository := &listingRepositoryStub{catalogItemID: 41}
	authorizer := &listingAuthorizerStub{allowed: true}
	publishing := NewListingPublishingService(repository, authorizer)
	ctx := context.Background()

	draft, err := publishing.CreateDraft(ctx, CreateListingDraftCommand{
		MarketSlug:           "commerce-market",
		CatalogItemVersionID: 51,
		Slug:                 "listing-optimizer",
		Visibility:           listing.VisibilityPublic,
		AccessMode:           listing.AccessModeDirect,
		DisplayName:          "商品优化应用",
		Tagline:              "批量优化商品信息",
		Description:          "面向跨境电商运营团队",
		Outcomes:             []byte(`["提升转化"]`),
		UseCases:             []byte(`["批量优化"]`),
		TargetAudience:       []byte(`["跨境运营"]`),
		Requirements:         []byte(`[]`),
		TaxonomyTags: []ListingTaxonomyTag{
			{Slug: "cross-border-commerce", DisplayName: "跨境电商", Kind: "industry"},
		},
		ReleaseNotes:     "首次发布",
		SpaceSlugs:       []string{"operations"},
		PrimarySpaceSlug: "operations",
		ActorUserID:      14,
	})
	require.NoError(t, err)
	require.Equal(t, listing.StatusDraft, draft.Status)

	submitted, err := publishing.Submit(ctx, ListingCommand{
		MarketSlug:       "commerce-market",
		ListingSlug:      draft.Slug,
		ExpectedRevision: draft.Revision,
		ActorUserID:      14,
	})
	require.NoError(t, err)
	require.Equal(t, listing.StatusSubmitted, submitted.Status)

	approved, err := publishing.Approve(ctx, ListingCommand{
		MarketSlug:       "commerce-market",
		ListingSlug:      draft.Slug,
		ExpectedRevision: submitted.Revision,
		ActorUserID:      21,
	})
	require.NoError(t, err)
	require.Equal(t, listing.StatusApproved, approved.Status)

	published, err := publishing.Publish(ctx, PublishListingCommand{
		MarketSlug:       "commerce-market",
		ListingSlug:      draft.Slug,
		ExpectedRevision: approved.Revision,
		ActorUserID:      21,
		PublishedAt:      time.Now().UTC(),
	})
	require.NoError(t, err)
	require.Equal(t, listing.StatusPublished, published.Status)
	require.Equal(t, int64(71), published.CurrentVersionID)
}

func TestListingPublishingPassesNormalizedTaxonomyTags(t *testing.T) {
	repository := &listingRepositoryStub{catalogItemID: 41}
	publishing := NewListingPublishingService(repository, &listingAuthorizerStub{allowed: true})

	_, err := publishing.CreateDraft(context.Background(), CreateListingDraftCommand{
		MarketSlug:           "commerce-market",
		CatalogItemVersionID: 51,
		Slug:                 "listing-optimizer",
		Visibility:           listing.VisibilityPublic,
		AccessMode:           listing.AccessModeDirect,
		DisplayName:          "商品优化应用",
		Tagline:              "批量优化商品信息",
		Description:          "面向跨境电商运营团队",
		Outcomes:             []byte(`["提升转化"]`),
		UseCases:             []byte(`["批量优化"]`),
		TargetAudience:       []byte(`["跨境运营"]`),
		Requirements:         []byte(`[]`),
		TaxonomyTags: []ListingTaxonomyTag{
			{Slug: "cross-border-commerce", DisplayName: "跨境电商", Kind: "industry"},
			{Slug: "catalog-optimization", DisplayName: "商品优化", Kind: "capability"},
		},
		ReleaseNotes:     "首次发布",
		SpaceSlugs:       []string{"operations"},
		PrimarySpaceSlug: "operations",
		ActorUserID:      14,
	})

	require.NoError(t, err)
	require.Equal(t, []ListingTaxonomyTag{
		{Slug: "cross-border-commerce", DisplayName: "跨境电商", Kind: "industry"},
		{Slug: "catalog-optimization", DisplayName: "商品优化", Kind: "capability"},
	}, repository.taxonomyTags)
}

func TestListingPublishingRequiresNormalizedTaxonomyTags(t *testing.T) {
	repository := &listingRepositoryStub{catalogItemID: 41}
	publishing := NewListingPublishingService(repository, &listingAuthorizerStub{allowed: true})

	_, err := publishing.CreateDraft(context.Background(), CreateListingDraftCommand{
		MarketSlug: "commerce-market", CatalogItemVersionID: 51, Slug: "listing-optimizer",
		Visibility: listing.VisibilityPublic, AccessMode: listing.AccessModeDirect,
		DisplayName: "商品优化应用", Tagline: "批量优化商品信息", Description: "面向跨境电商运营团队",
		Outcomes: []byte(`["提升转化"]`), UseCases: []byte(`["批量优化"]`),
		TargetAudience: []byte(`["跨境运营"]`), Requirements: []byte(`[]`),
		SpaceSlugs: []string{"operations"}, PrimarySpaceSlug: "operations", ActorUserID: 14,
	})

	require.ErrorIs(t, err, ErrInvalidListingTaxonomy)
	require.Nil(t, repository.item)
}

func TestListingPublishRequiresAuthorization(t *testing.T) {
	repository := listingRepositoryWithApproved(t)
	publishing := NewListingPublishingService(repository, &listingAuthorizerStub{})

	_, err := publishing.Publish(context.Background(), PublishListingCommand{
		MarketSlug: "commerce-market", ListingSlug: "listing-optimizer",
		ExpectedRevision: 3, ActorUserID: 21, PublishedAt: time.Now().UTC(),
	})
	require.ErrorIs(t, err, ErrListingPublishForbidden)
}

func TestListingApprovalRequiresAuthorizedIndependentReviewer(t *testing.T) {
	repository := listingRepositoryWithSubmitted(t)
	publishing := NewListingPublishingService(repository, &listingAuthorizerStub{})

	_, err := publishing.Approve(context.Background(), ListingCommand{
		MarketSlug:       "commerce-market",
		ListingSlug:      "listing-optimizer",
		ExpectedRevision: 2,
		ActorUserID:      21,
	})
	require.ErrorIs(t, err, ErrListingReviewForbidden)

	publishing = NewListingPublishingService(repository, &listingAuthorizerStub{allowed: true})
	_, err = publishing.Approve(context.Background(), ListingCommand{
		MarketSlug:       "commerce-market",
		ListingSlug:      "listing-optimizer",
		ExpectedRevision: 2,
		ActorUserID:      14,
	})
	require.ErrorIs(t, err, listing.ErrSubmitterCannotReview)
}

func listingRepositoryWithSubmitted(t *testing.T) *listingRepositoryStub {
	t.Helper()
	item, err := listing.New(42, 41, "listing-optimizer")
	require.NoError(t, err)
	require.NoError(t, item.Submit(14))
	require.NoError(t, item.AdvanceRevision(1))
	version, err := listing.NewVersion(61, 51, 1, "商品优化应用", "批量优化",
		"完整介绍", []byte(`[]`), []byte(`[]`), []byte(`[]`), []byte(`[]`), nil, "")
	require.NoError(t, err)
	require.NoError(t, version.Submit())
	return &listingRepositoryStub{item: item, version: version, catalogItemID: 41}
}

func listingRepositoryWithApproved(t *testing.T) *listingRepositoryStub {
	t.Helper()
	repository := listingRepositoryWithSubmitted(t)
	require.NoError(t, repository.item.Approve(21))
	require.NoError(t, repository.item.AdvanceRevision(2))
	require.NoError(t, repository.version.Approve())
	return repository
}

type listingRepositoryStub struct {
	item          *listing.Listing
	version       *listing.Version
	catalogItemID int64
	taxonomyTags  []ListingTaxonomyTag
}

func (r *listingRepositoryStub) ResolveListingDraftReferences(
	context.Context,
	string,
	int64,
) (int64, int64, error) {
	return 42, r.catalogItemID, nil
}

func (r *listingRepositoryStub) CreateListingDraft(
	_ context.Context,
	_ string,
	item *listing.Listing,
	version *listing.Version,
	tags []ListingTaxonomyTag,
	_ []string,
	_ string,
) error {
	item.ID = 61
	version.AssignID(71)
	r.item = item
	r.version = version
	r.taxonomyTags = append([]ListingTaxonomyTag(nil), tags...)
	return nil
}

func (r *listingRepositoryStub) GetListingForCommand(
	context.Context,
	string,
	string,
) (*listing.Listing, *listing.Version, error) {
	return r.item, r.version, nil
}

func (r *listingRepositoryStub) HasPublishedPrimarySpace(context.Context, int64) (bool, error) {
	return true, nil
}

func (r *listingRepositoryStub) SaveListingCommand(
	_ context.Context,
	item *listing.Listing,
	version *listing.Version,
	_ int64,
) error {
	r.item = item
	r.version = version
	return nil
}

type listingAuthorizerStub struct{ allowed bool }

func (a *listingAuthorizerStub) CanReviewListing(context.Context, int64, int64) (bool, error) {
	return a.allowed, nil
}

func (a *listingAuthorizerStub) CanPublishListing(context.Context, int64, int64) (bool, error) {
	return a.allowed, nil
}
