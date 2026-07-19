package agentpod

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func ensureWorkerSpecSnapshotTable(t *testing.T, db *gorm.DB) {
	t.Helper()
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS worker_spec_snapshots (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		organization_id INTEGER NOT NULL,
		version INTEGER NOT NULL,
		spec_json BLOB NOT NULL,
		summary_json BLOB NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS worker_spec_dependency_artifacts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		organization_id INTEGER NOT NULL,
		worker_spec_snapshot_id INTEGER NOT NULL,
		artifact_json BLOB NOT NULL,
		artifact_digest TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`).Error)
}
