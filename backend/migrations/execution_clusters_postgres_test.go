package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestMigration000206ExecutionClustersUpDownPostgres(t *testing.T) {
	dsn, err := migrationPostgresDSN()
	require.NoError(t, err)
	if dsn == "" {
		t.Skip("MIGRATIONS_POSTGRES_TEST_DSN is not configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer db.Close()
	conn, err := db.Conn(ctx)
	require.NoError(t, err)
	defer conn.Close()

	schema := fmt.Sprintf("execution_clusters_%d", time.Now().UnixNano())
	require.NoError(t, execSQL(ctx, conn, `CREATE SCHEMA `+schema))
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `DROP SCHEMA IF EXISTS `+schema+` CASCADE`)
	})
	require.NoError(t, execSQL(ctx, conn, `SET search_path TO `+schema))
	for index, statement := range executionClusterFixtureStatements {
		require.NoErrorf(t, execSQL(ctx, conn, statement), "fixture statement %d", index)
	}

	up, err := FS.ReadFile("000206_execution_clusters.up.sql")
	require.NoError(t, err)
	require.NoError(t, execSQL(ctx, conn, string(up)))

	require.Equal(t, int64(2), executionClusterCount(t, ctx, conn, 1))
	require.Equal(t, int64(2), executionClusterCount(t, ctx, conn, 2))
	require.Equal(t, "online", executionClusterSlug(t, ctx, conn, "runners", 10))
	require.Equal(t, "online", executionClusterSlug(t, ctx, conn, "pods", 100))
	require.Equal(t, "local", executionClusterSlug(t, ctx, conn, "runner_grpc_registration_tokens", 1000))
	require.Equal(t, "online", executionClusterSlug(t, ctx, conn, "runner_pending_auths", 2000))
	require.Equal(t, "local", executionClusterSlug(t, ctx, conn, "runner_pending_auths", 2001))
	var normalizedAuthorized bool
	require.NoError(t, conn.QueryRowContext(ctx, `
SELECT authorized FROM runner_pending_auths WHERE id = 2005
`).Scan(&normalizedAuthorized))
	require.False(t, normalizedAuthorized)

	var pendingCluster sql.NullInt64
	require.NoError(t, conn.QueryRowContext(ctx, `
SELECT cluster_id FROM runner_pending_auths WHERE id = 2002
`).Scan(&pendingCluster))
	require.False(t, pendingCluster.Valid)

	otherOrgClusterID := executionClusterID(t, ctx, conn, 2, "online")
	_, err = conn.ExecContext(ctx, `UPDATE runners SET cluster_id = $1 WHERE id = 10`, otherOrgClusterID)
	require.Error(t, err)

	_, err = conn.ExecContext(ctx, `
INSERT INTO runner_pending_auths (id, organization_id, runner_id, authorized, cluster_id)
VALUES (2003, NULL, NULL, FALSE, NULL)
`)
	require.NoError(t, err)
	_, err = conn.ExecContext(ctx, `
INSERT INTO runner_pending_auths (id, organization_id, runner_id, authorized, cluster_id)
VALUES (2004, NULL, NULL, TRUE, NULL)
`)
	require.Error(t, err)

	down, err := FS.ReadFile("000206_execution_clusters.down.sql")
	require.NoError(t, err)
	require.NoError(t, execSQL(ctx, conn, string(down)))
	require.False(t, postgresColumnExists(ctx, t, conn, "runners", "cluster_id"))
	require.False(t, postgresColumnExists(ctx, t, conn, "pods", "cluster_id"))
	require.False(t, postgresColumnExists(ctx, t, conn, "runner_grpc_registration_tokens", "cluster_id"))
	require.False(t, postgresColumnExists(ctx, t, conn, "runner_pending_auths", "cluster_id"))
}

var executionClusterFixtureStatements = []string{
	`CREATE TABLE organizations (id BIGINT PRIMARY KEY)`,
	`CREATE TABLE runners (
  id BIGINT PRIMARY KEY,
  organization_id BIGINT NOT NULL,
  node_id TEXT NOT NULL
)`,
	`CREATE TABLE pods (
  id BIGINT PRIMARY KEY,
  organization_id BIGINT NOT NULL,
  runner_id BIGINT NOT NULL
)`,
	`CREATE TABLE runner_grpc_registration_tokens (
  id BIGINT PRIMARY KEY,
  organization_id BIGINT NOT NULL
)`,
	`CREATE TABLE runner_pending_auths (
  id BIGINT PRIMARY KEY,
  organization_id BIGINT,
  runner_id BIGINT,
  authorized BOOLEAN DEFAULT FALSE
)`,
	`INSERT INTO organizations (id) VALUES (1), (2)`,
	`INSERT INTO runners (id, organization_id, node_id) VALUES (10, 1, 'online-1'), (20, 2, 'online-2')`,
	`INSERT INTO pods (id, organization_id, runner_id) VALUES (100, 1, 10)`,
	`INSERT INTO runner_grpc_registration_tokens (id, organization_id) VALUES (1000, 1)`,
	`INSERT INTO runner_pending_auths (id, organization_id, runner_id, authorized) VALUES
  (2000, 1, 10, TRUE),
  (2001, 1, NULL, FALSE),
  (2002, NULL, NULL, FALSE),
  (2005, NULL, NULL, NULL)`,
}

func executionClusterCount(t *testing.T, ctx context.Context, conn *sql.Conn, orgID int64) int64 {
	t.Helper()
	var count int64
	require.NoError(t, conn.QueryRowContext(ctx, `
SELECT count(*) FROM execution_clusters WHERE organization_id = $1
`, orgID).Scan(&count))
	return count
}

func executionClusterSlug(
	t *testing.T,
	ctx context.Context,
	conn *sql.Conn,
	table string,
	id int64,
) string {
	t.Helper()
	var slug string
	require.NoError(t, conn.QueryRowContext(ctx, `
SELECT cluster.slug
FROM `+table+` AS record
JOIN execution_clusters AS cluster ON cluster.id = record.cluster_id
WHERE record.id = $1
`, id).Scan(&slug))
	return slug
}

func executionClusterID(
	t *testing.T,
	ctx context.Context,
	conn *sql.Conn,
	orgID int64,
	slug string,
) int64 {
	t.Helper()
	var id int64
	require.NoError(t, conn.QueryRowContext(ctx, `
SELECT id FROM execution_clusters WHERE organization_id = $1 AND slug = $2
`, orgID, slug).Scan(&id))
	return id
}
