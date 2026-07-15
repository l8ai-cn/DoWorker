package migrations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOrchestrationGoalLoopLinkMigrationPinsResourceRevision(t *testing.T) {
	up := readMigrationForTest(
		t,
		"000217_orchestration_goal_loop_link.up.sql",
	)
	for _, fragment := range []string{
		"ALTER TABLE goal_loops",
		"orchestration_resource_id BIGINT",
		"orchestration_resource_revision BIGINT",
		"goal_loops_orchestration_mode_check",
		"goal_loops_worker_spec_snapshot_org_fkey",
		"FOREIGN KEY (organization_id, worker_spec_snapshot_id)",
		"goal_loops_orchestration_revision_fkey",
		"REFERENCES orchestration_resource_revisions",
		"ON DELETE RESTRICT",
		"DEFERRABLE INITIALLY DEFERRED",
		"CREATE UNIQUE INDEX idx_goal_loops_orchestration_resource",
		"WHERE orchestration_resource_id IS NOT NULL",
	} {
		require.Contains(t, up, fragment)
	}
	require.NotContains(t, up, "000215")
	require.NotContains(t, up, "000216")
}

func TestOrchestrationGoalLoopLinkMigrationRollsBackInDependencyOrder(
	t *testing.T,
) {
	down := readMigrationForTest(
		t,
		"000217_orchestration_goal_loop_link.down.sql",
	)
	index := strings.Index(
		down,
		"DROP INDEX IF EXISTS idx_goal_loops_orchestration_resource",
	)
	revisionFK := strings.Index(
		down,
		"DROP CONSTRAINT IF EXISTS goal_loops_orchestration_revision_fkey",
	)
	resourceColumn := strings.Index(
		down,
		"DROP COLUMN IF EXISTS orchestration_resource_id",
	)
	legacySnapshotFK := strings.Index(
		down,
		"ADD CONSTRAINT goal_loops_worker_spec_snapshot_id_fkey",
	)
	require.GreaterOrEqual(t, index, 0)
	require.Greater(t, revisionFK, index)
	require.Greater(t, resourceColumn, revisionFK)
	require.Greater(t, legacySnapshotFK, resourceColumn)
}
