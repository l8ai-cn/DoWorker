package migrations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigration000199WorkerSpecLinks(t *testing.T) {
	up, err := FS.ReadFile("000199_worker_spec_links.up.sql")
	require.NoError(t, err)
	upSQL := string(up)
	for _, fragment := range []string{
		"ADD COLUMN worker_spec_snapshot_id BIGINT",
		"UNIQUE (organization_id, id)",
		"FOREIGN KEY (organization_id, worker_spec_snapshot_id)",
		"REFERENCES worker_spec_snapshots (organization_id, id)",
		"ON DELETE RESTRICT",
		"CREATE INDEX idx_pods_worker_spec_snapshot_id",
		"WHERE worker_spec_snapshot_id IS NOT NULL",
	} {
		assert.Contains(t, upSQL, fragment)
	}

	down, err := FS.ReadFile("000199_worker_spec_links.down.sql")
	require.NoError(t, err)
	downSQL := string(down)
	for _, fragment := range []string{
		"DROP CONSTRAINT IF EXISTS pods_worker_spec_snapshot_org_fkey",
		"DROP INDEX IF EXISTS idx_pods_worker_spec_snapshot_id",
		"DROP COLUMN IF EXISTS worker_spec_snapshot_id",
		"DROP CONSTRAINT IF EXISTS worker_spec_snapshots_organization_id_id_key",
	} {
		assert.Contains(t, downSQL, fragment)
	}
	assert.Less(
		t,
		strings.Index(downSQL, "DROP CONSTRAINT IF EXISTS pods_worker_spec_snapshot_org_fkey"),
		strings.Index(downSQL, "DROP COLUMN IF EXISTS worker_spec_snapshot_id"),
	)
}
