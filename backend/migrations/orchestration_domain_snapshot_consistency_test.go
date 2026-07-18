package migrations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMigration000226EnforcesRevisionSnapshotConsistency(t *testing.T) {
	up := readMigrationForTest(
		t,
		"000226_enforce_orchestration_domain_snapshot_consistency.up.sql",
	)
	for _, fragment := range []string{
		"BEGIN;",
		"RAISE EXCEPTION",
		"experts",
		"workflows",
		"workflow_runs",
		"goal_loops",
		"orchestration_worker_launches",
		"orchestration_resource_revisions_org_revision_snapshot_unique",
		"UNIQUE (organization_id, resource_id, revision, worker_spec_snapshot_id)",
		"FOREIGN KEY (",
		"organization_id,",
		"orchestration_resource_id,",
		"orchestration_resource_revision,",
		"worker_spec_snapshot_id",
		"resource_revision,",
		"DEFERRABLE INITIALLY DEFERRED",
		"COMMIT;",
	} {
		require.Contains(t, up, fragment)
	}
	require.NotContains(t, up, "MATCH FULL")
	require.GreaterOrEqual(t, strings.Count(
		up,
		"REFERENCES orchestration_resource_revisions",
	), 5)
}

func TestMigration000226RestoresRevisionOnlyForeignKeysOnRollback(t *testing.T) {
	down := readMigrationForTest(
		t,
		"000226_enforce_orchestration_domain_snapshot_consistency.down.sql",
	)
	for _, fragment := range []string{
		"experts_orchestration_revision_fkey",
		"workflows_orchestration_revision_fkey",
		"workflow_runs_orchestration_revision_fkey",
		"goal_loops_orchestration_revision_fkey",
		"orchestration_worker_launches_revision_fkey",
		"FOREIGN KEY (organization_id, orchestration_resource_id, orchestration_resource_revision)",
		"FOREIGN KEY (organization_id, resource_id, resource_revision)",
		"DROP CONSTRAINT IF EXISTS orchestration_resource_revisions_org_revision_snapshot_unique",
	} {
		require.Contains(t, down, fragment)
	}
	require.Less(t,
		strings.Index(down, "DROP CONSTRAINT IF EXISTS experts_orchestration_revision_fkey"),
		strings.Index(down, "DROP CONSTRAINT IF EXISTS orchestration_resource_revisions_org_revision_snapshot_unique"),
	)
}
