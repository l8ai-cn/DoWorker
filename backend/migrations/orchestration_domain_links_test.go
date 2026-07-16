package migrations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOrchestrationDomainLinksMigrationPinsResourceRevisions(t *testing.T) {
	up := readMigrationForTest(t, "000215_orchestration_domain_links.up.sql")
	for _, fragment := range []string{
		"UNIQUE (organization_id, resource_id, revision)",
		"ALTER TABLE experts",
		"experts_orchestration_revision_fkey",
		"idx_experts_orchestration_resource",
		"ALTER TABLE workflows",
		"workflows_worker_spec_snapshot_org_fkey",
		"workflows_orchestration_revision_fkey",
		"idx_workflows_orchestration_resource",
		"ALTER TABLE workflow_runs",
		"workflow_runs_worker_spec_snapshot_org_fkey",
		"workflow_runs_orchestration_revision_fkey",
	} {
		require.Contains(t, up, fragment)
	}
	require.GreaterOrEqual(t, strings.Count(
		up,
		"DEFERRABLE INITIALLY DEFERRED",
	), 3)
}

func TestOrchestrationDomainLinksMigrationRollsBackInDependencyOrder(t *testing.T) {
	down := readMigrationForTest(t, "000215_orchestration_domain_links.down.sql")
	require.Less(t,
		strings.Index(down, "DROP CONSTRAINT IF EXISTS workflow_runs_orchestration_revision_fkey"),
		strings.Index(down, "DROP CONSTRAINT IF EXISTS orchestration_resource_revisions_org_revision_unique"),
	)
	require.Contains(t, down, "DROP COLUMN IF EXISTS orchestration_resource_id")
	require.Contains(t, down, "DROP COLUMN IF EXISTS worker_spec_snapshot_id")
}

func readMigrationForTest(t *testing.T, name string) string {
	t.Helper()
	content, err := FS.ReadFile(name)
	require.NoError(t, err)
	return string(content)
}
