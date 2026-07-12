package migrations

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCoreCatalogMigrationContract(t *testing.T) {
	foundation, err := os.ReadFile("000001_market_foundation.up.sql")
	require.NoError(t, err)
	catalog, err := os.ReadFile("000002_core_catalog.up.sql")
	require.NoError(t, err)
	integrity, err := os.ReadFile("000003_catalog_version_integrity.up.sql")
	require.NoError(t, err)
	sql := string(foundation) + string(catalog) + string(integrity)

	requiredFragments := []string{
		"CREATE SCHEMA IF NOT EXISTS marketplace",
		"CREATE TABLE marketplace.marketplaces",
		"CREATE TABLE marketplace.marketplace_domains",
		"CREATE TABLE marketplace.marketplace_spaces",
		"CREATE TABLE marketplace.marketplace_publishers",
		"CREATE TABLE marketplace.marketplace_catalog_items",
		"CREATE TABLE marketplace.marketplace_catalog_item_versions",
		"CREATE TABLE marketplace.marketplace_listings",
		"CREATE TABLE marketplace.marketplace_listing_versions",
		"CREATE TABLE marketplace.marketplace_listing_spaces",
		"slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$'",
		"char_length(slug) BETWEEN 2 AND 100",
		"UNIQUE (marketplace_id, slug)",
		"UNIQUE (marketplace_id, catalog_item_id)",
		"idx_marketplace_listings_public",
		"CHECK (status <> 'published' OR (current_version_id IS NOT NULL",
		"FOREIGN KEY (id, current_version_id)",
		"REFERENCES marketplace.marketplace_listing_versions(listing_id, id)",
		"FOREIGN KEY (listing_id, catalog_item_id)",
		"FOREIGN KEY (catalog_item_version_id, catalog_item_id)",
		"FOREIGN KEY (marketplace_id, listing_id)",
		"FOREIGN KEY (marketplace_id, space_id)",
		"validate_listing_publication",
		"OLD.listing_id IS DISTINCT FROM NEW.listing_id",
		"assert_listing_is_publishable(OLD.listing_id)",
	}
	for _, fragment := range requiredFragments {
		require.Contains(t, sql, fragment)
	}

	require.NotContains(t, sql, "REFERENCES public.")
}

func TestCoreCatalogDownMigrationDropsDependenciesFirst(t *testing.T) {
	integrity, err := os.ReadFile("000003_catalog_version_integrity.down.sql")
	require.NoError(t, err)
	catalog, err := os.ReadFile("000002_core_catalog.down.sql")
	require.NoError(t, err)
	foundation, err := os.ReadFile("000001_market_foundation.down.sql")
	require.NoError(t, err)
	sql := string(integrity) + string(catalog) + string(foundation)

	listingSpaces := strings.Index(sql, "DROP TABLE IF EXISTS marketplace.marketplace_listing_spaces")
	listings := strings.Index(sql, "DROP TABLE IF EXISTS marketplace.marketplace_listings")
	markets := strings.Index(sql, "DROP TABLE IF EXISTS marketplace.marketplaces")
	schema := strings.Index(sql, "DROP SCHEMA IF EXISTS marketplace")

	require.NotEqual(t, -1, listingSpaces)
	require.NotEqual(t, -1, listings)
	require.NotEqual(t, -1, markets)
	require.NotEqual(t, -1, schema)
	require.Less(t, listingSpaces, listings)
	require.Less(t, listings, markets)
	require.Less(t, markets, schema)
}
