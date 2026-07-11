package migrations

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConsoleRevisionMigrationContract(t *testing.T) {
	up, err := os.ReadFile("000004_console_revisions.up.sql")
	require.NoError(t, err)
	sql := string(up)

	for _, table := range []string{
		"marketplaces",
		"marketplace_spaces",
		"marketplace_catalog_items",
		"marketplace_listings",
	} {
		require.Contains(t, sql, "ALTER TABLE marketplace."+table)
	}
	require.Equal(t, 4, strings.Count(sql, "ADD COLUMN revision BIGINT NOT NULL DEFAULT 1"))
}

func TestConsoleRevisionDownRemovesAllColumns(t *testing.T) {
	down, err := os.ReadFile("000004_console_revisions.down.sql")
	require.NoError(t, err)
	require.Equal(t, 4, strings.Count(string(down), "DROP COLUMN IF EXISTS revision"))
}
