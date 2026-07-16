package migrations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOrchestrationWorkerLaunchMigrationIsDurableAndTenantBound(t *testing.T) {
	up := readMigrationForTest(
		t,
		"000216_orchestration_worker_launches.up.sql",
	)
	for _, fragment := range []string{
		"CREATE TABLE orchestration_worker_launches",
		"UNIQUE (organization_id, plan_id)",
		"UNIQUE (organization_id, resource_id)",
		"REFERENCES orchestration_resource_plans (organization_id, id)",
		"REFERENCES orchestration_resource_revisions (",
		"resource_revision",
		"REFERENCES worker_spec_snapshots (organization_id, id)",
		"state IN ('pending', 'materializing', 'dispatched')",
		"claim_token UUID",
		"lease_expires_at TIMESTAMPTZ",
		"orchestration_worker_launch_id BIGINT",
		"idx_pods_orchestration_worker_launch",
		"REFERENCES orchestration_worker_launches (organization_id, id)",
		"REFERENCES pods (organization_id, id, pod_key)",
	} {
		require.Contains(t, up, fragment)
	}
	require.Contains(t, up, "DEFERRABLE INITIALLY DEFERRED")
}

func TestOrchestrationWorkerLaunchMigrationRollsBackInDependencyOrder(t *testing.T) {
	down := readMigrationForTest(
		t,
		"000216_orchestration_worker_launches.down.sql",
	)
	podLink := strings.Index(
		down,
		"DROP COLUMN IF EXISTS orchestration_worker_launch_id",
	)
	table := strings.Index(
		down,
		"DROP TABLE IF EXISTS orchestration_worker_launches",
	)
	planUnique := strings.Index(
		down,
		"DROP CONSTRAINT IF EXISTS orchestration_resource_plans_org_id_unique",
	)
	require.GreaterOrEqual(t, podLink, 0)
	require.Greater(t, table, podLink)
	require.Greater(t, planUnique, table)
}
