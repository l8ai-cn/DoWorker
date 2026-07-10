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

func TestMigration000197WorkerSpecModelBindingPostgres(t *testing.T) {
	dsn := os.Getenv("MIGRATIONS_POSTGRES_TEST_DSN")
	if dsn == "" {
		t.Skip("MIGRATIONS_POSTGRES_TEST_DSN is not configured")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer db.Close()

	conn, err := db.Conn(ctx)
	require.NoError(t, err)
	defer conn.Close()

	schema := fmt.Sprintf("worker_binding_%d", time.Now().UnixNano())
	require.NoError(t, execSQL(ctx, conn, `CREATE SCHEMA `+schema))
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `DROP SCHEMA IF EXISTS `+schema+` CASCADE`)
	})
	require.NoError(t, execSQL(ctx, conn, `SET search_path TO `+schema))
	require.NoError(t, execSQL(ctx, conn, `
CREATE TABLE organizations (id BIGINT PRIMARY KEY);
INSERT INTO organizations (id) VALUES (1);
`))

	base, err := FS.ReadFile("000194_worker_spec_snapshots.up.sql")
	require.NoError(t, err)
	require.NoError(t, execMigrationSQL(ctx, conn, string(base)))

	up, err := FS.ReadFile("000197_worker_spec_model_binding.up.sql")
	require.NoError(t, err)

	insertLegacyWorkerSpecSnapshot(t, ctx, conn)
	_, err = conn.ExecContext(ctx, string(up))
	require.ErrorContains(t, err, "worker_spec_snapshots must be empty")
	require.NoError(t, execSQL(ctx, conn, `DELETE FROM worker_spec_snapshots`))
	require.NoError(t, execMigrationSQL(ctx, conn, string(up)))

	insertValidWorkerSpecSnapshot(t, ctx, conn)
	requireWorkerSpecSnapshotRejected(t, ctx, conn, `{"runtime":{"model_binding":{"resource_id":"1001","resource_revision":7,"connection_id":2001,"connection_revision":9,"provider_key":"openai","model_id":"gpt-5"}},"version":1}`)
	requireWorkerSpecSnapshotRejected(t, ctx, conn, `{"runtime":{"model_binding":{"resource_id":1001,"resource_revision":7,"connection_id":2001,"provider_key":"openai","model_id":"gpt-5"}},"version":1}`)
	requireWorkerSpecSnapshotRejected(t, ctx, conn, `{"runtime":{"model_binding":{"resource_id":1001,"resource_revision":0,"connection_id":2001,"connection_revision":9,"provider_key":"openai","model_id":"gpt-5"}},"version":1}`)
	requireWorkerSpecSnapshotRejected(t, ctx, conn, `{"runtime":{"model_binding":{"resource_id":1001,"resource_revision":7.5,"connection_id":2001,"connection_revision":9,"provider_key":"openai","model_id":"gpt-5"}},"version":1}`)
	requireWorkerSpecSnapshotRejected(t, ctx, conn, `{"runtime":{"model_binding":{"resource_id":9223372036854775808,"resource_revision":7,"connection_id":2001,"connection_revision":9,"provider_key":"openai","model_id":"gpt-5"}},"version":1}`)
	requireWorkerSpecSnapshotRejected(t, ctx, conn, `{"runtime":{"model_binding":{"resource_id":1001,"resource_revision":7,"connection_id":2001,"connection_revision":9,"provider_key":"OpenAI","model_id":"gpt-5"}},"version":1}`)
	requireWorkerSpecSnapshotRejected(t, ctx, conn, `{"runtime":{"model_binding":{"resource_id":1001,"resource_revision":7,"connection_id":2001,"connection_revision":9,"provider_key":"a","model_id":"gpt-5"}},"version":1}`)
	requireWorkerSpecSnapshotRejected(t, ctx, conn, `{"runtime":{"model_binding":{"resource_id":1001,"resource_revision":7,"connection_id":2001,"connection_revision":9,"provider_key":"openai","model_id":" "}},"version":1}`)
	requireWorkerSpecSnapshotRejected(t, ctx, conn, `{"runtime":{"model_binding":{"resource_id":1001,"resource_revision":7,"connection_id":2001,"connection_revision":9,"provider_key":"openai","model_id":"gpt-5","unexpected":true}},"version":1}`)
	requireWorkerSpecSnapshotRejectedWithSummary(t, ctx, conn,
		`{"runtime":{"model_binding":{"resource_id":1001,"resource_revision":7,"connection_id":2001,"connection_revision":9,"provider_key":"openai","model_id":"gpt-5"}},"version":1}`,
		`{"model_binding":{"resource_id":1002,"resource_revision":7,"connection_id":2001,"connection_revision":9,"provider_key":"openai","model_id":"gpt-5"},"version":1}`,
	)

	require.NoError(t, execSQL(ctx, conn, `DELETE FROM worker_spec_snapshots`))
	down, err := FS.ReadFile("000197_worker_spec_model_binding.down.sql")
	require.NoError(t, err)
	require.NoError(t, execMigrationSQL(ctx, conn, string(down)))
	insertLegacyWorkerSpecSnapshot(t, ctx, conn)
}

func insertLegacyWorkerSpecSnapshot(
	t *testing.T,
	ctx context.Context,
	conn *sql.Conn,
) {
	t.Helper()
	_, err := conn.ExecContext(ctx, `
INSERT INTO worker_spec_snapshots (organization_id, version, spec_json, summary_json)
VALUES (
	1,
	1,
	'{"runtime":{"model_resource_id":1001},"version":1}'::jsonb,
	'{"model_resource_id":1001,"version":1}'::jsonb
)`)
	require.NoError(t, err)
}

func insertValidWorkerSpecSnapshot(
	t *testing.T,
	ctx context.Context,
	conn *sql.Conn,
) {
	t.Helper()
	requireWorkerSpecSnapshotAccepted(t, ctx, conn,
		`{"runtime":{"model_binding":{"resource_id":1001,"resource_revision":7,"connection_id":2001,"connection_revision":9,"provider_key":"openai","model_id":"gpt-5"}},"version":1}`,
		`{"model_binding":{"resource_id":1001,"resource_revision":7,"connection_id":2001,"connection_revision":9,"provider_key":"openai","model_id":"gpt-5"},"version":1}`,
	)
}

func requireWorkerSpecSnapshotRejected(
	t *testing.T,
	ctx context.Context,
	conn *sql.Conn,
	specJSON string,
) {
	t.Helper()
	requireWorkerSpecSnapshotRejectedWithSummary(t, ctx, conn, specJSON,
		`{"model_binding":{"resource_id":1001,"resource_revision":7,"connection_id":2001,"connection_revision":9,"provider_key":"openai","model_id":"gpt-5"},"version":1}`,
	)
}

func requireWorkerSpecSnapshotRejectedWithSummary(
	t *testing.T,
	ctx context.Context,
	conn *sql.Conn,
	specJSON, summaryJSON string,
) {
	t.Helper()
	_, err := conn.ExecContext(ctx, `
INSERT INTO worker_spec_snapshots (organization_id, version, spec_json, summary_json)
VALUES (1, 1, $1::jsonb, $2::jsonb)`, specJSON, summaryJSON)
	require.Error(t, err)
}

func requireWorkerSpecSnapshotAccepted(
	t *testing.T,
	ctx context.Context,
	conn *sql.Conn,
	specJSON, summaryJSON string,
) {
	t.Helper()
	_, err := conn.ExecContext(ctx, `
INSERT INTO worker_spec_snapshots (organization_id, version, spec_json, summary_json)
VALUES (1, 1, $1::jsonb, $2::jsonb)`, specJSON, summaryJSON)
	require.NoError(t, err)
}

func execMigrationSQL(ctx context.Context, conn *sql.Conn, query string) error {
	_, err := conn.ExecContext(ctx, query)
	return err
}
