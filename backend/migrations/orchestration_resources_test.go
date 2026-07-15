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
		"uid UUID NOT NULL DEFAULT uuid_generate_v4()",
		"api_version VARCHAR(64) NOT NULL",
		"CHECK (api_version = 'agentsmesh.io/v1alpha1')",
		"CHECK (kind ~ '^[A-Z][A-Za-z0-9]{1,99}$')",
		"CREATE FUNCTION orchestration_identifier_valid",
		"value NOT IN ('about','admin','agents'",
		"CHECK (orchestration_identifier_valid(namespace))",
		"CHECK (orchestration_identifier_valid(name))",
		"jsonb_typeof(labels) = 'object'",
		"jsonb_typeof(status) = 'object'",
		"UNIQUE (organization_id, api_version, kind, namespace, name)",
		"FOREIGN KEY (organization_id, namespace)",
		"REFERENCES organizations (id, slug)",
		"CREATE TABLE orchestration_resource_revisions",
		"FOREIGN KEY (organization_id, resource_id)",
		"REFERENCES orchestration_resources (organization_id, id)",
		"ON DELETE CASCADE",
		"FOREIGN KEY (organization_id, worker_spec_snapshot_id)",
		"REFERENCES worker_spec_snapshots (organization_id, id)",
		"UNIQUE (resource_id, revision)",
		"resource_version BIGINT NOT NULL",
		"FOREIGN KEY (id, active_revision)",
		"DEFERRABLE INITIALLY DEFERRED",
		"jsonb_typeof(canonical_manifest) = 'object'",
		"jsonb_typeof(canonical_spec) = 'object'",
		"jsonb_typeof(resolved_refs) = 'array'",
		"~ '^sha256:[0-9a-f]{64}$'",
		"CREATE TABLE orchestration_resource_plans",
		"operation IN ('create', 'update')",
		"base_resource_version IS NOT NULL",
		"consumption_result IN ('applied', 'cancelled', 'expired')",
		"result_resource_version IS NOT NULL",
		"result_revision IS NOT NULL",
		"consumed_at < expires_at",
		"consumed_at >= expires_at",
		"jsonb_typeof(semantic_diff) = 'array'",
		"jsonb_typeof(issues) = 'array'",
		"jsonb_typeof(artifact_json) = 'object'",
		"artifact_digest VARCHAR(71) NOT NULL",
		"options_revision VARCHAR(128) NOT NULL",
		"options_revision = btrim(options_revision)",
		"consumed_at >= created_at",
		"expires_at > created_at",
		"FOREIGN KEY (organization_id, target_resource_id, base_head_uid",
		"FOREIGN KEY (organization_id, result_resource_id, result_resource_uid",
		"CREATE TRIGGER orchestration_resources_keep_identity",
		"CREATE TRIGGER orchestration_resource_revisions_immutable",
		"BEFORE UPDATE OR DELETE ON orchestration_resource_revisions",
		"CREATE CONSTRAINT TRIGGER orchestration_resources_validate_active_revision",
		"CREATE CONSTRAINT TRIGGER orchestration_resource_revisions_validate_head",
		"validate_orchestration_resource_revision_link",
		"CREATE TRIGGER orchestration_resource_plans_guard",
		"BEFORE INSERT OR UPDATE OR DELETE ON orchestration_resource_plans",
		"orchestration resource plans must be inserted pending",
		"CREATE INDEX idx_orchestration_resources_tenant_list",
		"CREATE INDEX idx_orchestration_resources_tenant_head",
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
