package migrations

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultMarketplaceSeedContract(t *testing.T) {
	up, err := os.ReadFile("000010_default_marketplace_seed.up.sql")
	require.NoError(t, err)
	down, err := os.ReadFile("000010_default_marketplace_seed.down.sql")
	require.NoError(t, err)

	for _, fragment := range []string{
		"'agent-cloud-market'",
		"'market.l8ai.cn'",
		"'dowork.l8ai.cn'",
		"'software-delivery'",
		"'software-delivery-expert'",
		"'expert', NULL",
		`"runtime_snapshot"`,
		`"market_application_slug":"software-delivery-expert"`,
		"'organization-starter'",
		"'d0000000-0000-4000-8000-000000000001'",
		"'default_marketplace_seed'",
	} {
		require.Contains(t, string(up), fragment)
	}
	require.Contains(t, string(down), "active installation history")
	require.NotContains(t, string(up), "REFERENCES public.")
	require.False(t, strings.Contains(string(up), "ON CONFLICT"))
}

func TestDefaultMarketplaceSeedTaxonomyContract(t *testing.T) {
	up, err := os.ReadFile("000012_default_marketplace_taxonomy.up.sql")
	require.NoError(t, err)
	down, err := os.ReadFile("000012_default_marketplace_taxonomy.down.sql")
	require.NoError(t, err)

	for _, fragment := range []string{
		"'software-delivery'",
		"'enterprise-services'",
		"'engineering-team'",
		"'runner-required'",
		"marketplace_listing_version_tags",
	} {
		require.Contains(t, string(up), fragment)
	}
	require.Contains(t, string(down), "marketplace_listing_version_tags")
}

func TestMarketplaceRuntimeSnapshotMigrationTargetsBackendExpertTemplate(t *testing.T) {
	up, err := os.ReadFile("000013_backend_expert_template_runtime_snapshot.up.sql")
	require.NoError(t, err)
	down, err := os.ReadFile("000013_backend_expert_template_runtime_snapshot.down.sql")
	require.NoError(t, err)

	require.Contains(t, string(up), "000010 creates the immutable 1.0.0 catalog version")
	require.Contains(t, string(down), "Catalog versions are immutable")
	require.NotContains(t, string(up), "UPDATE marketplace.marketplace_catalog_item_versions")
	require.NotContains(t, string(down), "UPDATE marketplace.marketplace_catalog_item_versions")
}

func TestMarketplaceInlineExpertSnapshotMigration(t *testing.T) {
	up, err := os.ReadFile("000014_inline_expert_runtime_snapshot.up.sql")
	require.NoError(t, err)
	down, err := os.ReadFile("000014_inline_expert_runtime_snapshot.down.sql")
	require.NoError(t, err)

	require.Contains(t, string(up), `"skill_slugs":["delivery-worktree","delivery-e2e","delivery-github-merge","delivery-gitlab-merge"]`)
	require.Contains(t, string(up), `"worker_spec":{"version":1`)
	require.Contains(t, string(up), `"interaction_mode":"acp"`)
	require.Contains(t, string(up), "inline-expert-v3")
	require.Contains(t, string(up), "'1.1.0'")
	require.Contains(t, string(up), "marketplace_expert_runtime_compatibility_guard")
	require.Contains(t, string(up), "published expert listing requires a valid compatible agent identifier")
	require.Contains(t, string(up), "~ '^[a-z0-9]+(-[a-z0-9]+)*$'")
	require.NotContains(t, string(up), "SET source_revision")
	require.Contains(t, string(down), "current_version_id = source_listing_version_id")
	require.NotContains(t, string(down), "DELETE FROM marketplace.marketplace_catalog_item_versions")
}
