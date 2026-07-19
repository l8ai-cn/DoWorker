package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMigration000231RepairsPost229ResourceLineagePostgres(t *testing.T) {
	t.Run("recreates wrong adapter constraint", func(t *testing.T) {
		dsn := newMigrationLineageSchema(t)
		migrateTo229(t, dsn)
		db := openMigrationDB(t, dsn)
		defer db.Close()
		_, err := db.Exec(`
ALTER TABLE agents DROP CONSTRAINT IF EXISTS agents_adapter_id_check;
ALTER TABLE agents ADD CONSTRAINT agents_adapter_id_check CHECK (adapter_id IS NOT NULL);
`)
		require.NoError(t, err)

		migratePost229LineageRepair(t, dsn)

		_, err = db.Exec(`
INSERT INTO agents(
  slug, name, launch_command, executable, adapter_id, is_builtin, is_active,
  supported_modes, agentfile_source, created_at, updated_at
) VALUES (
  'invalid-adapter', 'Invalid', 'invalid', 'invalid', 'Bad_Adapter', false, true,
  'pty', 'AGENT invalid', NOW(), NOW()
);
`)
		require.Error(t, err)
	})

	t.Run("does not operate on later search path schemas", func(t *testing.T) {
		dsn := newMigrationLineageSchema(t)
		db := openMigrationDB(t, dsn)
		defer db.Close()
		conn, err := db.Conn(context.Background())
		require.NoError(t, err)
		defer conn.Close()
		emptySchema := fmt.Sprintf("empty_lineage_%d", time.Now().UnixNano())
		shadowSchema := fmt.Sprintf("shadow_lineage_%d", time.Now().UnixNano())
		_, err = conn.ExecContext(context.Background(), fmt.Sprintf(`
CREATE SCHEMA %s;
CREATE SCHEMA %s;
CREATE TABLE %s.agents (
  slug VARCHAR(100) PRIMARY KEY,
  launch_command TEXT,
  executable TEXT,
  adapter_id VARCHAR(100),
  supported_modes TEXT,
  agentfile_source TEXT,
  updated_at TIMESTAMPTZ
);
INSERT INTO %s.agents(
  slug, launch_command, executable, adapter_id, supported_modes, agentfile_source, updated_at
) VALUES (
  'cursor-cli', 'cursor-agent', 'cursor-agent', 'cursor-pty', 'pty', 'legacy', NOW()
);
`, shadowSchema, emptySchema, shadowSchema, shadowSchema))
		require.NoError(t, err)
		_, err = conn.ExecContext(context.Background(), fmt.Sprintf(`
SELECT set_config(
  'search_path',
  '%s,%s,public',
  false
);
`, emptySchema, shadowSchema))
		require.NoError(t, err)

		err = applyMigration000231ErrConn(context.Background(), conn)
		_, _ = conn.ExecContext(context.Background(), `ROLLBACK`)

		require.ErrorContains(t, err, "current schema agents table is required")
		got := queryString(t, db, fmt.Sprintf(`
SELECT adapter_id FROM %s.agents WHERE slug = 'cursor-cli'
`, shadowSchema))
		require.Equal(t, "cursor-pty", got)
	})

	t.Run("failed repair is atomic", func(t *testing.T) {
		dsn := newMigrationLineageSchema(t)
		migrateTo229(t, dsn)
		db := openMigrationDB(t, dsn)
		defer db.Close()
		breakPost229ResourceContract(t, dsn)
		_, err := db.Exec(`
ALTER TABLE agents DROP CONSTRAINT IF EXISTS agents_adapter_id_check;
INSERT INTO agents(
  slug, name, launch_command, executable, adapter_id, is_builtin, is_active,
  supported_modes, agentfile_source, created_at, updated_at
) VALUES (
  'custom-agent', 'Custom', 'custom', 'custom', 'Bad_Adapter', false, true,
  'pty', 'AGENT custom', NOW(), NOW()
);
`)
		require.NoError(t, err)

		instance := newPostgresMigrator(t, FS, dsn)
		err = instance.Migrate(231)
		closePostgresMigrator(t, instance)

		require.ErrorContains(t, err, "agent adapter data cannot be repaired deterministically")
		requireMigrationVersion(t, dsn, 231, true)
		require.Equal(t, "cursor-pty", queryString(t, db, `
SELECT adapter_id FROM agents WHERE slug = 'cursor-cli'
`))
		require.Equal(t, "Bad_Adapter", queryString(t, db, `
SELECT adapter_id FROM agents WHERE slug = 'custom-agent'
`))
		require.False(t, queryBool(t, db, `
SELECT worker_spec_model_binding_is_valid(
  '{"resource_id":1,"resource_revision":1,"connection_id":1,"connection_revision":1,"provider_key":"openai","protocol_adapter":"openai-chat","model_id":"gpt"}'::JSONB
)
`))
	})
}

func migratePost229LineageRepair(t *testing.T, dsn string) {
	t.Helper()
	instance := newPostgresMigrator(t, FS, dsn)
	require.NoError(t, instance.Migrate(231))
	closePostgresMigrator(t, instance)
	requireMigrationVersion(t, dsn, 231, false)
}

func migrateTo229(t *testing.T, dsn string) {
	t.Helper()
	instance := newPostgresMigrator(t, FS, dsn)
	require.NoError(t, instance.Migrate(229))
	closePostgresMigrator(t, instance)
	requireMigrationVersion(t, dsn, 229, false)
}

func openMigrationDB(t *testing.T, dsn string) *sql.DB {
	t.Helper()
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	return db
}

func applyMigration000231ErrConn(ctx context.Context, conn *sql.Conn) error {
	raw, err := FS.ReadFile("000231_repair_post_229_resource_lineage.up.sql")
	if err != nil {
		return err
	}
	_, err = conn.ExecContext(ctx, strings.TrimSpace(string(raw)))
	return err
}

func queryString(t *testing.T, db *sql.DB, query string) string {
	t.Helper()
	var got string
	require.NoError(t, db.QueryRow(query).Scan(&got))
	return got
}

func queryBool(t *testing.T, db *sql.DB, query string) bool {
	t.Helper()
	var got bool
	require.NoError(t, db.QueryRow(query).Scan(&got))
	return got
}
