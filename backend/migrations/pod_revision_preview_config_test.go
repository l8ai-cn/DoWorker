package migrations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMigration000204PodRevisionPreviewConfigContract(t *testing.T) {
	up, err := FS.ReadFile("000204_add_preview_config_to_pod_revisions.up.sql")
	require.NoError(t, err)
	upSQL := string(up)
	for _, fragment := range []string{
		"ADD COLUMN preview_port INTEGER NOT NULL DEFAULT 0",
		"ADD COLUMN preview_path VARCHAR(255) NOT NULL DEFAULT '/'",
		"UPDATE pod_config_revisions",
		"FROM pods",
		"preview_port = pods.preview_port",
		"NULLIF(pods.preview_path, '')",
		"pod_config_revisions_preview_port_check",
		"preview_port = 0 OR preview_port BETWEEN 1024 AND 65535",
		"pod_config_revisions_preview_path_check",
	} {
		require.Contains(t, upSQL, fragment)
	}

	down, err := FS.ReadFile("000204_add_preview_config_to_pod_revisions.down.sql")
	require.NoError(t, err)
	downSQL := string(down)
	require.Contains(t, downSQL, "DROP COLUMN IF EXISTS preview_path")
	require.Contains(t, downSQL, "DROP COLUMN IF EXISTS preview_port")
	require.NotContains(t, downSQL, "pods ")
	require.Less(
		t,
		strings.Index(downSQL, "DROP COLUMN IF EXISTS preview_path"),
		strings.Index(downSQL, "DROP COLUMN IF EXISTS preview_port"),
	)
}
