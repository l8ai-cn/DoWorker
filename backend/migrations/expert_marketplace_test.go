package migrations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMigration000208ExpertMarketplaceContract(t *testing.T) {
	up, err := FS.ReadFile("000208_expert_marketplace.up.sql")
	require.NoError(t, err)
	upSQL := string(up)

	for _, fragment := range []string{
		"CREATE TABLE expert_market_applications",
		"CONSTRAINT expert_market_applications_slug_unique UNIQUE (slug)",
		"expert_market_applications_slug_check",
		"CREATE TABLE expert_market_releases",
		"UNIQUE (application_id, version)",
		"UNIQUE (application_id, id)",
		"status IN ('draft', 'pending_review', 'published', 'rejected', 'withdrawn')",
		"jsonb_typeof(expert_snapshot) = 'object'",
		"jsonb_typeof(worker_spec_snapshot) = 'object'",
		"jsonb_typeof(skill_dependencies) = 'array'",
		"CREATE FUNCTION prevent_expert_market_release_immutable_update",
		"CREATE TRIGGER expert_market_releases_immutable",
		"NEW.id IS DISTINCT FROM OLD.id",
		"NEW.created_at IS DISTINCT FROM OLD.created_at",
		"FOREIGN KEY (id, latest_published_release_id)",
		"REFERENCES expert_market_releases(application_id, id)",
		"ADD COLUMN source_market_application_id",
		"ADD COLUMN source_market_release_id",
		"FOREIGN KEY (source_market_application_id, source_market_release_id)",
		"ON DELETE SET NULL",
		"CREATE UNIQUE INDEX idx_experts_org_market_application",
		"WHERE source_market_application_id IS NOT NULL",
	} {
		require.Contains(t, upSQL, fragment)
	}
	require.NotContains(t, upSQL, "BEFORE DELETE")
	require.NotContains(t, upSQL, "organization_id, source_market_application_id, source_market_release_id")

	down, err := FS.ReadFile("000208_expert_marketplace.down.sql")
	require.NoError(t, err)
	downSQL := string(down)
	for _, fragment := range []string{
		"DROP CONSTRAINT IF EXISTS experts_market_release_fkey",
		"DROP COLUMN IF EXISTS source_market_release_id",
		"DROP COLUMN IF EXISTS source_market_application_id",
		"DROP CONSTRAINT IF EXISTS expert_market_applications_latest_release_fkey",
		"DROP TRIGGER IF EXISTS expert_market_releases_immutable",
		"DROP FUNCTION IF EXISTS prevent_expert_market_release_immutable_update",
		"DROP TABLE IF EXISTS expert_market_releases",
		"DROP TABLE IF EXISTS expert_market_applications",
	} {
		require.Contains(t, downSQL, fragment)
	}
	require.Less(t,
		strings.Index(downSQL, "DROP TABLE IF EXISTS expert_market_releases"),
		strings.Index(downSQL, "DROP TABLE IF EXISTS expert_market_applications"),
	)
}
