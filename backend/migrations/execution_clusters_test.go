package migrations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMigration000206ExecutionClustersContract(t *testing.T) {
	up, err := FS.ReadFile("000206_execution_clusters.up.sql")
	require.NoError(t, err)
	upSQL := string(up)

	for _, fragment := range []string{
		"CREATE TABLE execution_clusters",
		"UNIQUE (organization_id, slug)",
		"UNIQUE (id, organization_id)",
		"execution_clusters_slug_check",
		"execution_clusters_kind_check",
		"execution_clusters_status_check",
		"INSERT INTO execution_clusters",
		"'online'",
		"'local'",
		"ADD COLUMN cluster_id BIGINT",
		"ADD COLUMN tunnel_state VARCHAR(32) NOT NULL DEFAULT 'disconnected'",
		"ADD COLUMN tunnel_last_seen_at TIMESTAMPTZ",
		"ADD COLUMN tunnel_last_error VARCHAR(255)",
		"runner_grpc_registration_tokens",
		"runner_pending_auths",
		"FOREIGN KEY (cluster_id, organization_id)",
		"runner_pending_auths_cluster_ownership_check",
		"UPDATE runner_pending_auths SET authorized = FALSE",
		"ALTER COLUMN authorized SET NOT NULL",
		"ALTER COLUMN cluster_id SET NOT NULL",
		"REFERENCES execution_clusters(id, organization_id)",
	} {
		require.Contains(t, upSQL, fragment)
	}
	require.NotContains(
		t,
		upSQL,
		"ALTER TABLE runner_pending_auths\n  ALTER COLUMN cluster_id SET NOT NULL",
	)

	down, err := FS.ReadFile("000206_execution_clusters.down.sql")
	require.NoError(t, err)
	downSQL := string(down)

	require.Contains(t, downSQL, "DROP TABLE IF EXISTS execution_clusters")
	require.Contains(t, downSQL, "DROP CONSTRAINT IF EXISTS runner_pending_auths_cluster_ownership_check")
	require.Contains(t, downSQL, "ALTER COLUMN authorized DROP NOT NULL")
	require.Less(
		t,
		strings.Index(downSQL, "DROP COLUMN IF EXISTS cluster_id"),
		strings.Index(downSQL, "DROP TABLE IF EXISTS execution_clusters"),
	)
}
