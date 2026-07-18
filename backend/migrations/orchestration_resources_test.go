package migrations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMigration000211OrchestrationResourcesContract(t *testing.T) {
	up, err := FS.ReadFile("000211_orchestration_resources.up.sql")
	require.NoError(t, err)
	integrityUp, err := FS.ReadFile("000212_orchestration_resource_integrity.up.sql")
	require.NoError(t, err)
	upSQL := string(up) + string(integrityUp)

	for _, fragment := range []string{
		"CREATE TABLE orchestration_resources",
		"CREATE TABLE orchestration_resource_revisions",
		"CREATE TABLE orchestration_resource_plans",
		"FOREIGN KEY (organization_id, namespace)",
		"REFERENCES worker_spec_snapshots (organization_id, id)",
		"FOREIGN KEY (id, active_revision)",
		"DEFERRABLE INITIALLY DEFERRED",
		"operation IN ('create', 'update')",
		"consumption_result IN ('applied', 'cancelled', 'expired')",
		"CREATE TRIGGER orchestration_resources_keep_identity",
		"CREATE TRIGGER orchestration_resource_revisions_immutable",
		"CREATE CONSTRAINT TRIGGER orchestration_resources_validate_active_revision",
		"CREATE CONSTRAINT TRIGGER orchestration_resource_revisions_validate_head",
		"CREATE TRIGGER orchestration_resource_plans_guard",
		"orchestration resource plans must be inserted pending",
		"CREATE INDEX idx_orchestration_resources_tenant_list",
		"CREATE INDEX idx_orchestration_resource_revisions_history",
		"CREATE INDEX idx_orchestration_resource_plans_expiry",
	} {
		require.Contains(t, upSQL, fragment)
	}
	require.NotContains(t, upSQL, "CREATE EXTENSION")
}

func TestMigration000211OrchestrationResourcesDownOrder(t *testing.T) {
	integrityDown, err := FS.ReadFile("000212_orchestration_resource_integrity.down.sql")
	require.NoError(t, err)
	down, err := FS.ReadFile("000211_orchestration_resources.down.sql")
	require.NoError(t, err)
	downSQL := string(integrityDown) + string(down)

	ordered := []string{
		"DROP TRIGGER IF EXISTS orchestration_resource_plans_guard",
		"DROP TRIGGER IF EXISTS orchestration_resource_revisions_validate_head",
		"DROP TRIGGER IF EXISTS orchestration_resources_validate_active_revision",
		"DROP TRIGGER IF EXISTS orchestration_resource_revisions_immutable",
		"DROP TRIGGER IF EXISTS orchestration_resources_keep_identity",
		"DROP FUNCTION IF EXISTS guard_orchestration_resource_plan",
		"DROP FUNCTION IF EXISTS validate_orchestration_resource_revision_link",
		"DROP FUNCTION IF EXISTS prevent_orchestration_resource_revision_mutation",
		"DROP FUNCTION IF EXISTS keep_orchestration_resource_identity",
		"DROP TABLE IF EXISTS orchestration_resource_plans",
		"DROP CONSTRAINT IF EXISTS orchestration_resources_active_revision_fkey",
		"DROP TABLE IF EXISTS orchestration_resource_revisions",
		"DROP TABLE IF EXISTS orchestration_resources",
		"DROP FUNCTION IF EXISTS orchestration_identifier_valid",
		"DROP CONSTRAINT IF EXISTS orchestration_organizations_id_slug_key",
	}
	previous := -1
	for _, fragment := range ordered {
		index := strings.Index(downSQL, fragment)
		require.Greater(t, index, previous, fragment)
		previous = index
	}
	require.NotContains(t, downSQL, "DROP TABLE IF EXISTS worker_spec_snapshots")
}
