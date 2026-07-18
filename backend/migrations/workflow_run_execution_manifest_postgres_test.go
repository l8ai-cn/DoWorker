package migrations

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMigration000220WorkflowRunExecutionManifestPostgres(t *testing.T) {
	ctx, conn := openSnapshotConsistencyPostgres(t)
	requireSnapshotConsistencySchema(t, ctx, conn, false)
	requireSnapshotMigration(
		t,
		ctx,
		conn,
		"000219_enforce_orchestration_domain_snapshot_consistency.up.sql",
	)
	requireWorkflowRunFinishedAtColumn(t, ctx, conn)
	require.NoError(t, execMigrationSQL(
		ctx,
		conn,
		`UPDATE workflow_runs SET finished_at = NOW()`,
	))
	requireSnapshotMigration(
		t,
		ctx,
		conn,
		"000220_workflow_run_execution_manifest.up.sql",
	)

	require.NoError(t, insertResourceWorkflowRun(
		ctx,
		conn,
		30,
		workflowRunManifestJSON,
		false,
	))
	for id, field := range []struct {
		name  string
		value any
	}{
		{name: "timeout_minutes", value: "bad"},
		{name: "timeout_minutes", value: 2147483648},
		{name: "idle_timeout_seconds", value: -1},
		{name: "session_persistence", value: true},
	} {
		require.Error(t, insertResourceWorkflowRun(
			ctx,
			conn,
			int64(40+id),
			workflowRunManifestWithField(t, field.name, field.value),
			false,
		))
	}
	require.Error(t, insertResourceWorkflowRun(
		ctx,
		conn,
		49,
		`{"version":1}`,
		false,
	))
	require.Error(t, insertResourceWorkflowRun(ctx, conn, 31, "", false))
	require.NoError(t, insertResourceWorkflowRun(ctx, conn, 32, "", true))
	require.Error(t, insertLegacyWorkflowRunWithManifest(
		ctx,
		conn,
		33,
		workflowRunManifestJSON,
	))

	down, err := FS.ReadFile("000220_workflow_run_execution_manifest.down.sql")
	require.NoError(t, err)
	require.Error(t, execMigrationSQL(ctx, conn, string(down)))
	require.NoError(t, execMigrationSQL(ctx, conn, "ROLLBACK"))
	require.True(t, postgresColumnExists(
		ctx,
		t,
		conn,
		"workflow_runs",
		"execution_manifest",
	))
	require.NoError(t, execMigrationSQL(
		ctx,
		conn,
		`UPDATE workflow_runs SET finished_at = NOW() WHERE finished_at IS NULL`,
	))
	require.NoError(t, execMigrationSQL(ctx, conn, string(down)))
	require.False(t, postgresColumnExists(
		ctx,
		t,
		conn,
		"workflow_runs",
		"execution_manifest",
	))
}

func TestMigration000220ActiveLegacyRunFailsClosedPostgres(t *testing.T) {
	ctx, conn := openSnapshotConsistencyPostgres(t)
	requireSnapshotConsistencySchema(t, ctx, conn, false)
	requireSnapshotMigration(
		t,
		ctx,
		conn,
		"000219_enforce_orchestration_domain_snapshot_consistency.up.sql",
	)
	requireWorkflowRunFinishedAtColumn(t, ctx, conn)
	require.NoError(t, execMigrationSQL(
		ctx,
		conn,
		`UPDATE workflow_runs SET finished_at = NOW()`,
	))
	require.NoError(t, insertLegacyWorkflowRun(ctx, conn, 60))

	up, err := FS.ReadFile("000220_workflow_run_execution_manifest.up.sql")
	require.NoError(t, err)
	require.Error(t, execMigrationSQL(ctx, conn, string(up)))
	require.NoError(t, execMigrationSQL(ctx, conn, "ROLLBACK"))
	require.False(t, postgresColumnExists(
		ctx,
		t,
		conn,
		"workflow_runs",
		"execution_manifest",
	))
}

func insertLegacyWorkflowRunWithManifest(
	ctx context.Context,
	conn *sql.Conn,
	id int64,
	manifest string,
) error {
	_, err := conn.ExecContext(ctx, `
INSERT INTO workflow_runs (
	id,
	organization_id,
	worker_spec_snapshot_id,
	execution_manifest
)
VALUES ($1, 1, 100, $2::jsonb)`,
		id,
		manifest,
	)
	return err
}

func insertLegacyWorkflowRun(
	ctx context.Context,
	conn *sql.Conn,
	id int64,
) error {
	_, err := conn.ExecContext(ctx, `
INSERT INTO workflow_runs (
	id,
	organization_id,
	worker_spec_snapshot_id
)
VALUES ($1, 1, 100)`,
		id,
	)
	return err
}

func TestMigration000220ActiveRunFailsClosedPostgres(t *testing.T) {
	ctx, conn := openSnapshotConsistencyPostgres(t)
	requireSnapshotConsistencySchema(t, ctx, conn, false)
	requireSnapshotMigration(
		t,
		ctx,
		conn,
		"000219_enforce_orchestration_domain_snapshot_consistency.up.sql",
	)
	requireWorkflowRunFinishedAtColumn(t, ctx, conn)

	up, err := FS.ReadFile("000220_workflow_run_execution_manifest.up.sql")
	require.NoError(t, err)
	require.Error(t, execMigrationSQL(ctx, conn, string(up)))
	require.NoError(t, execMigrationSQL(ctx, conn, "ROLLBACK"))
	require.False(t, postgresColumnExists(
		ctx,
		t,
		conn,
		"workflow_runs",
		"execution_manifest",
	))
}

func requireWorkflowRunFinishedAtColumn(
	t *testing.T,
	ctx context.Context,
	conn *sql.Conn,
) {
	t.Helper()
	require.NoError(t, execMigrationSQL(
		ctx,
		conn,
		`ALTER TABLE workflow_runs ADD COLUMN finished_at TIMESTAMPTZ`,
	))
}

func insertResourceWorkflowRun(
	ctx context.Context,
	conn *sql.Conn,
	id int64,
	manifest string,
	finished bool,
) error {
	var manifestValue any
	if manifest != "" {
		manifestValue = manifest
	}
	var finishedAt any
	if finished {
		finishedAt = "2026-07-16T00:00:00Z"
	}
	_, err := conn.ExecContext(ctx, `
INSERT INTO workflow_runs (
	id,
	organization_id,
	worker_spec_snapshot_id,
	orchestration_resource_id,
	orchestration_resource_revision,
	execution_manifest,
	finished_at
)
VALUES ($1, 1, 100, 7, 1, $2::jsonb, $3::timestamptz)`,
		id,
		manifestValue,
		finishedAt,
	)
	return err
}

const workflowRunManifestJSON = `{
	"version": 1,
	"organization_id": 1,
	"workflow_name": "Nightly",
	"workflow_slug": "nightly",
	"created_by_id": 1,
	"execution_mode": "direct",
	"autopilot": {},
	"sandbox_strategy": "fresh",
	"session_persistence": false,
	"max_retained_runs": 30,
	"timeout_minutes": 60,
	"idle_timeout_seconds": 30
}`

func workflowRunManifestWithField(
	t *testing.T,
	field string,
	value any,
) string {
	t.Helper()
	var manifest map[string]any
	require.NoError(t, json.Unmarshal(
		[]byte(workflowRunManifestJSON),
		&manifest,
	))
	manifest[field] = value
	content, err := json.Marshal(manifest)
	require.NoError(t, err)
	return string(content)
}
