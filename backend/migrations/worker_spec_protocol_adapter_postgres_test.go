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

func TestMigration000209WorkerSpecProtocolAdapterPostgres(t *testing.T) {
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

	schema := fmt.Sprintf("worker_protocol_%d", time.Now().UnixNano())
	require.NoError(t, execSQL(ctx, conn, `CREATE SCHEMA `+schema))
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `DROP SCHEMA IF EXISTS `+schema+` CASCADE`)
	})
	require.NoError(t, execSQL(ctx, conn, `SET search_path TO `+schema))
	require.NoError(t, execSQL(ctx, conn, `
CREATE TABLE organizations (id BIGINT PRIMARY KEY);
INSERT INTO organizations (id) VALUES (1);
`))
	for _, name := range []string{
		"000194_worker_spec_snapshots.up.sql",
		"000197_worker_spec_model_binding.up.sql",
	} {
		migration, readErr := FS.ReadFile(name)
		require.NoError(t, readErr)
		require.NoError(t, execMigrationSQL(ctx, conn, string(migration)))
	}
	legacySpec := `{"runtime":{"model_binding":{"resource_id":1,"resource_revision":1,"connection_id":1,"connection_revision":1,"provider_key":"openai","model_id":"legacy"}},"version":1}`
	legacySummary := `{"model_binding":{"resource_id":1,"resource_revision":1,"connection_id":1,"connection_revision":1,"provider_key":"openai","model_id":"legacy"},"version":1}`
	requireWorkerSpecSnapshotAccepted(t, ctx, conn, legacySpec, legacySummary)
	up, err := FS.ReadFile("000209_worker_spec_protocol_adapter.up.sql")
	require.NoError(t, err)
	err = execMigrationSQL(ctx, conn, string(up))
	require.ErrorContains(t, err, "require protocol_adapter backfill")
	require.NoError(t, execSQL(ctx, conn, `DELETE FROM worker_spec_snapshots`))
	require.NoError(t, execMigrationSQL(ctx, conn, string(up)))

	tool := validToolModelBindingJSON()
	requireWorkerSpecSnapshotAccepted(t, ctx, conn,
		`{"runtime":{"model_binding":{},"tool_model_bindings":[`+tool+`]},"version":1}`,
		`{"model_binding":{},"tool_model_bindings":[`+tool+`],"version":1}`,
	)
	requireWorkerSpecSnapshotRejectedWithSummary(t, ctx, conn,
		legacySpec,
		legacySummary,
	)

	rejectToolModelSnapshot(t, ctx, conn,
		strings.Replace(tool, `"openai-compatible"`, `"OpenAI_Compatible"`, 1),
	)
	rejectToolModelSnapshot(t, ctx, conn,
		strings.Replace(tool, `,"protocol_adapter":"openai-compatible"`, "", 1),
	)
	rejectToolModelSnapshot(t, ctx, conn,
		tool+`,`+tool,
	)
	secondRole := strings.Replace(tool, `"video-generator"`, `"image-generator"`, 1)
	rejectToolModelSnapshot(t, ctx, conn, tool+`,`+secondRole)
	rejectToolModelSnapshot(t, ctx, conn,
		strings.Replace(tool, `"video"`, `"unknown"`, 1),
	)
	requireWorkerSpecSnapshotRejectedWithSummary(t, ctx, conn,
		`{"runtime":{"model_binding":{},"tool_model_bindings":[`+tool+`]},"version":1}`,
		`{"model_binding":{},"version":1}`,
	)

	require.NoError(t, execSQL(ctx, conn, `DELETE FROM worker_spec_snapshots`))
	down, err := FS.ReadFile("000209_worker_spec_protocol_adapter.down.sql")
	require.NoError(t, err)
	require.NoError(t, execMigrationSQL(ctx, conn, string(down)))
	requireWorkerSpecSnapshotRejected(t, ctx, conn,
		`{"runtime":{"model_binding":{}},"version":1}`,
	)
}

func validToolModelBindingJSON() string {
	return `{"role":"video-generator","model_binding":{"resource_id":11,"resource_revision":2,"connection_id":21,"connection_revision":3,"provider_key":"volcengine","protocol_adapter":"openai-compatible","model_id":"video-1"},"modality":"video","capability":"video-generation","environment":{"api_key":"VIDEO_API_KEY","base_url":"VIDEO_BASE_URL","model_id":"VIDEO_MODEL_ID"}}`
}

func rejectToolModelSnapshot(
	t *testing.T,
	ctx context.Context,
	conn *sql.Conn,
	tools string,
) {
	t.Helper()
	requireWorkerSpecSnapshotRejectedWithSummary(t, ctx, conn,
		`{"runtime":{"model_binding":{},"tool_model_bindings":[`+tools+`]},"version":1}`,
		`{"model_binding":{},"tool_model_bindings":[`+tools+`],"version":1}`,
	)
}
