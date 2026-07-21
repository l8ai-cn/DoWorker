package migrations

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	validOrchestrationDigest      = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	orchestrationResourceFixtures = `
CREATE TABLE organizations (
	id BIGINT PRIMARY KEY,
	slug VARCHAR(100) NOT NULL UNIQUE
);
CREATE TABLE users (id BIGINT PRIMARY KEY);
CREATE TABLE worker_spec_snapshots (
	id BIGINT PRIMARY KEY,
	organization_id BIGINT NOT NULL,
	UNIQUE (organization_id, id)
);
INSERT INTO organizations (id, slug) VALUES (1, 'acme'), (2, 'other'), (3, 'empty-org');
INSERT INTO users (id) VALUES (10), (11);
INSERT INTO worker_spec_snapshots (id, organization_id) VALUES (100, 1), (200, 2);
`
)

func insertOrchestrationResource(
	t *testing.T,
	ctx context.Context,
	conn *sql.Conn,
	organizationID int,
	namespace, name string,
	snapshotID any,
) (int64, string) {
	t.Helper()
	tx, err := conn.BeginTx(ctx, nil)
	require.NoError(t, err)
	defer tx.Rollback()
	var id int64
	var uid string
	err = tx.QueryRowContext(ctx, `
INSERT INTO orchestration_resources
	(organization_id, api_version, kind, namespace, name, created_by_id, updated_by_id)
VALUES ($1, 'agentcloud.io/v1alpha1', 'WorkerTemplate', $2, $3, 10, 10)
RETURNING id, uid::text`, organizationID, namespace, name).Scan(&id, &uid)
	require.NoError(t, err)
	_, err = tx.ExecContext(ctx, `
INSERT INTO orchestration_resource_revisions
	(organization_id, resource_id, revision, generation, resource_version,
	 canonical_manifest, canonical_spec, resolved_refs, digest,
	 worker_spec_snapshot_id, actor_id)
VALUES ($1, $2, 1, 1, 1, '{}', '{}', '[]', $3, $4, 10)`,
		organizationID, id, validOrchestrationDigest, snapshotID)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	return id, uid
}

func insertOrchestrationPlan(
	t *testing.T,
	ctx context.Context,
	conn *sql.Conn,
	operation string,
	resourceID int64,
	resourceUID, name string,
) string {
	t.Helper()
	var targetResourceID any
	var baseUID any
	var baseVersion any
	if operation == "update" {
		targetResourceID = resourceID
		baseUID = resourceUID
		require.NoError(t, conn.QueryRowContext(ctx,
			`SELECT resource_version FROM orchestration_resources WHERE id = $1`,
			resourceID,
		).Scan(&baseVersion))
	}
	var planID string
	err := conn.QueryRowContext(ctx, `
INSERT INTO orchestration_resource_plans
	(organization_id, actor_id, target_resource_id, target_api_version, target_kind,
	 target_namespace, target_name, operation, base_head_uid, base_resource_version,
	 draft_hash, plan_hash, canonical_manifest, resolved_refs, semantic_diff, issues,
	 artifact_kind, artifact_json, artifact_digest, options_revision, expires_at)
VALUES (1, 10, $1, 'agentcloud.io/v1alpha1', 'WorkerTemplate', 'acme', $2,
	$3, $4, $5, $6, $6, '{}', '[]', '[]', '[]', 'WorkerSpec', '{}', $6,
	'runtime-catalog-1',
	now() + interval '5 minutes')
RETURNING id::text`,
		targetResourceID, name, operation, baseUID, baseVersion, validOrchestrationDigest,
	).Scan(&planID)
	require.NoError(t, err)
	return planID
}

func requireExecError(
	t *testing.T,
	ctx context.Context,
	conn *sql.Conn,
	query string,
	args ...any,
) {
	t.Helper()
	_, err := conn.ExecContext(ctx, query, args...)
	require.Error(t, err)
}

func requireOrchestrationCommitError(
	t *testing.T,
	ctx context.Context,
	conn *sql.Conn,
	query string,
	args ...any,
) {
	t.Helper()
	tx, err := conn.BeginTx(ctx, nil)
	require.NoError(t, err)
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, query, args...)
	require.NoError(t, err)
	require.Error(t, tx.Commit())
}

func execOrchestrationMigrationSQL(
	ctx context.Context,
	conn *sql.Conn,
	query string,
) error {
	_, err := conn.ExecContext(ctx, query)
	return err
}
