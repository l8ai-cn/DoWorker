package postgres

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestOrganizationApplicationsRepositoryFiltersInstallationStatuses(t *testing.T) {
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
		"000006_quota_foundation.up.sql",
		"000007_entitlement_installation.up.sql",
	} {
		applyMigration(t, sqlDB, migration)
	}
	seedInstallableListing(t, sqlDB)
	seedOrganizationApplications(t, sqlDB)

	db, err := gorm.Open(postgres.Open(dsn))
	require.NoError(t, err)
	items, err := NewOrganizationApplicationsRepository(db).ListOrganizationApplications(
		context.Background(),
		9,
	)

	require.NoError(t, err)
	require.Len(t, items, 2)
	require.Equal(t, "verifying", items[0].Status)
	require.Equal(t, "active", items[1].Status)
	require.Equal(t, "commerce-market", items[1].MarketSlug)
	require.Equal(t, "listing-optimizer", items[1].ListingSlug)
	require.Equal(t, "商品优化应用", items[1].DisplayName)
	require.Equal(t, "application", items[1].ResourceType)
	require.Equal(t, "expert-18", items[1].RuntimeRef)
}

func seedOrganizationApplications(t *testing.T, db *sql.DB) {
	t.Helper()
	seed := strings.ReplaceAll(`
INSERT INTO marketplace.marketplace_entitlements
  (id, marketplace_id, listing_id, subject_type, subject_platform_id,
   target_platform_org_id, status, source)
VALUES
  ('10000000-0000-4000-8000-000000000001', 1, 41, 'organization', 9, 9, 'active', 'direct'),
  ('10000000-0000-4000-8000-000000000002', 1, 41, 'organization', 9, 9, 'active', 'direct'),
  ('10000000-0000-4000-8000-000000000003', 1, 41, 'organization', 9, 9, 'active', 'direct'),
  ('10000000-0000-4000-8000-000000000004', 1, 41, 'organization', 10, 10, 'active', 'direct');
INSERT INTO marketplace.marketplace_installations
  (id, marketplace_id, listing_id, listing_version_id, entitlement_id,
   target_platform_org_id, quota_charge_scope, quota_account_id,
   installed_by_platform_user_id, status, runtime_ref, plan_digest, created_at)
VALUES
  ('20000000-0000-4000-8000-000000000001', 1, 41, 61,
   '10000000-0000-4000-8000-000000000001', 9, 'organization',
   'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa', 14, 'active', 'expert-18', '$PLAN_DIGEST$',
   TIMESTAMPTZ '2026-07-12 08:00:00+00'),
  ('20000000-0000-4000-8000-000000000002', 1, 41, 61,
   '10000000-0000-4000-8000-000000000002', 9, 'organization',
   'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa', 15, 'verifying', '', '$PLAN_DIGEST$',
   TIMESTAMPTZ '2026-07-12 09:00:00+00'),
  ('20000000-0000-4000-8000-000000000003', 1, 41, 61,
   '10000000-0000-4000-8000-000000000003', 9, 'organization',
   'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa', 14, 'failed', '', '$PLAN_DIGEST$',
   TIMESTAMPTZ '2026-07-12 10:00:00+00'),
  ('20000000-0000-4000-8000-000000000004', 1, 41, 61,
   '10000000-0000-4000-8000-000000000004', 10, 'organization',
   'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa', 14, 'active', 'expert-19', '$PLAN_DIGEST$',
   TIMESTAMPTZ '2026-07-12 11:00:00+00')
`, "$PLAN_DIGEST$", strings.Repeat("a", 64))
	_, err := db.Exec(seed)
	require.NoError(t, err)
}
