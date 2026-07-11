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

func TestMigration000203ExpertWorkerSpecLinkPostgres(t *testing.T) {
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

	schema := fmt.Sprintf("expert_worker_spec_link_%d", time.Now().UnixNano())
	require.NoError(t, execSQL(ctx, conn, `CREATE SCHEMA `+schema))
	t.Cleanup(func() {
		_, _ = db.ExecContext(
			context.Background(),
			`DROP SCHEMA IF EXISTS `+schema+` CASCADE`,
		)
	})
	require.NoError(t, execSQL(ctx, conn, `SET search_path TO `+schema))
	require.NoError(t, execSQL(ctx, conn, `
CREATE TABLE worker_spec_snapshots (
	id BIGSERIAL PRIMARY KEY,
	organization_id BIGINT NOT NULL,
	UNIQUE (organization_id, id)
);
CREATE TABLE experts (
	id BIGSERIAL PRIMARY KEY,
	organization_id BIGINT NOT NULL
);
INSERT INTO worker_spec_snapshots(id, organization_id) VALUES (10, 77);
`))

	up, err := FS.ReadFile("000203_expert_worker_spec_link.up.sql")
	require.NoError(t, err)
	require.NoError(t, execSQL(ctx, conn, string(up)))
	require.True(
		t,
		postgresColumnExists(
			ctx,
			t,
			conn,
			"experts",
			"worker_spec_snapshot_id",
		),
	)
	require.NoError(t, execSQL(ctx, conn, `
INSERT INTO experts(id, organization_id, worker_spec_snapshot_id)
VALUES (1, 77, 10)
`))
	err = execSQL(ctx, conn, `
INSERT INTO experts(id, organization_id, worker_spec_snapshot_id)
VALUES (2, 78, 10)
`)
	require.Error(t, err)

	down, err := FS.ReadFile("000203_expert_worker_spec_link.down.sql")
	require.NoError(t, err)
	require.NoError(t, execSQL(ctx, conn, string(down)))
	require.False(
		t,
		postgresColumnExists(
			ctx,
			t,
			conn,
			"experts",
			"worker_spec_snapshot_id",
		),
	)
}
