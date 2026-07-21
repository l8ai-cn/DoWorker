package postgres

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"

	"github.com/l8ai-cn/agentcloud/marketplace/internal/service"
	"github.com/l8ai-cn/agentcloud/marketplace/migrations"
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
	defer func() { _ = sqlDB.Close() }()
	_, _ = sqlDB.Exec(`DROP SCHEMA IF EXISTS marketplace CASCADE`)
	t.Cleanup(func() {
		_, _ = sqlDB.Exec(`DROP SCHEMA IF EXISTS marketplace CASCADE`)
	})
	applyMigration(t, sqlDB, "000001_market_foundation.up.sql")
	applyMigration(t, sqlDB, "000002_core_catalog.up.sql")
	applyMigration(t, sqlDB, "000003_catalog_version_integrity.up.sql")
	applyMigration(t, sqlDB, "000004_console_revisions.up.sql")
	applyMigration(t, sqlDB, "000005_console_publication_integrity.up.sql")
	applyMigration(t, sqlDB, "000011_listing_taxonomy.up.sql")
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

func TestStorefrontRepositoryFiltersTaxonomyTagsInPostgres(t *testing.T) {
	dsn := os.Getenv("MARKETPLACE_POSTGRES_TEST_DSN")
	if dsn == "" {
		t.Skip("MARKETPLACE_POSTGRES_TEST_DSN is not configured")
	}
	sqlDB, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer func() { _ = sqlDB.Close() }()
	_, _ = sqlDB.Exec(`DROP SCHEMA IF EXISTS marketplace CASCADE`)
	t.Cleanup(func() {
		_, _ = sqlDB.Exec(`DROP SCHEMA IF EXISTS marketplace CASCADE`)
	})
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
	seedStorefront(t, sqlDB)
	_, err = sqlDB.Exec(storefrontTaxonomySeedSQL)
	require.NoError(t, err)

	db, err := gorm.Open(postgres.Open(dsn))
	require.NoError(t, err)
	repository := NewStorefrontRepository(db)
	items, err := repository.ListPublishedListings(
		service.WithListingQuery(context.Background(), service.ListingQuery{
			Scene: "software-delivery", Industry: "enterprise-services",
			Capability: "code-review", Sort: "featured",
		}),
		1,
		20,
	)

	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, "public-app", items[0].Slug)
	require.Equal(t, []service.TaxonomyTagView{
		{Slug: "code-review", DisplayName: "代码评审", Kind: "capability"},
		{Slug: "enterprise-services", DisplayName: "企业服务", Kind: "industry"},
		{Slug: "software-delivery", DisplayName: "软件交付", Kind: "scene"},
	}, items[0].Tags)
}

func TestStorefrontRepositoryPaginatesWithSortCursor(t *testing.T) {
	dsn := os.Getenv("MARKETPLACE_POSTGRES_TEST_DSN")
	if dsn == "" {
		t.Skip("MARKETPLACE_POSTGRES_TEST_DSN is not configured")
	}
	sqlDB, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer func() { _ = sqlDB.Close() }()
	_, _ = sqlDB.Exec(`DROP SCHEMA IF EXISTS marketplace CASCADE`)
	t.Cleanup(func() {
		_, _ = sqlDB.Exec(`DROP SCHEMA IF EXISTS marketplace CASCADE`)
	})
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
	seedStorefront(t, sqlDB)
	_, err = sqlDB.Exec(storefrontTaxonomySeedSQL)
	require.NoError(t, err)

	db, err := gorm.Open(postgres.Open(dsn))
	require.NoError(t, err)
	repository := NewStorefrontRepository(db)
	firstPage, err := repository.ListPublishedListings(
		service.WithListingQuery(context.Background(), service.ListingQuery{Sort: "latest"}),
		1,
		1,
	)
	require.NoError(t, err)
	require.Len(t, firstPage, 1)
	require.Equal(t, "latest", firstPage[0].PageCursor.Sort)

	cursor, err := service.EncodeListingCursor(firstPage[0].PageCursor)
	require.NoError(t, err)
	next, err := service.DecodeListingCursor(cursor)
	require.NoError(t, err)
	secondPage, err := repository.ListPublishedListings(
		service.WithListingQuery(context.Background(), service.ListingQuery{
			Sort: "latest", Cursor: &next,
		}),
		1,
		1,
	)
	require.NoError(t, err)
	require.Len(t, secondPage, 1)
	require.NotEqual(t, firstPage[0].ListingID, secondPage[0].ListingID)

	detail, err := repository.GetPublishedListing(context.Background(), 1, firstPage[0].Slug)
	require.NoError(t, err)
	require.Empty(t, detail.PageCursor.Sort)
}

func TestListingTaxonomyMigrationPreservesLegacyTagsOnUpAndDown(t *testing.T) {
	dsn := os.Getenv("MARKETPLACE_POSTGRES_TEST_DSN")
	if dsn == "" {
		t.Skip("MARKETPLACE_POSTGRES_TEST_DSN is not configured")
	}
	sqlDB, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer func() { _ = sqlDB.Close() }()
	_, _ = sqlDB.Exec(`DROP SCHEMA IF EXISTS marketplace CASCADE`)
	t.Cleanup(func() {
		_, _ = sqlDB.Exec(`DROP SCHEMA IF EXISTS marketplace CASCADE`)
	})
	for _, migration := range []string{
		"000001_market_foundation.up.sql",
		"000002_core_catalog.up.sql",
		"000003_catalog_version_integrity.up.sql",
		"000004_console_revisions.up.sql",
		"000005_console_publication_integrity.up.sql",
	} {
		applyMigration(t, sqlDB, migration)
	}
	seedStorefront(t, sqlDB)
	_, err = sqlDB.Exec(`ALTER TABLE marketplace.marketplace_listing_versions
DISABLE TRIGGER marketplace_listing_version_immutable`)
	require.NoError(t, err)
	_, err = sqlDB.Exec(`UPDATE marketplace.marketplace_listing_versions
SET tags = ARRAY['delivery'] WHERE id = 71`)
	require.NoError(t, err)
	_, err = sqlDB.Exec(`ALTER TABLE marketplace.marketplace_listing_versions
ENABLE TRIGGER marketplace_listing_version_immutable`)
	require.NoError(t, err)

	applyMigration(t, sqlDB, "000011_listing_taxonomy.up.sql")
	requireListingTagTriggerContract(t, sqlDB)
	var columnExists bool
	var relationCount int
	err = sqlDB.QueryRow(`SELECT EXISTS (
  SELECT 1 FROM information_schema.columns
  WHERE table_schema = 'marketplace' AND table_name = 'marketplace_listing_versions' AND column_name = 'tags'
)`).Scan(&columnExists)
	require.NoError(t, err)
	err = sqlDB.QueryRow(`SELECT COUNT(*)
FROM marketplace.marketplace_listing_version_tags lvt
JOIN marketplace.marketplace_taxonomy_tags tags ON tags.id = lvt.taxonomy_tag_id
WHERE lvt.listing_version_id = 71 AND tags.display_name = 'delivery'`).Scan(&relationCount)
	require.NoError(t, err)
	require.True(t, columnExists)
	require.Equal(t, 1, relationCount)

	applyMigration(t, sqlDB, "000011_listing_taxonomy.down.sql")
	requireListingTagTriggerContract(t, sqlDB)
	var tags string
	err = sqlDB.QueryRow(`SELECT tags::text FROM marketplace.marketplace_listing_versions WHERE id = 71`).Scan(&tags)
	require.NoError(t, err)
	require.Equal(t, "{delivery}", tags)
}

func requireListingTagTriggerContract(t *testing.T, db *sql.DB) {
	t.Helper()
	var definition string
	err := db.QueryRow(`SELECT pg_get_functiondef(
  'marketplace.prevent_submitted_listing_version_update()'::regprocedure
)`).Scan(&definition)
	require.NoError(t, err)
	require.Contains(t, definition, "NEW.tags")
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

const storefrontTaxonomySeedSQL = `
INSERT INTO marketplace.marketplace_catalog_items
  (id, publisher_id, slug, resource_type, name, summary, platform_resource_type,
   platform_resource_id, status, created_by_platform_user_id)
VALUES (44, 31, 'course-builder', 'application', '课程构建专家', '课程内容构建', 'expert', 104, 'active', 14);
INSERT INTO marketplace.marketplace_catalog_item_versions
  (id, catalog_item_id, version, source_revision, content_digest, manifest,
   validation_status, created_by_platform_user_id)
VALUES (54, 44, '1.0.0', 'sha-4', repeat('d', 64), '{}', 'passed', 14);
INSERT INTO marketplace.marketplace_listings
  (id, marketplace_id, catalog_item_id, slug, status, visibility)
VALUES (64, 1, 44, 'course-builder', 'approved', 'public');
INSERT INTO marketplace.marketplace_listing_versions
  (id, listing_id, catalog_item_id, catalog_item_version_id, revision, display_name, tagline,
   description, review_status)
VALUES (74, 64, 44, 54, 1, '课程构建专家', '构建教学内容', '课程详情', 'approved');
INSERT INTO marketplace.marketplace_listing_spaces
  (marketplace_id, listing_id, space_id, is_primary)
VALUES (1, 64, 11, true);
UPDATE marketplace.marketplace_listings
SET status = 'published', current_version_id = 74, published_at = NOW()
WHERE id = 64;
INSERT INTO marketplace.marketplace_taxonomy_tags
  (id, marketplace_id, slug, display_name, kind)
VALUES
  (81, 1, 'software-delivery', '软件交付', 'scene'),
  (82, 1, 'enterprise-services', '企业服务', 'industry'),
  (83, 1, 'course-building', '课程建设', 'scene'),
  (84, 1, 'higher-education', '高等教育', 'industry'),
  (85, 1, 'code-review', '代码评审', 'capability');
INSERT INTO marketplace.marketplace_listing_version_tags
  (marketplace_id, listing_id, listing_version_id, taxonomy_tag_id)
VALUES
  (1, 61, 71, 81),
  (1, 61, 71, 82),
  (1, 61, 71, 85),
  (1, 64, 74, 83),
  (1, 64, 74, 84);
UPDATE marketplace.marketplace_listings
SET published_at = TIMESTAMPTZ '2026-07-12 12:00:00+00'
WHERE id = 61;
UPDATE marketplace.marketplace_listings
SET published_at = TIMESTAMPTZ '2026-07-12 11:00:00+00'
WHERE id = 64;
`
