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

func TestMigration000195PodConfigRevisionsUpDownPostgres(t *testing.T) {
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

	schema := fmt.Sprintf("pod_cfg_rev_%d", time.Now().UnixNano())
	require.NoError(t, execSQL(ctx, conn, `CREATE SCHEMA `+schema))
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `DROP SCHEMA IF EXISTS `+schema+` CASCADE`)
	})
	require.NoError(t, execSQL(ctx, conn, `SET search_path TO `+schema))
	require.NoError(t, execSQL(ctx, conn, `
CREATE TABLE users (id BIGINT PRIMARY KEY);
CREATE TABLE model_resources (id BIGINT PRIMARY KEY);
CREATE TABLE pods (id BIGSERIAL PRIMARY KEY);
`))

	up, err := FS.ReadFile("000195_pod_config_revisions.up.sql")
	require.NoError(t, err)
	require.NoError(t, execSQL(ctx, conn, string(up)))
	require.True(t, postgresTableExists(ctx, t, conn, "pod_config_revisions"))
	require.True(t, postgresColumnExists(ctx, t, conn, "pods", "generation"))
	require.True(t, postgresColumnExists(ctx, t, conn, "pods", "active_config_revision_id"))
	require.True(t, postgresColumnExists(ctx, t, conn, "pods", "model_resource_id"))

	down, err := FS.ReadFile("000195_pod_config_revisions.down.sql")
	require.NoError(t, err)
	require.NoError(t, execSQL(ctx, conn, string(down)))
	require.False(t, postgresTableExists(ctx, t, conn, "pod_config_revisions"))
	require.False(t, postgresColumnExists(ctx, t, conn, "pods", "generation"))
	require.False(t, postgresColumnExists(ctx, t, conn, "pods", "model_resource_id"))
}

func execSQL(ctx context.Context, conn *sql.Conn, query string) error {
	for _, stmt := range strings.Split(query, ";") {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := conn.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}

func postgresTableExists(ctx context.Context, t *testing.T, conn *sql.Conn, table string) bool {
	t.Helper()
	var exists bool
	err := conn.QueryRowContext(ctx, `
SELECT EXISTS (
	SELECT 1 FROM information_schema.tables
	WHERE table_schema = current_schema()
	AND table_name = $1
)`, table).Scan(&exists)
	require.NoError(t, err)
	return exists
}

func postgresColumnExists(ctx context.Context, t *testing.T, conn *sql.Conn, table, column string) bool {
	t.Helper()
	var exists bool
	err := conn.QueryRowContext(ctx, `
SELECT EXISTS (
	SELECT 1 FROM information_schema.columns
	WHERE table_schema = current_schema()
	AND table_name = $1
	AND column_name = $2
)`, table, column).Scan(&exists)
	require.NoError(t, err)
	return exists
}
