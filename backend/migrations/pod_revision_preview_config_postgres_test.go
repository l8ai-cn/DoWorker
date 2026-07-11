package migrations

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestMigration000204PodRevisionPreviewConfigUpDownPostgres(t *testing.T) {
	dsn, err := migrationPostgresDSN()
	require.NoError(t, err)
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
	preview_path VARCHAR(255) NOT NULL DEFAULT '',
	active_config_revision_id BIGINT
);
CREATE TABLE pod_config_revisions (
	id BIGINT PRIMARY KEY,
	pod_id BIGINT NOT NULL
);
INSERT INTO pods(id, preview_port, preview_path, active_config_revision_id)
VALUES
	(1, 3000, '/app//api/', 11),
	(2, 0, '', 12),
	(3, 80, '/legacy', 13),
	(4, 70000, 'relative', 14),
	(5, 4000, '/app/%2e%2e/admin', 15),
	(6, 4000, '/files/report%23draft.pdf', 16),
	(7, 4000, '/route/%3F', 17),
	(8, 4000, '/bad%2', 18),
	(9, 4000, '/raw?query=1', 19),
	(10, 4000, '/raw#fragment', 20),
	(11, 4000, '/raw/../traversal', 21),
	(12, 4000, '/app/./api/', 22),
	(13, 4000, '/.', 23);
INSERT INTO pod_config_revisions(id, pod_id)
VALUES
	(11, 1), (12, 2), (13, 3), (14, 4), (15, 5), (16, 6),
	(17, 7), (18, 8), (19, 9), (20, 10), (21, 11), (22, 12),
	(23, 13);
`))

	up, err := FS.ReadFile("000204_add_preview_config_to_pod_revisions.up.sql")
	require.NoError(t, err)
	require.NoError(t, execSQL(ctx, conn, string(up)))
	requirePreviewConfig(t, ctx, conn, 1, 3000, "/app/api")
	requirePreviewConfig(t, ctx, conn, 2, 0, "/")
	requirePreviewConfig(t, ctx, conn, 3, 0, "/legacy")
	requirePreviewConfig(t, ctx, conn, 4, 0, "/")
	requirePreviewConfig(t, ctx, conn, 5, 4000, "/")
	requirePreviewConfig(t, ctx, conn, 6, 4000, "/files/report%23draft.pdf")
	requirePreviewConfig(t, ctx, conn, 7, 4000, "/route/%3F")
	requirePreviewConfig(t, ctx, conn, 8, 4000, "/")
	requirePreviewConfig(t, ctx, conn, 9, 4000, "/")
	requirePreviewConfig(t, ctx, conn, 10, 4000, "/")
	requirePreviewConfig(t, ctx, conn, 11, 4000, "/")
	requirePreviewConfig(t, ctx, conn, 12, 4000, "/app/api")
	requirePreviewConfig(t, ctx, conn, 13, 4000, "/")

	// Database checks cover canonical storage form. Decoded URL behavior is
	// exercised separately by the relay service contract tests.
	for _, path := range []string{
		"relative",
		"/app/../admin",
		"/app/%2e%2e/admin",
		"/raw?query=1",
		"/raw#fragment",
		"/bad%2",
	} {
		_, err := conn.ExecContext(ctx, `
INSERT INTO pod_config_revisions(id, pod_id, preview_port, preview_path)
VALUES (100, 1, 3000, $1)
`, path)
		require.Error(t, err, path)
	}
	require.Error(t, execSQL(ctx, conn, `
INSERT INTO pod_config_revisions(id, pod_id, preview_port, preview_path)
VALUES (100, 1, 80, '/')
`))
	require.NoError(t, execSQL(ctx, conn, `
INSERT INTO pod_config_revisions(id, pod_id, preview_port, preview_path)
VALUES
	(101, 1, 3000, '/files/report%23draft.pdf'),
	(102, 1, 3000, '/route/%3F')
`))
	require.Error(t, execSQL(ctx, conn, `UPDATE pods SET preview_port = 80 WHERE id = 1`))
	require.Error(t, execSQL(ctx, conn, `UPDATE pods SET preview_path = '/bad%2' WHERE id = 1`))

	down, err := FS.ReadFile("000204_add_preview_config_to_pod_revisions.down.sql")
	require.NoError(t, err)
	require.NoError(t, execSQL(ctx, conn, string(down)))
	require.False(t, postgresColumnExists(ctx, t, conn, "pod_config_revisions", "preview_port"))
	require.False(t, postgresColumnExists(ctx, t, conn, "pod_config_revisions", "preview_path"))
	require.True(t, postgresColumnExists(ctx, t, conn, "pod_config_revisions", "pod_id"))
	require.NoError(t, execSQL(ctx, conn, `UPDATE pods SET preview_port = 80, preview_path = 'relative' WHERE id = 1`))
}

func TestMigrationPostgresDSNRequiredInCI(t *testing.T) {
	t.Setenv("CI", "true")
	t.Setenv("MIGRATIONS_POSTGRES_TEST_DSN", "")

	_, err := migrationPostgresDSN()
	require.Error(t, err)
}

func migrationPostgresDSN() (string, error) {
	dsn := os.Getenv("MIGRATIONS_POSTGRES_TEST_DSN")
	if dsn == "" && os.Getenv("CI") == "true" {
		return "", errors.New("MIGRATIONS_POSTGRES_TEST_DSN is required in CI")
	}
	return dsn, nil
}

func requirePreviewConfig(
	t *testing.T,
	ctx context.Context,
	conn *sql.Conn,
	podID int64,
	wantPort int,
	wantPath string,
) {
	t.Helper()
	var podPort, revisionPort int
	var podPath, revisionPath string
	err := conn.QueryRowContext(
		ctx,
		`SELECT p.preview_port, p.preview_path, r.preview_port, r.preview_path
		 FROM pods p
		 JOIN pod_config_revisions r ON r.id = p.active_config_revision_id
		 WHERE p.id = $1`,
		podID,
	).Scan(&podPort, &podPath, &revisionPort, &revisionPath)
	require.NoError(t, err)
	require.Equal(t, wantPort, podPort)
	require.Equal(t, wantPort, revisionPort)
	require.Equal(t, wantPath, podPath)
	require.Equal(t, wantPath, revisionPath)
}
