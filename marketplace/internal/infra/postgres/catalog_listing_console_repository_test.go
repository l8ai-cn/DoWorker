package postgres

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/anthropics/agentsmesh/marketplace/internal/domain/listing"
	"github.com/anthropics/agentsmesh/marketplace/internal/service"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestCatalogAndListingConsoleRepositoriesPublishImmutableListing(t *testing.T) {
	dsn := os.Getenv("MARKETPLACE_POSTGRES_TEST_DSN")
	if dsn == "" {
		t.Skip("MARKETPLACE_POSTGRES_TEST_DSN is not configured")
	}
	sqlDB, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer func() { _ = sqlDB.Close() }()
	_, _ = sqlDB.Exec(`DROP SCHEMA IF EXISTS marketplace CASCADE`)
	t.Cleanup(func() { _, _ = sqlDB.Exec(`DROP SCHEMA IF EXISTS marketplace CASCADE`) })
	for _, migration := range []string{
		"000001_market_foundation.up.sql",
		"000002_core_catalog.up.sql",
		"000003_catalog_version_integrity.up.sql",
		"000004_console_revisions.up.sql",
		"000005_console_publication_integrity.up.sql",
		"000011_listing_taxonomy.up.sql",
	} {
		applyMigration(t, sqlDB, migration)
	}

	db, err := gorm.Open(gormpostgres.Open(dsn))
	require.NoError(t, err)
	ctx := context.Background()
	market, space := createConsoleMarketAndSpace(t, ctx, db)

	_, err = sqlDB.Exec(`
INSERT INTO marketplace.marketplace_publishers
  (id, slug, publisher_type, display_name, verification_status)
VALUES (31, 'commerce-lab', 'platform', 'Commerce Lab', 'verified')`)
	require.NoError(t, err)

	registration := service.NewCatalogRegistrationService(NewCatalogConsoleRepository(db))
	item, err := registration.RegisterItem(ctx, service.RegisterCatalogItemCommand{
		PublisherID: 31, Slug: "listing-optimizer", ResourceType: "application",
		Name: "商品优化应用", Summary: "开箱即用", PlatformResourceType: "expert",
		PlatformResourceID: 101, ActorUserID: 14,
	})
	require.NoError(t, err)
	version, err := registration.RegisterVersion(ctx, service.RegisterCatalogVersionCommand{
		CatalogItemID: item.CatalogItemID, Version: "1.0.0", SourceRevision: "git-sha",
		ContentDigest: strings.Repeat("a", 64), Manifest: []byte(`{"schema_version":"1"}`),
		Compatibility: []byte(`{"agents":["codex-cli"]}`),
		ActorUserID:   14,
	})
	require.NoError(t, err)
	_, err = registration.MarkVersionPassed(ctx, version.CatalogItemVersionID)
	require.NoError(t, err)

	publishing := service.NewListingPublishingService(
		NewListingConsoleRepository(db),
		allowAllListingReviews{},
	)
	draft, err := publishing.CreateDraft(ctx, service.CreateListingDraftCommand{
		MarketSlug: "commerce-market", CatalogItemVersionID: version.CatalogItemVersionID,
		Slug: "listing-optimizer", Visibility: listing.VisibilityPublic,
		AccessMode: listing.AccessModeDirect, DisplayName: "商品优化应用",
		Tagline: "批量优化商品信息", Description: "面向跨境电商运营团队",
		Outcomes: []byte(`["提升转化"]`), UseCases: []byte(`["批量优化"]`),
		TargetAudience: []byte(`["跨境运营"]`), Requirements: []byte(`[]`),
		TaxonomyTags: []service.ListingTaxonomyTag{
			{Slug: "cross-border-commerce", DisplayName: "跨境电商", Kind: "industry"},
		},
		ReleaseNotes: "首次发布",
		SpaceSlugs:   []string{space.Slug}, PrimarySpaceSlug: space.Slug, ActorUserID: 14,
	})
	require.NoError(t, err)
	_, err = sqlDB.Exec(`UPDATE marketplace.marketplace_listing_versions
SET review_status = 'submitted', description = 'changed' WHERE listing_id = $1`, draft.ListingID)
	require.Error(t, err)
	submitted, err := publishing.Submit(ctx, service.ListingCommand{
		MarketSlug: "commerce-market", ListingSlug: draft.Slug,
		ExpectedRevision: draft.Revision, ActorUserID: 14,
	})
	require.NoError(t, err)
	approved, err := publishing.Approve(ctx, service.ListingCommand{
		MarketSlug: "commerce-market", ListingSlug: draft.Slug,
		ExpectedRevision: submitted.Revision, ActorUserID: 21,
	})
	require.NoError(t, err)
	_, err = publishing.Publish(ctx, service.PublishListingCommand{
		MarketSlug: "commerce-market", ListingSlug: draft.Slug,
		ExpectedRevision: approved.Revision, ActorUserID: 21, PublishedAt: time.Now().UTC(),
	})
	require.NoError(t, err)

	_, err = sqlDB.Exec(`UPDATE marketplace.marketplace_catalog_item_versions
SET validation_status = 'deprecated' WHERE id = $1`, version.CatalogItemVersionID)
	require.Error(t, err)
	_, err = sqlDB.Exec(`UPDATE marketplace.marketplace_catalog_item_versions
SET manifest = '{"changed":true}' WHERE id = $1`, version.CatalogItemVersionID)
	require.Error(t, err)
	_, err = sqlDB.Exec(`UPDATE marketplace.marketplace_listing_versions
SET description = 'changed' WHERE listing_id = $1`, draft.ListingID)
	require.Error(t, err)
	_, err = sqlDB.Exec(`UPDATE marketplace.marketplace_spaces
SET status = 'hidden' WHERE id = $1`, space.SpaceID)
	require.Error(t, err)

	_, err = sqlDB.Exec(`UPDATE marketplace.marketplaces
SET status = 'published', visibility = 'public' WHERE id = $1`, market.MarketplaceID)
	require.NoError(t, err)
	storefront := NewStorefrontRepository(db)
	resolved, err := storefront.ResolveMarket(ctx, "commerce-market", "commerce-market.markets.example.com")
	require.NoError(t, err)
	items, err := storefront.ListPublishedListings(ctx, resolved.MarketplaceID, 20)
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, "listing-optimizer", items[0].Slug)
	require.Equal(t, []service.TaxonomyTagView{
		{Slug: "cross-border-commerce", DisplayName: "跨境电商", Kind: "industry"},
	}, items[0].Tags)
}

func createConsoleMarketAndSpace(
	t *testing.T,
	ctx context.Context,
	db *gorm.DB,
) (service.MarketResult, service.SpaceResult) {
	t.Helper()
	lifecycle, err := service.NewMarketLifecycleService(
		NewMarketConsoleRepository(db),
		"markets.example.com",
	)
	require.NoError(t, err)
	market, err := lifecycle.CreateMarket(ctx, service.CreateMarketCommand{
		Slug: "commerce-market", Name: "跨境电商市场", Summary: "开箱即用",
		OwnerOrgID: 9, ActorUserID: 14,
	})
	require.NoError(t, err)
	space, err := lifecycle.CreateSpace(ctx, service.CreateSpaceCommand{
		MarketSlug: "commerce-market", Slug: "operations", Name: "运营",
		Summary: "运营应用", ActorUserID: 14,
	})
	require.NoError(t, err)
	space, err = lifecycle.PublishSpace(ctx, service.PublishSpaceCommand{
		MarketSlug: "commerce-market", SpaceSlug: "operations",
		ExpectedRevision: space.Revision, PublishedAt: time.Now().UTC(),
	})
	require.NoError(t, err)
	return market, space
}

type allowAllListingReviews struct{}

func (allowAllListingReviews) CanReviewListing(context.Context, int64, int64) (bool, error) {
	return true, nil
}

func (allowAllListingReviews) CanPublishListing(context.Context, int64, int64) (bool, error) {
	return true, nil
}
