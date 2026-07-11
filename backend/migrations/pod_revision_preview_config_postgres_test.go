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

func TestMigration000204PodRevisionPreviewConfigUpDownPostgres(t *testing.T) {
	dsn := os.Getenv("MIGRATIONS_POSTGRES_TEST_DSN")
	if dsn == "" {
		t.Skip("MIGRATIONS_POSTGRES_TEST_DSN is not configured")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer db.Close()
	conn, err := db.Conn(ctx)
	require.NoError(t, err)
	defer conn.Close()

	schema := fmt.Sprintf("pod_revision_preview_%d", time.Now().UnixNano())
	require.NoError(t, execSQL(ctx, conn, `CREATE SCHEMA `+schema))
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `DROP SCHEMA IF EXISTS `+schema+` CASCADE`)
	})
	require.NoError(t, execSQL(ctx, conn, `SET search_path TO `+schema))
	require.NoError(t, execSQL(ctx, conn, `
CREATE TABLE pods (
	id BIGINT PRIMARY KEY,
	preview_port INTEGER NOT NULL DEFAULT 0,
	preview_path VARCHAR(255) NOT NULL DEFAULT ''
);
CREATE TABLE pod_config_revisions (
	id BIGINT PRIMARY KEY,
	pod_id BIGINT NOT NULL
);
INSERT INTO pods(id, preview_port, preview_path)
VALUES (1, 3000, '/app'), (2, 0, '');
INSERT INTO pod_config_revisions(id, pod_id)
VALUES (11, 1), (12, 2);
`))

	up, err := FS.ReadFile("000204_add_preview_config_to_pod_revisions.up.sql")
	require.NoError(t, err)
	require.NoError(t, execSQL(ctx, conn, string(up)))
	requirePreviewRevision(t, ctx, conn, 11, 3000, "/app")
	requirePreviewRevision(t, ctx, conn, 12, 0, "/")
	require.Error(t, execSQL(ctx, conn, `
INSERT INTO pod_config_revisions(id, pod_id, preview_port, preview_path)
VALUES (13, 1, 80, '/')
`))
	require.Error(t, execSQL(ctx, conn, `
INSERT INTO pod_config_revisions(id, pod_id, preview_port, preview_path)
VALUES (14, 1, 3000, 'relative')
`))

	down, err := FS.ReadFile("000204_add_preview_config_to_pod_revisions.down.sql")
	require.NoError(t, err)
	require.NoError(t, execSQL(ctx, conn, string(down)))
	require.False(t, postgresColumnExists(ctx, t, conn, "pod_config_revisions", "preview_port"))
	require.False(t, postgresColumnExists(ctx, t, conn, "pod_config_revisions", "preview_path"))
	require.True(t, postgresColumnExists(ctx, t, conn, "pod_config_revisions", "pod_id"))
}

func requirePreviewRevision(
	t *testing.T,
	ctx context.Context,
	conn *sql.Conn,
	id int64,
	wantPort int,
	wantPath string,
) {
	t.Helper()
	var port int
	var path string
	err := conn.QueryRowContext(
		ctx,
		`SELECT preview_port, preview_path FROM pod_config_revisions WHERE id = $1`,
		id,
	).Scan(&port, &path)
	require.NoError(t, err)
	require.Equal(t, wantPort, port)
	require.Equal(t, wantPath, path)
}
