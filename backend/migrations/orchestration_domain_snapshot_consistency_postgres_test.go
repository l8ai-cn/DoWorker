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

func TestMigration000219DomainSnapshotConsistencyPostgres(t *testing.T) {
	ctx, conn := openSnapshotConsistencyPostgres(t)
	requireSnapshotConsistencySchema(t, ctx, conn, false)
	requireSnapshotMigration(t, ctx, conn, "000219_enforce_orchestration_domain_snapshot_consistency.up.sql")

	require.NoError(t, insertSnapshotOnlyExpert(ctx, conn, 1000, 101))
	require.NoError(t, insertSnapshotOnlyGoalLoop(ctx, conn, 1001, 101))

	for _, tc := range []struct {
		name  string
		query string
	}{
		{
			name: "experts",
			query: `
INSERT INTO experts
	(id, organization_id, worker_spec_snapshot_id, orchestration_resource_id, orchestration_resource_revision)
VALUES (2000, 1, 101, 7, 1)`,
		},
		{
			name: "workflows",
			query: `
INSERT INTO workflows
	(id, organization_id, worker_spec_snapshot_id, orchestration_resource_id, orchestration_resource_revision)
VALUES (2001, 1, 101, 7, 1)`,
		},
		{
			name: "workflow_runs",
			query: `
INSERT INTO workflow_runs
	(id, organization_id, worker_spec_snapshot_id, orchestration_resource_id, orchestration_resource_revision)
VALUES (2002, 1, 101, 7, 1)`,
		},
		{
			name: "goal_loops",
			query: `
INSERT INTO goal_loops
	(id, organization_id, worker_spec_snapshot_id, orchestration_resource_id, orchestration_resource_revision)
VALUES (2003, 1, 101, 7, 1)`,
		},
		{
			name: "orchestration_worker_launches",
			query: `
INSERT INTO orchestration_worker_launches
	(id, organization_id, worker_spec_snapshot_id, resource_id, resource_revision)
VALUES (2004, 1, 101, 7, 1)`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			requireSnapshotCommitError(t, ctx, conn, tc.query)
		})
	}

	requireSnapshotMigration(t, ctx, conn, "000219_enforce_orchestration_domain_snapshot_consistency.down.sql")
	require.NoError(t, insertMismatchedExpert(ctx, conn, 3000))
}

func TestMigration000219ExistingMismatchFailsClosedPostgres(t *testing.T) {
	ctx, conn := openSnapshotConsistencyPostgres(t)
	requireSnapshotConsistencySchema(t, ctx, conn, true)

	up, err := FS.ReadFile("000219_enforce_orchestration_domain_snapshot_consistency.up.sql")
	require.NoError(t, err)
	require.Error(t, execMigrationSQL(ctx, conn, string(up)))
	require.NoError(t, execSQL(ctx, conn, `ROLLBACK`))
	require.False(t, snapshotConstraintExists(
		t,
		ctx,
		conn,
		"orchestration_resource_revisions_org_revision_snapshot_unique",
	))
	require.NoError(t, insertMismatchedExpert(ctx, conn, 4000))
	requireSnapshotCommitError(t, ctx, conn, `
INSERT INTO experts
	(id, organization_id, worker_spec_snapshot_id, orchestration_resource_id, orchestration_resource_revision)
VALUES (4001, 1, 100, 999, 1)`)
}

func openSnapshotConsistencyPostgres(t *testing.T) (context.Context, *sql.Conn) {
	t.Helper()
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
	t.Cleanup(cancel)
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	conn, err := db.Conn(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, conn.Close()) })

	schema := fmt.Sprintf("snapshot_consistency_%d", time.Now().UnixNano())
	require.NoError(t, execSQL(ctx, conn, `CREATE SCHEMA `+schema))
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `DROP SCHEMA IF EXISTS `+schema+` CASCADE`)
	})
	require.NoError(t, execSQL(ctx, conn, `SET search_path TO `+schema))
	return ctx, conn
}

func requireSnapshotConsistencySchema(
	t *testing.T,
	ctx context.Context,
	conn *sql.Conn,
	withMismatch bool,
) {
	t.Helper()
	require.NoError(t, execSQL(ctx, conn, snapshotConsistencyBaseDDL))
	require.NoError(t, insertMatchedSnapshotRows(ctx, conn))
	require.NoError(t, insertSnapshotOnlyExpert(ctx, conn, 20, 100))
	require.NoError(t, insertSnapshotOnlyGoalLoop(ctx, conn, 21, 100))
	if withMismatch {
		require.NoError(t, insertMismatchedExpert(ctx, conn, 22))
	}
}

func requireSnapshotMigration(
	t *testing.T,
	ctx context.Context,
	conn *sql.Conn,
	name string,
) {
	t.Helper()
	content, err := FS.ReadFile(name)
	require.NoError(t, err)
	require.NoError(t, execMigrationSQL(ctx, conn, string(content)))
}

func requireSnapshotCommitError(
	t *testing.T,
	ctx context.Context,
	conn *sql.Conn,
	query string,
) {
	t.Helper()
	tx, err := conn.BeginTx(ctx, nil)
	require.NoError(t, err)
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, query)
	require.NoError(t, err)
	require.Error(t, tx.Commit())
}

func insertMatchedSnapshotRows(ctx context.Context, conn *sql.Conn) error {
	for _, query := range []string{
		`INSERT INTO experts
			(id, organization_id, worker_spec_snapshot_id, orchestration_resource_id, orchestration_resource_revision)
		 VALUES (10, 1, 100, 7, 1)`,
		`INSERT INTO workflows
			(id, organization_id, worker_spec_snapshot_id, orchestration_resource_id, orchestration_resource_revision)
		 VALUES (11, 1, 100, 7, 1)`,
		`INSERT INTO workflow_runs
			(id, organization_id, worker_spec_snapshot_id, orchestration_resource_id, orchestration_resource_revision)
		 VALUES (12, 1, 100, 7, 1)`,
		`INSERT INTO goal_loops
			(id, organization_id, worker_spec_snapshot_id, orchestration_resource_id, orchestration_resource_revision)
		 VALUES (13, 1, 100, 7, 1)`,
		`INSERT INTO orchestration_worker_launches
			(id, organization_id, worker_spec_snapshot_id, resource_id, resource_revision)
		 VALUES (14, 1, 100, 7, 1)`,
	} {
		if _, err := conn.ExecContext(ctx, query); err != nil {
			return err
		}
	}
	return nil
}

func insertSnapshotOnlyExpert(ctx context.Context, conn *sql.Conn, id, snapshotID int64) error {
	_, err := conn.ExecContext(ctx, `
INSERT INTO experts
	(id, organization_id, worker_spec_snapshot_id, orchestration_resource_id, orchestration_resource_revision)
VALUES ($1, 1, $2, NULL, NULL)`, id, snapshotID)
	return err
}

func insertSnapshotOnlyGoalLoop(ctx context.Context, conn *sql.Conn, id, snapshotID int64) error {
	_, err := conn.ExecContext(ctx, `
INSERT INTO goal_loops
	(id, organization_id, worker_spec_snapshot_id, orchestration_resource_id, orchestration_resource_revision)
VALUES ($1, 1, $2, NULL, NULL)`, id, snapshotID)
	return err
}

func insertMismatchedExpert(ctx context.Context, conn *sql.Conn, id int64) error {
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, `
INSERT INTO experts
	(id, organization_id, worker_spec_snapshot_id, orchestration_resource_id, orchestration_resource_revision)
VALUES ($1, 1, 101, 7, 1)`, id)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func snapshotConstraintExists(
	t *testing.T,
	ctx context.Context,
	conn *sql.Conn,
	name string,
) bool {
	t.Helper()
	var exists bool
	err := conn.QueryRowContext(ctx, `
SELECT EXISTS (
	SELECT 1
	FROM pg_constraint
	WHERE conname = $1
	  AND connamespace = current_schema()::regnamespace
)`, name).Scan(&exists)
	require.NoError(t, err)
	return exists
}

const snapshotConsistencyBaseDDL = `
CREATE TABLE worker_spec_snapshots (
	id BIGINT PRIMARY KEY,
	organization_id BIGINT NOT NULL,
	UNIQUE (organization_id, id)
);

CREATE TABLE orchestration_resource_revisions (
	id BIGINT PRIMARY KEY,
	organization_id BIGINT NOT NULL,
	resource_id BIGINT NOT NULL,
	revision BIGINT NOT NULL,
	worker_spec_snapshot_id BIGINT,
	CONSTRAINT orchestration_resource_revisions_org_revision_unique
		UNIQUE (organization_id, resource_id, revision),
	CONSTRAINT orchestration_resource_revisions_snapshot_fkey
		FOREIGN KEY (organization_id, worker_spec_snapshot_id)
		REFERENCES worker_spec_snapshots (organization_id, id)
);

CREATE TABLE experts (
	id BIGINT PRIMARY KEY,
	organization_id BIGINT NOT NULL,
	worker_spec_snapshot_id BIGINT,
	orchestration_resource_id BIGINT,
	orchestration_resource_revision BIGINT,
	CONSTRAINT experts_worker_spec_snapshot_org_fkey
		FOREIGN KEY (organization_id, worker_spec_snapshot_id)
		REFERENCES worker_spec_snapshots (organization_id, id),
	CONSTRAINT experts_orchestration_revision_fkey
		FOREIGN KEY (organization_id, orchestration_resource_id, orchestration_resource_revision)
		REFERENCES orchestration_resource_revisions (organization_id, resource_id, revision)
		DEFERRABLE INITIALLY DEFERRED
);

CREATE TABLE workflows (
	id BIGINT PRIMARY KEY,
	organization_id BIGINT NOT NULL,
	worker_spec_snapshot_id BIGINT,
	orchestration_resource_id BIGINT,
	orchestration_resource_revision BIGINT,
	CONSTRAINT workflows_worker_spec_snapshot_org_fkey
		FOREIGN KEY (organization_id, worker_spec_snapshot_id)
		REFERENCES worker_spec_snapshots (organization_id, id),
	CONSTRAINT workflows_orchestration_revision_fkey
		FOREIGN KEY (organization_id, orchestration_resource_id, orchestration_resource_revision)
		REFERENCES orchestration_resource_revisions (organization_id, resource_id, revision)
		DEFERRABLE INITIALLY DEFERRED
);

CREATE TABLE workflow_runs (
	id BIGINT PRIMARY KEY,
	organization_id BIGINT NOT NULL,
	worker_spec_snapshot_id BIGINT,
	orchestration_resource_id BIGINT,
	orchestration_resource_revision BIGINT,
	CONSTRAINT workflow_runs_worker_spec_snapshot_org_fkey
		FOREIGN KEY (organization_id, worker_spec_snapshot_id)
		REFERENCES worker_spec_snapshots (organization_id, id),
	CONSTRAINT workflow_runs_orchestration_revision_fkey
		FOREIGN KEY (organization_id, orchestration_resource_id, orchestration_resource_revision)
		REFERENCES orchestration_resource_revisions (organization_id, resource_id, revision)
		DEFERRABLE INITIALLY DEFERRED
);

CREATE TABLE goal_loops (
	id BIGINT PRIMARY KEY,
	organization_id BIGINT NOT NULL,
	worker_spec_snapshot_id BIGINT NOT NULL,
	orchestration_resource_id BIGINT,
	orchestration_resource_revision BIGINT,
	CONSTRAINT goal_loops_worker_spec_snapshot_org_fkey
		FOREIGN KEY (organization_id, worker_spec_snapshot_id)
		REFERENCES worker_spec_snapshots (organization_id, id),
	CONSTRAINT goal_loops_orchestration_revision_fkey
		FOREIGN KEY (organization_id, orchestration_resource_id, orchestration_resource_revision)
		REFERENCES orchestration_resource_revisions (organization_id, resource_id, revision)
		DEFERRABLE INITIALLY DEFERRED
);

CREATE TABLE orchestration_worker_launches (
	id BIGINT PRIMARY KEY,
	organization_id BIGINT NOT NULL,
	worker_spec_snapshot_id BIGINT NOT NULL,
	resource_id BIGINT NOT NULL,
	resource_revision BIGINT NOT NULL,
	CONSTRAINT orchestration_worker_launches_snapshot_fkey
		FOREIGN KEY (organization_id, worker_spec_snapshot_id)
		REFERENCES worker_spec_snapshots (organization_id, id),
	CONSTRAINT orchestration_worker_launches_revision_fkey
		FOREIGN KEY (organization_id, resource_id, resource_revision)
		REFERENCES orchestration_resource_revisions (organization_id, resource_id, revision)
		DEFERRABLE INITIALLY DEFERRED
);

INSERT INTO worker_spec_snapshots (id, organization_id)
VALUES (100, 1), (101, 1), (200, 2);

INSERT INTO orchestration_resource_revisions
	(id, organization_id, resource_id, revision, worker_spec_snapshot_id)
VALUES
	(1, 1, 7, 1, 100),
	(2, 1, 8, 1, 101),
	(3, 2, 9, 1, 200);
`
