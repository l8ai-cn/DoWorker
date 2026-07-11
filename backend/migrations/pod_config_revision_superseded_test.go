package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestMigration000205AddsSupersededRevisionStatus(t *testing.T) {
	up, err := FS.ReadFile("000205_add_superseded_pod_config_revision_status.up.sql")
	require.NoError(t, err)
	upSQL := string(up)
	require.Contains(t, upSQL, "DROP CONSTRAINT pod_config_revisions_status_check")
	require.Contains(t, upSQL, "'superseded'")

	down, err := FS.ReadFile("000205_add_superseded_pod_config_revision_status.down.sql")
	require.NoError(t, err)
	downSQL := string(down)
	require.Contains(t, downSQL, "WHERE status = 'superseded'")
	require.NotContains(t, statusConstraint(downSQL), "'superseded'")
}

func TestMigration000205SupersededStatusUpDownPostgres(t *testing.T) {
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

	schema := fmt.Sprintf("pod_revision_superseded_%d", time.Now().UnixNano())
	require.NoError(t, execSQL(ctx, conn, `CREATE SCHEMA `+schema))
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `DROP SCHEMA IF EXISTS `+schema+` CASCADE`)
	})
	require.NoError(t, execSQL(ctx, conn, `SET search_path TO `+schema))
	require.NoError(t, execSQL(ctx, conn, `
CREATE TABLE pod_config_revisions (
	id BIGSERIAL PRIMARY KEY,
	status VARCHAR(20) NOT NULL,
	CONSTRAINT pod_config_revisions_status_check
	CHECK (status IN ('draft', 'applying', 'active', 'failed'))
);
INSERT INTO pod_config_revisions(status) VALUES ('active');
`))

	up, err := FS.ReadFile("000205_add_superseded_pod_config_revision_status.up.sql")
	require.NoError(t, err)
	require.NoError(t, execSQL(ctx, conn, string(up)))
	require.NoError(t, execSQL(ctx, conn, `
INSERT INTO pod_config_revisions(status) VALUES ('superseded')
`))
	require.Error(t, execSQL(ctx, conn, `
INSERT INTO pod_config_revisions(status) VALUES ('invalid')
`))

	down, err := FS.ReadFile("000205_add_superseded_pod_config_revision_status.down.sql")
	require.NoError(t, err)
	require.NoError(t, execSQL(ctx, conn, string(down)))
	require.Error(t, execSQL(ctx, conn, `
INSERT INTO pod_config_revisions(status) VALUES ('superseded')
`))
	var supersededCount int
	require.NoError(t, conn.QueryRowContext(
		ctx,
		`SELECT count(*) FROM pod_config_revisions WHERE status = 'superseded'`,
	).Scan(&supersededCount))
	require.Zero(t, supersededCount)
}

func statusConstraint(sql string) string {
	index := strings.LastIndex(sql, "CHECK (status IN")
	if index < 0 {
		return ""
	}
	return sql[index:]
}
