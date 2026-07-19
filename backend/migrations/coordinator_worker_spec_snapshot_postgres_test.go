package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestMigration000230CoordinatorWorkerSpecSnapshotPostgres(t *testing.T) {
	dsn := os.Getenv("MIGRATIONS_POSTGRES_TEST_DSN")
	if dsn == "" {
		t.Skip("MIGRATIONS_POSTGRES_TEST_DSN is not configured")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer db.Close()

	t.Run("same org fk and cross org rejection", func(t *testing.T) {
		conn := coordinatorMigrationConn(t, ctx, db, "coord_snapshot_fk")
		defer conn.Close()
		createCoordinatorSnapshotBase(t, ctx, conn)
		applyCoordinatorSnapshotMigration(t, ctx, conn, "000230_coordinator_worker_spec_snapshot.up.sql")

		require.NoError(t, execSQL(ctx, conn, `
INSERT INTO coordinator_projects(id, organization_id, worker_spec_snapshot_id) VALUES (1, 10, 100);
`))
		err := execSQL(ctx, conn, `
INSERT INTO coordinator_projects(id, organization_id, worker_spec_snapshot_id) VALUES (2, 20, 100);
`)
		require.Error(t, err)
	})

	t.Run("existing projects remain null until audited binding", func(t *testing.T) {
		conn := coordinatorMigrationConn(t, ctx, db, "coord_snapshot_existing")
		defer conn.Close()
		createCoordinatorSnapshotBase(t, ctx, conn)
		require.NoError(t, execSQL(ctx, conn, `
INSERT INTO coordinator_projects(id, organization_id) VALUES (1, 10);
`))

		applyCoordinatorSnapshotMigration(t, ctx, conn, "000230_coordinator_worker_spec_snapshot.up.sql")

		require.True(t, postgresColumnExists(ctx, t, conn, "coordinator_projects", "worker_spec_snapshot_id"))
		var snapshotID sql.NullInt64
		require.NoError(t, conn.QueryRowContext(ctx, `
SELECT worker_spec_snapshot_id FROM coordinator_projects WHERE id = 1
`).Scan(&snapshotID))
		require.False(t, snapshotID.Valid)
		require.NoError(t, execSQL(ctx, conn, `
UPDATE coordinator_projects SET worker_spec_snapshot_id = 100 WHERE id = 1;
`))
	})

	t.Run("non empty down is blocked", func(t *testing.T) {
		conn := coordinatorMigrationConn(t, ctx, db, "coord_snapshot_down")
		defer conn.Close()
		createCoordinatorSnapshotBase(t, ctx, conn)
		applyCoordinatorSnapshotMigration(t, ctx, conn, "000230_coordinator_worker_spec_snapshot.up.sql")
		require.NoError(t, execSQL(ctx, conn, `
INSERT INTO coordinator_projects(id, organization_id, worker_spec_snapshot_id) VALUES (1, 10, 100);
`))

		err := applyCoordinatorSnapshotMigrationErr(ctx, conn, "000230_coordinator_worker_spec_snapshot.down.sql")

		require.Error(t, err)
		require.Contains(t, err.Error(), "worker spec snapshot bindings")
		require.NoError(t, execSQL(ctx, conn, `ROLLBACK`))
		require.True(t, postgresColumnExists(ctx, t, conn, "coordinator_projects", "worker_spec_snapshot_id"))
	})

	t.Run("empty table down up is reversible", func(t *testing.T) {
		conn := coordinatorMigrationConn(t, ctx, db, "coord_snapshot_empty")
		defer conn.Close()
		createCoordinatorSnapshotBase(t, ctx, conn)
		applyCoordinatorSnapshotMigration(t, ctx, conn, "000230_coordinator_worker_spec_snapshot.up.sql")
		require.True(t, postgresColumnExists(ctx, t, conn, "coordinator_projects", "worker_spec_snapshot_id"))

		applyCoordinatorSnapshotMigration(t, ctx, conn, "000230_coordinator_worker_spec_snapshot.down.sql")
		require.False(t, postgresColumnExists(ctx, t, conn, "coordinator_projects", "worker_spec_snapshot_id"))

		applyCoordinatorSnapshotMigration(t, ctx, conn, "000230_coordinator_worker_spec_snapshot.up.sql")
		require.True(t, postgresColumnExists(ctx, t, conn, "coordinator_projects", "worker_spec_snapshot_id"))
	})

	t.Run("down lock observes concurrent binding", func(t *testing.T) {
		conn := coordinatorMigrationConn(t, ctx, db, "coord_snapshot_lock")
		defer conn.Close()
		createCoordinatorSnapshotBase(t, ctx, conn)
		applyCoordinatorSnapshotMigration(t, ctx, conn, "000230_coordinator_worker_spec_snapshot.up.sql")

		var schema string
		require.NoError(t, conn.QueryRowContext(ctx, `SELECT current_schema()`).Scan(&schema))
		writer := coordinatorMigrationSchemaConn(t, ctx, db, schema)
		defer writer.Close()
		down := coordinatorMigrationSchemaConn(t, ctx, db, schema)
		defer down.Close()

		require.NoError(t, execSQL(ctx, writer, `
BEGIN;
INSERT INTO coordinator_projects(id, organization_id, worker_spec_snapshot_id) VALUES (1, 10, 100);
`))
		done := make(chan error, 1)
		go func() {
			done <- applyCoordinatorSnapshotMigrationErr(
				context.Background(),
				down,
				"000230_coordinator_worker_spec_snapshot.down.sql",
			)
		}()
		select {
		case err := <-done:
			require.Failf(t, "down completed before writer", "error: %v", err)
		case <-time.After(100 * time.Millisecond):
		}
		require.NoError(t, execSQL(ctx, writer, `COMMIT`))
		err := <-done
		require.ErrorContains(t, err, "worker spec snapshot bindings")
		require.NoError(t, execSQL(ctx, down, `ROLLBACK`))
		require.True(t, postgresColumnExists(ctx, t, conn, "coordinator_projects", "worker_spec_snapshot_id"))
	})
}

func coordinatorMigrationConn(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	prefix string,
) *sql.Conn {
	t.Helper()
	conn, err := db.Conn(ctx)
	require.NoError(t, err)
	schema := fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
	require.NoError(t, execSQL(ctx, conn, `CREATE SCHEMA `+schema))
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `DROP SCHEMA IF EXISTS `+schema+` CASCADE`)
	})
	require.NoError(t, execSQL(ctx, conn, `SET search_path TO `+schema+`, public`))
	return conn
}

func coordinatorMigrationSchemaConn(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	schema string,
) *sql.Conn {
	t.Helper()
	conn, err := db.Conn(ctx)
	require.NoError(t, err)
	require.NoError(t, execSQL(ctx, conn, `SET search_path TO `+schema+`, public`))
	return conn
}

func createCoordinatorSnapshotBase(t *testing.T, ctx context.Context, conn *sql.Conn) {
	t.Helper()
	require.NoError(t, execSQL(ctx, conn, `
CREATE TABLE worker_spec_snapshots (
	id BIGINT NOT NULL,
	organization_id BIGINT NOT NULL,
	PRIMARY KEY (id),
	UNIQUE (organization_id, id)
);
CREATE TABLE coordinator_projects (
	id BIGINT PRIMARY KEY,
	organization_id BIGINT NOT NULL
);
INSERT INTO worker_spec_snapshots(id, organization_id) VALUES (100, 10);
`))
}

func applyCoordinatorSnapshotMigration(
	t *testing.T,
	ctx context.Context,
	conn *sql.Conn,
	name string,
) {
	t.Helper()
	require.NoError(t, applyCoordinatorSnapshotMigrationErr(ctx, conn, name))
}

func applyCoordinatorSnapshotMigrationErr(ctx context.Context, conn *sql.Conn, name string) error {
	raw, err := FS.ReadFile(name)
	if err != nil {
		return err
	}
	_, err = conn.ExecContext(ctx, strings.TrimSpace(string(raw)))
	return err
}
