package migrations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigration000203ExpertWorkerSpecLink(t *testing.T) {
	up, err := FS.ReadFile("000203_expert_worker_spec_link.up.sql")
	require.NoError(t, err)
	upSQL := string(up)
	for _, fragment := range []string{
		"ALTER TABLE experts",
		"ADD COLUMN worker_spec_snapshot_id BIGINT",
		"FOREIGN KEY (organization_id, worker_spec_snapshot_id)",
		"REFERENCES worker_spec_snapshots (organization_id, id)",
		"ON DELETE RESTRICT",
		"CREATE INDEX idx_experts_worker_spec_snapshot_id",
		"WHERE worker_spec_snapshot_id IS NOT NULL",
	} {
		assert.Contains(t, upSQL, fragment)
	}

	down, err := FS.ReadFile("000203_expert_worker_spec_link.down.sql")
	require.NoError(t, err)
	downSQL := string(down)
	for _, fragment := range []string{
		"DROP CONSTRAINT IF EXISTS experts_worker_spec_snapshot_org_fkey",
		"DROP INDEX IF EXISTS idx_experts_worker_spec_snapshot_id",
		"DROP COLUMN IF EXISTS worker_spec_snapshot_id",
	} {
		assert.Contains(t, downSQL, fragment)
	}
	assert.Less(
		t,
		strings.Index(
			downSQL,
			"DROP CONSTRAINT IF EXISTS experts_worker_spec_snapshot_org_fkey",
		),
		strings.Index(downSQL, "DROP COLUMN IF EXISTS worker_spec_snapshot_id"),
	)
}
