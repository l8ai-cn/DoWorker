package migrations

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMigration000227RequiresExecutionManifestForActiveResourceRuns(
	t *testing.T,
) {
	up := readMigrationForTest(
		t,
		"000227_workflow_run_execution_manifest.up.sql",
	)
	for _, fragment := range []string{
		"ADD COLUMN execution_manifest JSONB",
		"workflow_runs contain active runs without execution manifests",
		"WHERE execution_manifest IS NULL",
		"AND finished_at IS NULL",
		"workflow_runs_execution_manifest_check",
		"jsonb_typeof(execution_manifest) = 'object'",
		"orchestration_resource_id IS NOT NULL",
		"jsonb_typeof(execution_manifest -> 'autopilot') = 'object'",
		"execution_manifest ->> 'execution_mode' IN ('direct', 'autopilot')",
		"execution_manifest ->> 'sandbox_strategy' IN ('fresh', 'persistent')",
		"jsonb_typeof(execution_manifest -> 'session_persistence') = 'boolean'",
		"2147483647",
		"9223372036854775807",
		"finished_at IS NOT NULL",
	} {
		require.Contains(t, up, fragment)
	}
	require.NotContains(
		t,
		up,
		"WHERE orchestration_resource_id IS NOT NULL\n      AND execution_manifest IS NULL",
	)
	require.NotContains(t, up, "UPDATE workflow_runs")
}

func TestMigration000227DropsExecutionManifestOnRollback(t *testing.T) {
	down := readMigrationForTest(
		t,
		"000227_workflow_run_execution_manifest.down.sql",
	)
	require.Contains(
		t,
		down,
		"workflow_runs contain active resource-native runs; drain them before rollback",
	)
	require.Contains(
		t,
		down,
		"DROP CONSTRAINT IF EXISTS workflow_runs_execution_manifest_check",
	)
	require.Contains(
		t,
		down,
		"DROP COLUMN IF EXISTS execution_manifest",
	)
}
