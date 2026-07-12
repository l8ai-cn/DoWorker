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
		"'do-worker-market'",
		"'market.l8ai.cn'",
		"'dowork.l8ai.cn'",
		"'software-delivery'",
		"'software-delivery-expert'",
		"'expert', NULL",
		`"runtime_snapshot"`,
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
