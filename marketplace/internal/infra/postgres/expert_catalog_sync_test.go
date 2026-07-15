package postgres

import (
	"context"
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestExpertCatalogSynchronizerPublishesVersionsAndRemovesListing(t *testing.T) {
	dsn := os.Getenv("MARKETPLACE_POSTGRES_TEST_DSN")
	if dsn == "" {
		t.Skip("MARKETPLACE_POSTGRES_TEST_DSN is not configured")
	}
	sqlDB, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })
	resetExpertCatalogSyncSchema(t, sqlDB)
	t.Cleanup(func() { resetExpertCatalogSyncSchema(t, sqlDB) })

	for _, migration := range expertCatalogMigrations {
		applyMigration(t, sqlDB, migration)
	}
	_, err = sqlDB.Exec(expertCatalogSourceSchema)
	require.NoError(t, err)
	insertExpertRelease(t, sqlDB, 1, 1)

	db, err := gorm.Open(gormpostgres.Open(dsn))
	require.NoError(t, err)
	syncer := NewExpertCatalogSynchronizer(db)
	ctx := context.Background()

	count, err := syncer.Sync(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, count)
	first := loadExpertListingState(t, sqlDB)
	require.Equal(t, "published", first.Status)
	require.Equal(t, "public", first.Visibility)
	require.Equal(t, 2, first.ListingRevision)
	require.Equal(t, 1, first.ReleaseRevision)
	require.Equal(t, "1.0.0", first.Version)
	require.Equal(t, "expert-release-1", first.SourceRevision)
	require.Equal(t, 1, first.WorkerVersion)

	_, err = syncer.Sync(ctx)
	require.NoError(t, err)
	require.Equal(t, first, loadExpertListingState(t, sqlDB))

	insertExpertRelease(t, sqlDB, 2, 2)
	_, err = sqlDB.Exec(`
UPDATE expert_market_applications
SET latest_published_release_id = 2
WHERE id = 1`)
	require.NoError(t, err)
	_, err = syncer.Sync(ctx)
	require.NoError(t, err)
	second := loadExpertListingState(t, sqlDB)
	require.Equal(t, "published", second.Status)
	require.Equal(t, 3, second.ListingRevision)
	require.Equal(t, 2, second.ReleaseRevision)
	require.Equal(t, "2.0.0", second.Version)
	require.Equal(t, "expert-release-2", second.SourceRevision)

	var versionCount int
	err = sqlDB.QueryRow(`
SELECT COUNT(*)
FROM marketplace.marketplace_catalog_item_versions civ
JOIN marketplace.marketplace_catalog_items ci ON ci.id = civ.catalog_item_id
WHERE ci.platform_resource_type = 'expert' AND ci.platform_resource_id = 1
`).Scan(&versionCount)
	require.NoError(t, err)
	require.Equal(t, 2, versionCount)

	_, err = sqlDB.Exec(`
UPDATE expert_market_applications
SET latest_published_release_id = NULL
WHERE id = 1`)
	require.NoError(t, err)
	_, err = syncer.Sync(ctx)
	require.NoError(t, err)
	removed := loadExpertListingState(t, sqlDB)
	require.Equal(t, "removed", removed.Status)
	require.Equal(t, "hidden", removed.Visibility)
	require.Equal(t, 4, removed.ListingRevision)

	_, err = syncer.Sync(ctx)
	require.NoError(t, err)
	require.Equal(t, removed, loadExpertListingState(t, sqlDB))

	_, err = sqlDB.Exec(`
UPDATE expert_market_applications
SET latest_published_release_id = 2
WHERE id = 1`)
	require.NoError(t, err)
	_, err = syncer.Sync(ctx)
	require.NoError(t, err)
	republished := loadExpertListingState(t, sqlDB)
	require.Equal(t, "published", republished.Status)
	require.Equal(t, 5, republished.ListingRevision)

	_, err = sqlDB.Exec(`DELETE FROM expert_market_applications WHERE id = 1`)
	require.NoError(t, err)
	count, err = syncer.Sync(ctx)
	require.NoError(t, err)
	require.Zero(t, count)
	orphaned := loadExpertListingState(t, sqlDB)
	require.Equal(t, "removed", orphaned.Status)
	require.Equal(t, "hidden", orphaned.Visibility)
	require.Equal(t, 6, orphaned.ListingRevision)
}

type expertListingState struct {
	Status          string
	Visibility      string
	ListingRevision int
	ReleaseRevision int
	Version         string
	SourceRevision  string
	WorkerVersion   int
}

func loadExpertListingState(t *testing.T, db *sql.DB) expertListingState {
	t.Helper()
	var state expertListingState
	err := db.QueryRow(`
SELECT l.status, l.visibility, l.revision, lv.revision, civ.version,
  civ.source_revision,
  (civ.manifest->'runtime_snapshot'->'worker_spec'->>'version')::int
FROM marketplace.marketplace_listings l
JOIN marketplace.marketplace_listing_versions lv ON lv.id = l.current_version_id
JOIN marketplace.marketplace_catalog_item_versions civ
  ON civ.id = lv.catalog_item_version_id
WHERE l.slug = 'video-production-expert'
`).Scan(
		&state.Status, &state.Visibility, &state.ListingRevision,
		&state.ReleaseRevision, &state.Version, &state.SourceRevision,
		&state.WorkerVersion,
	)
	require.NoError(t, err)
	return state
}

func insertExpertRelease(t *testing.T, db *sql.DB, id, version int) {
	t.Helper()
	if id == 1 {
		_, err := db.Exec(`
INSERT INTO organizations (id, slug, name)
VALUES (1, 'dev-org', 'Dev Organization');
INSERT INTO expert_market_applications
  (id, slug, publisher_organization_id, publisher_user_id, is_operator_owned)
VALUES (1, 'video-production-expert', 1, 1, TRUE);
`)
		require.NoError(t, err)
	}
	_, err := db.Exec(`
INSERT INTO expert_market_releases
  (id, application_id, publisher_user_id, reviewer_user_id, version, name,
   summary, description, tags, outcomes, expert_snapshot,
   worker_spec_snapshot, featured, published_at)
VALUES ($1, 1, 1, 2, $2, '视频制作专家', '一体化视频制作',
  '负责从创意到成片交付。', ARRAY['短视频'], ARRAY['成片','质检报告'],
  '{"agent_slug":"video-studio","skill_slugs":["video-delivery-qa"]}',
  '{"version":1,"spec":{"version":1,"metadata":{"alias":"video-production-expert"}}}',
  TRUE, NOW())
`, id, version)
	require.NoError(t, err)
	if id == 1 {
		_, err = db.Exec(`
UPDATE expert_market_applications
SET latest_published_release_id = 1
WHERE id = 1`)
		require.NoError(t, err)
	}
}

func resetExpertCatalogSyncSchema(t *testing.T, db *sql.DB) {
	t.Helper()
	_, err := db.Exec(`
DROP SCHEMA IF EXISTS marketplace CASCADE;
DROP TABLE IF EXISTS expert_market_releases;
DROP TABLE IF EXISTS expert_market_applications;
DROP TABLE IF EXISTS organizations;
`)
	require.NoError(t, err)
}

var expertCatalogMigrations = []string{
	"000001_market_foundation.up.sql",
	"000002_core_catalog.up.sql",
	"000003_catalog_version_integrity.up.sql",
	"000004_console_revisions.up.sql",
	"000005_console_publication_integrity.up.sql",
	"000006_quota_foundation.up.sql",
	"000007_entitlement_installation.up.sql",
	"000008_quota_ledger_audit.up.sql",
	"000009_inline_runtime_snapshot.up.sql",
	"000010_default_marketplace_seed.up.sql",
	"000011_listing_taxonomy.up.sql",
	"000012_default_marketplace_taxonomy.up.sql",
	"000013_backend_expert_template_runtime_snapshot.up.sql",
	"000014_inline_expert_runtime_snapshot.up.sql",
	"000015_single_active_installation.up.sql",
}

const expertCatalogSourceSchema = `
CREATE TABLE organizations (
  id BIGINT PRIMARY KEY,
  slug TEXT NOT NULL,
  name TEXT NOT NULL
);
CREATE TABLE expert_market_applications (
  id BIGINT PRIMARY KEY,
  slug TEXT NOT NULL,
  publisher_organization_id BIGINT NOT NULL,
  publisher_user_id BIGINT NOT NULL,
  is_operator_owned BOOLEAN NOT NULL,
  latest_published_release_id BIGINT
);
CREATE TABLE expert_market_releases (
  id BIGINT PRIMARY KEY,
  application_id BIGINT NOT NULL,
  publisher_user_id BIGINT NOT NULL,
  reviewer_user_id BIGINT,
  version INTEGER NOT NULL,
  name TEXT NOT NULL,
  summary TEXT NOT NULL,
  description TEXT NOT NULL,
  tags TEXT[] NOT NULL,
  outcomes TEXT[] NOT NULL,
  expert_snapshot JSONB NOT NULL,
  worker_spec_snapshot JSONB NOT NULL,
  featured BOOLEAN NOT NULL,
  published_at TIMESTAMPTZ NOT NULL
);
`
