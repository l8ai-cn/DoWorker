package migrations

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConsolePublicationIntegrityMigrationContract(t *testing.T) {
	up, err := os.ReadFile("000005_console_publication_integrity.up.sql")
	require.NoError(t, err)
	sql := string(up)

	for _, fragment := range []string{
		"fk_marketplace_catalog_latest_version",
		"FOREIGN KEY (latest_version_id, id)",
		"prevent_catalog_version_payload_update",
		"prevent_submitted_listing_version_update",
		"NEW.review_status <> 'draft'",
		"civ.validation_status = 'passed'",
		"lv.review_status = 'approved'",
		"s.status = 'published'",
		"marketplace_space_publication_guard",
	} {
		require.Contains(t, sql, fragment)
	}
	require.NotContains(t, sql, "OLD.validation_status = 'passed' AND NEW.validation_status = 'deprecated'")
}

func TestConsolePublicationIntegrityDownDropsTriggersAndRestoresForeignKey(t *testing.T) {
	down, err := os.ReadFile("000005_console_publication_integrity.down.sql")
	require.NoError(t, err)
	sql := string(down)

	for _, fragment := range []string{
		"DROP TRIGGER IF EXISTS marketplace_catalog_version_immutable",
		"DROP TRIGGER IF EXISTS marketplace_listing_version_immutable",
		"DROP TRIGGER IF EXISTS marketplace_space_publication_guard",
		"FOREIGN KEY (latest_version_id)",
	} {
		require.Contains(t, sql, fragment)
	}
}
