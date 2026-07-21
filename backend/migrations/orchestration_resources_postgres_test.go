package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestMigration000211OrchestrationResourcesPostgres(t *testing.T) {
	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		var err error
		dsn, err = migrationPostgresDSN()
		require.NoError(t, err)
	}
	if dsn == "" {
		t.Skip("TEST_POSTGRES_DSN and MIGRATIONS_POSTGRES_TEST_DSN are not configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer db.Close()
	conn, err := db.Conn(ctx)
	require.NoError(t, err)
	defer conn.Close()

	schema := fmt.Sprintf("orchestration_resources_%d", time.Now().UnixNano())
	require.NoError(t, execSQL(ctx, conn, `CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`))
	require.NoError(t, execSQL(ctx, conn, `CREATE SCHEMA `+schema))
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `DROP SCHEMA IF EXISTS `+schema+` CASCADE`)
	})
	require.NoError(t, execSQL(ctx, conn, `SET search_path TO `+schema+`, public`))
	require.NoError(t, execOrchestrationMigrationSQL(ctx, conn, orchestrationResourceFixtures))

	up, err := FS.ReadFile("000211_orchestration_resources.up.sql")
	require.NoError(t, err)
	require.NoError(t, execOrchestrationMigrationSQL(ctx, conn, string(up)))
	integrityUp, err := FS.ReadFile("000212_orchestration_resource_integrity.up.sql")
	require.NoError(t, err)
	require.NoError(t, execOrchestrationMigrationSQL(ctx, conn, string(integrityUp)))

	resourceID, resourceUID := insertOrchestrationResource(
		t, ctx, conn, 1, "acme", "worker-one", int64(100),
	)
	createPlanID := insertOrchestrationPlan(t, ctx, conn, "create", 0, "", "worker-two")
	updatePlanID := insertOrchestrationPlan(t, ctx, conn, "update", resourceID, resourceUID, "worker-one")

	t.Run("scoped foreign keys and namespace", func(t *testing.T) {
		requireExecError(t, ctx, conn, `
INSERT INTO orchestration_resource_revisions
		(organization_id, resource_id, revision, generation, resource_version, canonical_manifest,
		 canonical_spec, resolved_refs, digest, actor_id)
	VALUES (2, $1, 2, 1, 2, '{}', '{}', '[]', $2, 10)`,
			resourceID, validOrchestrationDigest)
		requireExecError(t, ctx, conn, `
INSERT INTO orchestration_resource_revisions
		(organization_id, resource_id, revision, generation, resource_version, canonical_manifest,
		 canonical_spec, resolved_refs, digest, worker_spec_snapshot_id, actor_id)
	VALUES (1, $1, 2, 1, 2, '{}', '{}', '[]', $2, 200, 10)`,
			resourceID, validOrchestrationDigest)
		requireExecError(t, ctx, conn, `
INSERT INTO orchestration_resources
	(organization_id, api_version, kind, namespace, name, created_by_id, updated_by_id)
VALUES (1, 'agentcloud.io/v1alpha1', 'WorkerTemplate', 'other', 'worker-three', 10, 10)`)
		requireExecError(t, ctx, conn, `
	INSERT INTO orchestration_resource_plans
	(organization_id, actor_id, target_resource_id, target_api_version, target_kind,
		 target_namespace, target_name, operation, base_head_uid, base_resource_version,
		 draft_hash, plan_hash, canonical_manifest, resolved_refs, semantic_diff, issues,
		 artifact_kind, artifact_json, artifact_digest, options_revision, expires_at)
	VALUES (2, 10, $1, 'agentcloud.io/v1alpha1', 'WorkerTemplate', 'acme', 'worker-one',
		'update', $2, 1, $3, $3, '{}', '[]', '[]', '[]', 'WorkerSpec', '{}', $3,
		'runtime-catalog-1',
		now() + interval '5 minutes')`,
			resourceID, resourceUID, validOrchestrationDigest)
	})

	t.Run("json digest and identifier checks", func(t *testing.T) {
		requireExecError(t, ctx, conn, `
INSERT INTO orchestration_resources
	(organization_id, api_version, kind, namespace, name, labels, created_by_id, updated_by_id)
VALUES (1, 'agentcloud.io/v1alpha1', 'WorkerTemplate', 'acme', 'bad-labels', '[]', 10, 10)`)
		requireExecError(t, ctx, conn, `
INSERT INTO orchestration_resources
	(organization_id, api_version, kind, namespace, name, created_by_id, updated_by_id)
VALUES (1, 'agentcloud.io/v1alpha1', 'WorkerTemplate', 'Acme', 'bad-namespace', 10, 10)`)
		requireExecError(t, ctx, conn, `
INSERT INTO orchestration_resources
	(organization_id, api_version, kind, namespace, name, created_by_id, updated_by_id)
VALUES (1, 'agentcloud.io/v1alpha1', 'WorkerTemplate', 'acme', 'admin', 10, 10)`)
		requireExecError(t, ctx, conn, `
INSERT INTO orchestration_resource_plans
	(organization_id, actor_id, target_api_version, target_kind, target_namespace,
		 target_name, operation, draft_hash, plan_hash, canonical_manifest, resolved_refs,
		 semantic_diff, issues, artifact_kind, artifact_json, artifact_digest, options_revision, expires_at)
	VALUES (1, 10, 'agentcloud.io/v1alpha1', 'WorkerTemplate', 'acme', 'admin',
		'create', $1, $1, '{}', '[]', '[]', '[]', 'WorkerSpec', '{}', $1,
		'runtime-catalog-1',
		now() + interval '5 minutes')`, validOrchestrationDigest)
		requireExecError(t, ctx, conn, `
INSERT INTO orchestration_resource_revisions
		(organization_id, resource_id, revision, generation, resource_version, canonical_manifest,
		 canonical_spec, resolved_refs, digest, actor_id)
	VALUES (1, $1, 2, 1, 2, '{}', '{}', '{}', $2, 10)`,
			resourceID, validOrchestrationDigest)
		requireExecError(t, ctx, conn, `
INSERT INTO orchestration_resource_revisions
		(organization_id, resource_id, revision, generation, resource_version, canonical_manifest,
		 canonical_spec, resolved_refs, digest, actor_id)
	VALUES (1, $1, 2, 1, 2, '{}', '{}', '[]',
		'SHA256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa', 10)`,
			resourceID)
		requireExecError(t, ctx, conn, `
	INSERT INTO orchestration_resource_revisions
		(organization_id, resource_id, revision, generation, resource_version, canonical_manifest,
		 canonical_spec, resolved_refs, digest, actor_id)
	VALUES (1, $1, 2, 1, 0, '{}', '{}', '[]', $2, 10)`,
			resourceID, validOrchestrationDigest)
		requireExecError(t, ctx, conn, `
	INSERT INTO orchestration_resource_plans
		(organization_id, actor_id, target_api_version, target_kind, target_namespace,
		 target_name, operation, draft_hash, plan_hash, canonical_manifest, resolved_refs,
		 semantic_diff, issues, artifact_kind, artifact_json, artifact_digest,
		 options_revision, expires_at)
	VALUES (1, 10, 'agentcloud.io/v1alpha1', 'WorkerTemplate', 'acme', 'worker-bad-digest',
		'create', $1, $1, '{}', '[]', '[]', '[]', 'WorkerSpec', '{}', 'sha256:BAD',
		'runtime-catalog-1', now() + interval '5 minutes')`, validOrchestrationDigest)
		requireExecError(t, ctx, conn, `
	INSERT INTO orchestration_resource_plans
		(organization_id, actor_id, target_api_version, target_kind, target_namespace,
		 target_name, operation, draft_hash, plan_hash, canonical_manifest, resolved_refs,
		 semantic_diff, issues, artifact_kind, artifact_json, artifact_digest,
		 options_revision, expires_at)
	VALUES (1, 10, 'agentcloud.io/v1alpha1', 'WorkerTemplate', 'acme', 'worker-infinite',
		'create', $1, $1, '{}', '[]', '[]', '[]', 'WorkerSpec', '{}', $1,
		'runtime-catalog-1', 'infinity')`, validOrchestrationDigest)
		requireExecError(t, ctx, conn, `
	INSERT INTO orchestration_resource_plans
		(organization_id, actor_id, target_api_version, target_kind, target_namespace,
		 target_name, operation, draft_hash, plan_hash, canonical_manifest, resolved_refs,
		 semantic_diff, issues, artifact_kind, artifact_json, artifact_digest,
		 options_revision, expires_at)
	VALUES (1, 10, 'agentcloud.io/v1alpha1', 'WorkerTemplate', 'acme', 'worker-bad-options',
		'create', $1, $1, '{}', '[]', '[]', '[]', 'WorkerSpec', '{}', $1,
		' runtime-catalog-1', now() + interval '5 minutes')`, validOrchestrationDigest)
		requireExecError(t, ctx, conn, `
	INSERT INTO orchestration_resource_revisions
		(organization_id, resource_id, revision, generation, resource_version,
		 canonical_manifest, canonical_spec, resolved_refs, digest, actor_id, created_at)
	VALUES (1, $1, 2, 1, 2, '{}', '{}', '[]', $2, 10, 'infinity')`,
			resourceID, validOrchestrationDigest)
	})

	t.Run("head mutation boundary", func(t *testing.T) {
		require.NoError(t, execSQL(ctx, conn, `
	BEGIN;
	INSERT INTO orchestration_resource_revisions
		(organization_id, resource_id, revision, generation, resource_version, canonical_manifest,
		 canonical_spec, resolved_refs, digest, actor_id)
		VALUES (1, `+fmt.Sprint(resourceID)+`, 2, 1, 2, '{}', '{}', '[]',
		'`+validOrchestrationDigest+`', 11);
	UPDATE orchestration_resources
	SET display_name = 'Worker One', labels = '{"tier":"primary"}',
			status = '{"ready":true}', generation = 1, resource_version = 2,
		active_revision = 2, updated_by_id = 11, updated_at = updated_at + interval '1 second'
	WHERE id = `+fmt.Sprint(resourceID)+`;
	COMMIT;`))
		requireOrchestrationCommitError(t, ctx, conn, `
	INSERT INTO orchestration_resource_revisions
		(organization_id, resource_id, revision, generation, resource_version,
		 canonical_manifest, canonical_spec, resolved_refs, digest, actor_id)
	VALUES (1, `+fmt.Sprint(resourceID)+`, 3, 1, 99, '{}', '{}', '[]',
		'`+validOrchestrationDigest+`', 10);
	UPDATE orchestration_resources
	SET active_revision = 3, generation = 1, resource_version = 3,
		updated_by_id = 10, updated_at = updated_at + interval '1 second'
	WHERE id = `+fmt.Sprint(resourceID))
		requireOrchestrationCommitError(t, ctx, conn, `
	UPDATE orchestration_resources
	SET status = '{"ready":false}', resource_version = 3,
		updated_by_id = 10, updated_at = updated_at + interval '1 second'
	WHERE id = `+fmt.Sprint(resourceID)+`;
	INSERT INTO orchestration_resource_revisions
		(organization_id, resource_id, revision, generation, resource_version,
		 canonical_manifest, canonical_spec, resolved_refs, digest, actor_id)
	VALUES (1, `+fmt.Sprint(resourceID)+`, 3, 1, 3, '{}', '{}', '[]',
		'`+validOrchestrationDigest+`', 10);
	UPDATE orchestration_resources
	SET active_revision = 3, generation = 1, resource_version = 4,
		updated_by_id = 10, updated_at = updated_at + interval '1 second'
	WHERE id = `+fmt.Sprint(resourceID))
		requireExecError(t, ctx, conn, `
UPDATE orchestration_resources SET name = 'renamed' WHERE id = $1`, resourceID)
		requireExecError(t, ctx, conn, `
UPDATE orchestration_resources SET uid = uuid_generate_v4() WHERE id = $1`, resourceID)
		requireExecError(t, ctx, conn, `
UPDATE orchestration_resources SET created_by_id = 11 WHERE id = $1`, resourceID)
	})

	t.Run("revision immutability", func(t *testing.T) {
		requireExecError(t, ctx, conn, `
UPDATE orchestration_resource_revisions SET generation = 2 WHERE resource_id = $1`,
			resourceID)
		requireExecError(t, ctx, conn, `
DELETE FROM orchestration_resource_revisions WHERE resource_id = $1`, resourceID)
	})

	t.Run("plan payload and single consumption", func(t *testing.T) {
		requireExecError(t, ctx, conn, `
INSERT INTO orchestration_resource_plans
	(organization_id, actor_id, target_resource_id, target_api_version, target_kind,
	 target_namespace, target_name, operation, base_head_uid, draft_hash, plan_hash,
	 canonical_manifest, resolved_refs, semantic_diff, issues, artifact_kind,
	 artifact_json, artifact_digest, options_revision, expires_at)
VALUES (1, 10, $1, 'agentcloud.io/v1alpha1', 'WorkerTemplate', 'acme', 'worker-one',
	'update', $2, $3, $3, '{}', '[]', '[]', '[]', 'WorkerSpec', '{}', $3,
	'runtime-catalog-1',
	now() + interval '5 minutes')`,
			resourceID, resourceUID, validOrchestrationDigest)
		requireExecError(t, ctx, conn, `
INSERT INTO orchestration_resource_plans
	(organization_id, actor_id, target_resource_id, target_api_version, target_kind,
	 target_namespace, target_name, operation, base_head_uid, base_resource_version,
	 draft_hash, plan_hash, canonical_manifest, resolved_refs, semantic_diff, issues,
	 artifact_kind, artifact_json, artifact_digest, options_revision, expires_at, consumed_at,
	 consumed_by_id, consumption_result, result_resource_id, result_resource_uid,
	 result_resource_version, result_revision, result_json)
VALUES (1, 10, $1, 'agentcloud.io/v1alpha1', 'WorkerTemplate', 'acme', 'worker-one',
	'update', $2, 1, $3, $3, '{}', '[]', '[]', '[]', 'WorkerSpec', '{}', $3,
	'runtime-catalog-1',
	now() + interval '5 minutes', now(), 10, 'applied', $1, $2, 2, 1, '{}')`,
			resourceID, resourceUID, validOrchestrationDigest)
		requireExecError(t, ctx, conn, `
UPDATE orchestration_resource_plans SET options_revision = 'runtime-catalog-2' WHERE id = $1`,
			updatePlanID)
		requireExecError(t, ctx, conn, `
UPDATE orchestration_resource_plans SET consumed_at = now() WHERE id = $1`, updatePlanID)
		requireExecError(t, ctx, conn, `
UPDATE orchestration_resource_plans
SET consumed_at = now(), consumed_by_id = 10, consumption_result = 'applied',
	result_resource_id = $2, result_resource_uid = $3, result_resource_version = NULL,
	result_revision = 1, result_json = '{"created":false}'
WHERE id = $1`, updatePlanID, resourceID, resourceUID)
		requireExecError(t, ctx, conn, `
UPDATE orchestration_resource_plans
SET consumed_at = now(), consumed_by_id = 10, consumption_result = 'applied',
	result_resource_id = $2, result_resource_uid = $3, result_resource_version = 2,
	result_revision = NULL, result_json = '{"created":false}'
WHERE id = $1`, updatePlanID, resourceID, resourceUID)
		requireExecError(t, ctx, conn, `
UPDATE orchestration_resource_plans
SET consumed_at = now(), consumed_by_id = 10, consumption_result = 'applied',
	result_resource_id = $2, result_resource_uid = $3, result_resource_version = 3,
	result_revision = 3, result_json = '{"created":false}'
WHERE id = $1`, updatePlanID, resourceID, resourceUID)

		freshPlanID := insertOrchestrationPlan(
			t, ctx, conn, "update", resourceID, resourceUID, "worker-one",
		)
		fakeResultPlanID := insertOrchestrationPlan(
			t, ctx, conn, "update", resourceID, resourceUID, "worker-one",
		)
		requireOrchestrationCommitError(t, ctx, conn, `
	UPDATE orchestration_resource_plans
	SET consumed_at = now(), consumed_by_id = 10, consumption_result = 'applied',
		result_resource_id = $2, result_resource_uid = $3,
		result_resource_version = 99, result_revision = 2, result_json = '{}'
	WHERE id = $1`, fakeResultPlanID, resourceID, resourceUID)
		require.NoError(t, execSQL(ctx, conn, `
BEGIN;
UPDATE orchestration_resource_plans
SET consumed_at = now(), consumed_by_id = 10, consumption_result = 'applied',
	result_resource_id = `+fmt.Sprint(resourceID)+`, result_resource_uid = '`+resourceUID+`',
	result_resource_version = 3, result_revision = 3, result_json = '{"created":false}'
WHERE id = '`+freshPlanID+`';
INSERT INTO orchestration_resource_revisions
	(organization_id, resource_id, revision, generation, resource_version, canonical_manifest,
	 canonical_spec, resolved_refs, digest, actor_id)
VALUES (1, `+fmt.Sprint(resourceID)+`, 3, 1, 3, '{}', '{}', '[]',
	'`+validOrchestrationDigest+`', 10);
UPDATE orchestration_resources
SET active_revision = 3, generation = 1, resource_version = 3,
	updated_by_id = 10, updated_at = updated_at + interval '1 second'
WHERE id = `+fmt.Sprint(resourceID)+`;
COMMIT;`))
		requireExecError(t, ctx, conn, `
UPDATE orchestration_resource_plans SET consumption_result = 'cancelled' WHERE id = $1`,
			freshPlanID)
		requireExecError(t, ctx, conn, `
UPDATE orchestration_resource_plans
SET consumed_at = NULL, consumed_by_id = NULL, consumption_result = NULL,
	result_resource_id = NULL, result_resource_uid = NULL, result_resource_version = NULL,
	result_revision = NULL, result_json = NULL
WHERE id = $1`, freshPlanID)
		requireExecError(t, ctx, conn, `
UPDATE orchestration_resource_plans
SET consumed_at = now(), consumed_by_id = 11, consumption_result = 'cancelled',
	result_json = '{"reason":"cancelled"}'
WHERE id = $1`, createPlanID)
		requireExecError(t, ctx, conn, `
UPDATE orchestration_resource_plans
SET consumed_at = created_at - interval '1 second', consumed_by_id = 10,
	consumption_result = 'cancelled', result_json = '{"reason":"cancelled"}'
WHERE id = $1`, createPlanID)
		requireExecError(t, ctx, conn, `
UPDATE orchestration_resource_plans
SET consumed_at = expires_at, consumed_by_id = 10, consumption_result = 'cancelled',
	result_json = '{"reason":"cancelled"}'
WHERE id = $1`, createPlanID)
		require.NoError(t, execSQL(ctx, conn, `
UPDATE orchestration_resource_plans
SET consumed_at = now(), consumed_by_id = 10, consumption_result = 'cancelled',
	result_json = '{"reason":"cancelled"}'
WHERE id = '`+createPlanID+`'`))
		requireExecError(t, ctx, conn, `
DELETE FROM orchestration_resource_plans WHERE id = $1`, createPlanID)
	})

	t.Run("status-only head mutation", func(t *testing.T) {
		require.NoError(t, execSQL(ctx, conn, `
UPDATE orchestration_resources
SET status = '{"ready":false}', resource_version = resource_version + 1,
	updated_by_id = 10, updated_at = updated_at + interval '1 second'
WHERE id = `+fmt.Sprint(resourceID)))
		var activeRevision, generation, resourceVersion int64
		require.NoError(t, conn.QueryRowContext(ctx, `
SELECT active_revision, generation, resource_version
FROM orchestration_resources WHERE id = $1`, resourceID).Scan(
			&activeRevision, &generation, &resourceVersion,
		))
		require.Equal(t, int64(3), activeRevision)
		require.Equal(t, int64(1), generation)
		require.Equal(t, int64(4), resourceVersion)
	})

	t.Run("cascade and restrict semantics", func(t *testing.T) {
		requireExecError(t, ctx, conn, `
DELETE FROM worker_spec_snapshots WHERE id = 100`)
		requireExecError(t, ctx, conn, `
DELETE FROM orchestration_resources WHERE id = $1`, resourceID)
		orphanID, _ := insertOrchestrationResource(
			t, ctx, conn, 3, "empty-org", "orphan", nil,
		)
		require.NoError(t, execSQL(ctx, conn, `DELETE FROM organizations WHERE id = 3`))
		var count int
		require.NoError(t, conn.QueryRowContext(ctx,
			`SELECT count(*) FROM orchestration_resources WHERE id = $1`, orphanID).Scan(&count))
		require.Zero(t, count)
		require.NoError(t, execSQL(ctx, conn, `DELETE FROM organizations WHERE id = 1`))
		for _, table := range []string{
			"orchestration_resources",
			"orchestration_resource_revisions",
			"orchestration_resource_plans",
		} {
			require.NoError(t, conn.QueryRowContext(ctx,
				`SELECT count(*) FROM `+table+` WHERE organization_id = 1`).Scan(&count))
			require.Zero(t, count, table)
		}
	})

	integrityDown, err := FS.ReadFile("000212_orchestration_resource_integrity.down.sql")
	require.NoError(t, err)
	require.NoError(t, execOrchestrationMigrationSQL(ctx, conn, string(integrityDown)))
	down, err := FS.ReadFile("000211_orchestration_resources.down.sql")
	require.NoError(t, err)
	require.NoError(t, execOrchestrationMigrationSQL(ctx, conn, string(down)))
	var snapshots int
	require.NoError(t, conn.QueryRowContext(ctx,
		`SELECT count(*) FROM worker_spec_snapshots`).Scan(&snapshots))
	require.Equal(t, 2, snapshots)
}
