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

func TestMigration000228AllowsEmptyOptionalWorkerModelBinding(
	t *testing.T,
) {
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

	schema := fmt.Sprintf("worker_optional_binding_%d", time.Now().UnixNano())
	require.NoError(t, execSQL(ctx, conn, `CREATE SCHEMA `+schema))
	t.Cleanup(func() {
		_, _ = db.ExecContext(
			context.Background(),
			`DROP SCHEMA IF EXISTS `+schema+` CASCADE`,
		)
	})
	require.NoError(t, execSQL(ctx, conn, `SET search_path TO `+schema))
	require.NoError(t, execSQL(ctx, conn, `
CREATE TABLE organizations (id BIGINT PRIMARY KEY);
INSERT INTO organizations (id) VALUES (1);
`))

	for _, name := range []string{
		"000194_worker_spec_snapshots.up.sql",
		"000197_worker_spec_model_binding.up.sql",
		"000209_worker_spec_protocol_adapter.up.sql",
		"000228_worker_spec_optional_model_binding.up.sql",
	} {
		migration, err := FS.ReadFile(name)
		require.NoError(t, err)
		require.NoError(t, execMigrationSQL(ctx, conn, string(migration)))
	}

	requireWorkerSpecSnapshotAccepted(t, ctx, conn,
		`{"runtime":{"model_binding":{}},"version":1}`,
		`{"model_binding":{},"version":1}`,
	)
	requireWorkerSpecSnapshotRejectedWithSummary(t, ctx, conn,
		`{"runtime":{"model_binding":{}},"version":1}`,
		`{"model_binding":{"resource_id":1001,"resource_revision":7,"connection_id":2001,"connection_revision":9,"provider_key":"openai","protocol_adapter":"openai-compatible","model_id":"gpt-5"}},"version":1}`,
	)
}

func TestMigration000228ProtectsOptionalModelBindingRollback(t *testing.T) {
	up := readMigrationForTest(
		t,
		"000228_worker_spec_optional_model_binding.up.sql",
	)
	require.Contains(t, up, "binding = '{}'::JSONB THEN TRUE")

	down := readMigrationForTest(
		t,
		"000228_worker_spec_optional_model_binding.down.sql",
	)
	require.Contains(
		t,
		down,
		"worker_spec_snapshots contain optional model bindings; remove them before rollback",
	)
}
