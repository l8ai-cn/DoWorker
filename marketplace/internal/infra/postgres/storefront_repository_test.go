package postgres

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"

	"github.com/anthropics/agentsmesh/marketplace/internal/service"
	"github.com/anthropics/agentsmesh/marketplace/migrations"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestStorefrontRepositoryEnforcesHostAndPublication(t *testing.T) {
	dsn := os.Getenv("MARKETPLACE_POSTGRES_TEST_DSN")
	if dsn == "" {
		t.Skip("MARKETPLACE_POSTGRES_TEST_DSN is not configured")
	}
	sqlDB, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer sqlDB.Close()
	_, _ = sqlDB.Exec(`DROP SCHEMA IF EXISTS marketplace CASCADE`)
	t.Cleanup(func() {
		_, _ = sqlDB.Exec(`DROP SCHEMA IF EXISTS marketplace CASCADE`)
	})
	applyMigration(t, sqlDB, "000001_market_foundation.up.sql")
	applyMigration(t, sqlDB, "000002_core_catalog.up.sql")
	applyMigration(t, sqlDB, "000003_catalog_version_integrity.up.sql")
	seedStorefront(t, sqlDB)

	db, err := gorm.Open(postgres.Open(dsn))
	require.NoError(t, err)
	repository := NewStorefrontRepository(db)
	ctx := context.Background()

	_, err = repository.ResolveMarket(ctx, "commerce-market", "wrong.example.com")
	require.ErrorIs(t, err, service.ErrMarketNotFound)
	_, err = repository.ResolveMarket(ctx, "campus-market", "campus.example.com")
	require.ErrorIs(t, err, service.ErrMarketNotFound)

	market, err := repository.ResolveMarket(ctx, "commerce-market", "market.example.com")
	require.NoError(t, err)
	require.Equal(t, int64(1), market.MarketplaceID)

	aliasMarket, err := repository.ResolveMarket(ctx, "commerce-market", "market-alias.example.com")
	require.NoError(t, err)
	require.Equal(t, market.MarketplaceID, aliasMarket.MarketplaceID)

	items, err := repository.ListPublishedListings(ctx, market.MarketplaceID, 20)
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, "public-app", items[0].Slug)
	require.Equal(t, int64(71), items[0].ListingVersionID)
	require.Equal(t, int64(20_000_000), items[0].EstimatedCredits)

	detail, err := repository.GetPublishedListing(ctx, market.MarketplaceID, "public-app")
	require.NoError(t, err)
	require.Equal(t, "1.0.0", detail.Version)

	_, err = repository.GetPublishedListing(ctx, market.MarketplaceID, "hidden-app")
	require.True(t, errors.Is(err, service.ErrListingNotFound))

	_, err = sqlDB.Exec(`
INSERT INTO marketplace.marketplace_listing_versions
  (id, listing_id, catalog_item_id, catalog_item_version_id, revision,
   display_name, tagline, description, review_status)
VALUES (74, 61, 42, 52, 2, '错误拼接', '错误拼接', '错误拼接', 'approved')
`)
	require.Error(t, err)
}

func applyMigration(t *testing.T, db *sql.DB, name string) {
	t.Helper()
	body, err := migrations.FS.ReadFile(name)
	require.NoError(t, err)
	_, err = db.Exec(string(body))
	require.NoError(t, err)
}

func seedStorefront(t *testing.T, db *sql.DB) {
	t.Helper()
	_, err := db.Exec(storefrontSeedSQL)
	require.NoError(t, err)
}

const storefrontSeedSQL = `
INSERT INTO marketplace.marketplaces
  (id, slug, name, summary, status, visibility, owner_platform_org_id, created_by_platform_user_id)
VALUES
  (1, 'commerce-market', '跨境电商市场', '跨境电商应用', 'published', 'public', 9, 14),
  (2, 'campus-market', '高校市场', '高校应用', 'published', 'private', 10, 14);
INSERT INTO marketplace.marketplace_domains
  (marketplace_id, host, kind, status, verification_token, is_primary)
VALUES
  (1, 'market.example.com', 'custom', 'active', 'token-1', true),
  (1, 'market-alias.example.com', 'custom', 'active', 'token-1-alias', false),
  (2, 'campus.example.com', 'custom', 'active', 'token-2', true);
INSERT INTO marketplace.marketplace_spaces
  (id, marketplace_id, slug, name, summary, status, created_by_platform_user_id)
VALUES
  (11, 1, 'operations', '运营', '运营应用', 'published', 14),
  (21, 2, 'teaching', '教学', '教学应用', 'published', 14);
INSERT INTO marketplace.marketplace_publishers
  (id, slug, publisher_type, display_name, verification_status)
VALUES (31, 'commerce-lab', 'platform', 'Commerce Lab', 'verified');
INSERT INTO marketplace.marketplace_catalog_items
  (id, publisher_id, slug, resource_type, name, summary, platform_resource_type,
   platform_resource_id, status, created_by_platform_user_id)
VALUES
  (41, 31, 'public-app', 'application', '公开应用', '公开摘要', 'expert', 101, 'active', 14),
  (42, 31, 'hidden-app', 'application', '隐藏应用', '隐藏摘要', 'expert', 102, 'active', 14),
  (43, 31, 'other-app', 'application', '其他应用', '其他摘要', 'expert', 103, 'active', 14);
INSERT INTO marketplace.marketplace_catalog_item_versions
  (id, catalog_item_id, version, source_revision, content_digest, manifest,
   validation_status, created_by_platform_user_id)
VALUES
  (51, 41, '1.0.0', 'sha-1', repeat('a', 64),
   '{"installation_credits":"20"}', 'passed', 14),
  (52, 42, '1.0.0', 'sha-2', repeat('b', 64), '{}', 'passed', 14),
  (53, 43, '1.0.0', 'sha-3', repeat('c', 64), '{}', 'passed', 14);
INSERT INTO marketplace.marketplace_listings
  (id, marketplace_id, catalog_item_id, slug, status, visibility)
VALUES
  (61, 1, 41, 'public-app', 'approved', 'public'),
  (62, 1, 42, 'hidden-app', 'approved', 'hidden'),
  (63, 2, 43, 'other-app', 'approved', 'public');
INSERT INTO marketplace.marketplace_listing_versions
  (id, listing_id, catalog_item_id, catalog_item_version_id, revision, display_name, tagline,
   description, review_status)
VALUES
  (71, 61, 41, 51, 1, '公开应用', '提升商品运营效率', '公开详情', 'approved'),
  (72, 62, 42, 52, 1, '隐藏应用', '隐藏内容', '隐藏详情', 'approved'),
  (73, 63, 43, 53, 1, '其他应用', '其他内容', '其他详情', 'approved');
INSERT INTO marketplace.marketplace_listing_spaces
  (marketplace_id, listing_id, space_id, is_primary)
VALUES (1, 61, 11, true), (1, 62, 11, true), (2, 63, 21, true);
UPDATE marketplace.marketplace_listings
SET status = 'published', current_version_id = id + 10, published_at = NOW();
`
