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

func TestMigration000217GoalLoopResourceLinkPostgres(t *testing.T) {
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

	schema := fmt.Sprintf("goal_loop_link_%d", time.Now().UnixNano())
	require.NoError(t, execSQL(ctx, conn, `CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`))
	require.NoError(t, execSQL(ctx, conn, `CREATE SCHEMA `+schema))
	t.Cleanup(func() {
		_, _ = db.ExecContext(
			context.Background(),
			`DROP SCHEMA IF EXISTS `+schema+` CASCADE`,
		)
	})
	require.NoError(t, execSQL(ctx, conn, `SET search_path TO `+schema+`, public`))
	require.NoError(t, execOrchestrationMigrationSQL(
		ctx,
		conn,
		orchestrationResourceFixtures,
	))
	requireMigrationSQL(t, ctx, conn, "000202_add_goal_loops.up.sql")
	requireMigrationSQL(t, ctx, conn, "000211_orchestration_resources.up.sql")
	requireMigrationSQL(t, ctx, conn, "000212_orchestration_resource_integrity.up.sql")
	require.NoError(t, execSQL(ctx, conn, `
ALTER TABLE orchestration_resource_revisions
ADD CONSTRAINT orchestration_resource_revisions_org_revision_unique
UNIQUE (organization_id, resource_id, revision)`))

	resourceID, _ := insertOrchestrationResource(
		t,
		ctx,
		conn,
		1,
		"acme",
		"goal-loop-source",
		int64(100),
	)
	requireMigrationSQL(
		t,
		ctx,
		conn,
		"000217_orchestration_goal_loop_link.up.sql",
	)

	require.NoError(t, insertGoalLoopLink(
		ctx, conn, "legacy-loop", 100, nil, nil,
	))
	revision := int64(1)
	require.NoError(t, insertGoalLoopLink(
		ctx, conn, "resource-loop", 100, &resourceID, &revision,
	))
	requireExecError(t, ctx, conn, `
INSERT INTO goal_loops (
	organization_id, created_by_id, name, slug, worker_spec_snapshot_id,
	objective, acceptance_criteria, verification_command,
	orchestration_resource_id, orchestration_resource_revision
) VALUES (1, 10, 'Duplicate', 'duplicate-loop', 100, 'Fix', '["Done"]',
	'go test ./...', $1, 1)`, resourceID)
	requireExecError(t, ctx, conn, `
INSERT INTO goal_loops (
	organization_id, created_by_id, name, slug, worker_spec_snapshot_id,
	objective, acceptance_criteria, verification_command
) VALUES (1, 10, 'Cross org', 'cross-org-loop', 200, 'Fix', '["Done"]',
	'go test ./...')`)
	requireExecError(t, ctx, conn, `
INSERT INTO goal_loops (
	organization_id, created_by_id, name, slug, worker_spec_snapshot_id,
	objective, acceptance_criteria, verification_command
) VALUES (1, 10, 'Missing snapshot', 'missing-snapshot-loop', 999, 'Fix',
	'["Done"]', 'go test ./...')`)
	requireExecError(t, ctx, conn, `
INSERT INTO goal_loops (
	organization_id, created_by_id, name, slug, worker_spec_snapshot_id,
	objective, acceptance_criteria, verification_command,
	orchestration_resource_id
) VALUES (1, 10, 'Partial link', 'partial-link-loop', 100, 'Fix', '["Done"]',
	'go test ./...', $1)`, resourceID)

	requireMigrationSQL(
		t,
		ctx,
		conn,
		"000217_orchestration_goal_loop_link.down.sql",
	)
	var linkedColumns int
	require.NoError(t, conn.QueryRowContext(ctx, `
SELECT count(*)
FROM information_schema.columns
WHERE table_schema = current_schema()
  AND table_name = 'goal_loops'
  AND column_name IN (
	'orchestration_resource_id',
	'orchestration_resource_revision'
  )`).Scan(&linkedColumns))
	require.Zero(t, linkedColumns)
}

func insertGoalLoopLink(
	ctx context.Context,
	conn *sql.Conn,
	slug string,
	snapshotID int64,
	resourceID, revision *int64,
) error {
	_, err := conn.ExecContext(ctx, `
INSERT INTO goal_loops (
	organization_id, created_by_id, name, slug, worker_spec_snapshot_id,
	objective, acceptance_criteria, verification_command,
	orchestration_resource_id, orchestration_resource_revision
) VALUES (1, 10, $1, $1, $2, 'Fix checkout', '["Tests pass"]',
	'go test ./...', $3, $4)`,
		slug,
		snapshotID,
		resourceID,
		revision,
	)
	return err
}

func requireMigrationSQL(
	t *testing.T,
	ctx context.Context,
	conn *sql.Conn,
	name string,
) {
	t.Helper()
	content, err := FS.ReadFile(name)
	require.NoError(t, err)
	require.NoError(t, execOrchestrationMigrationSQL(ctx, conn, string(content)))
}
