package migrations

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/stretchr/testify/require"
)

func TestMigrationLineageUpgradePostgres(t *testing.T) {
	t.Run("fresh database", func(t *testing.T) {
		dsn := newMigrationLineageSchema(t)
		instance := newPostgresMigrator(t, FS, dsn)
		require.NoError(t, instance.Up())
		closePostgresMigrator(t, instance)

		requireCurrentMigrationContract(t, dsn, false)
	})

	t.Run("legacy clean 000209 is blocked before forward repair", func(t *testing.T) {
		dsn := newMigrationLineageSchema(t)
		legacy := newPostgresMigrator(t, legacy000209MigrationFS(t), dsn)
		require.NoError(t, legacy.Up())
		closePostgresMigrator(t, legacy)
		requireLegacy000209Contract(t, dsn)

		bridge := newPostgresMigrator(t, FS, dsn)
		require.Error(t, bridge.Migrate(214))
		closePostgresMigrator(t, bridge)
		requireMigrationVersion(t, dsn, 210, true)
	})

	t.Run("current clean 000222", func(t *testing.T) {
		dsn := newMigrationLineageSchema(t)
		before := newPostgresMigrator(t, FS, dsn)
		require.NoError(t, before.Migrate(222))
		closePostgresMigrator(t, before)
		requireMigrationVersion(t, dsn, 222, false)

		after := newPostgresMigrator(t, FS, dsn)
		require.NoError(t, after.Up())
		closePostgresMigrator(t, after)

		requireCurrentMigrationContract(t, dsn, false)
	})

	t.Run("current clean 000229 forward repair is retained across down up", func(t *testing.T) {
		dsn := newMigrationLineageSchema(t)
		before := newPostgresMigrator(t, FS, dsn)
		require.NoError(t, before.Migrate(229))
		closePostgresMigrator(t, before)
		requireMigrationVersion(t, dsn, 229, false)
		breakPost229ResourceContract(t, dsn)

		after := newPostgresMigrator(t, FS, dsn)
		require.NoError(t, after.Up())
		closePostgresMigrator(t, after)
		requireCurrentMigrationContract(t, dsn, false)

		rollback := newPostgresMigrator(t, FS, dsn)
		require.NoError(t, rollback.Steps(-2))
		closePostgresMigrator(t, rollback)
		requireMigrationVersion(t, dsn, 229, false)

		reapply := newPostgresMigrator(t, FS, dsn)
		require.NoError(t, reapply.Up())
		closePostgresMigrator(t, reapply)
		requireCurrentMigrationContract(t, dsn, false)
	})
}

func TestMigrationLineageRejectsDirty000222Postgres(t *testing.T) {
	dsn := newMigrationLineageSchema(t)
	before := newPostgresMigrator(t, FS, dsn)
	require.NoError(t, before.Migrate(221))
	closePostgresMigrator(t, before)

	probe := newPostgresMigrator(t, failing000222MigrationFS(t), dsn)
	require.Error(t, probe.Up())
	closePostgresMigrator(t, probe)
	requireMigrationVersion(t, dsn, 222, true)

	current := newPostgresMigrator(t, FS, dsn)
	err := current.Up()
	var dirty migrate.ErrDirty
	require.True(t, errors.As(err, &dirty))
	require.Equal(t, 222, dirty.Version)
	closePostgresMigrator(t, current)

	requireMigrationVersion(t, dsn, 222, true)
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer db.Close()
	var videoStudioExists bool
	require.NoError(t, db.QueryRow(`
SELECT EXISTS (SELECT 1 FROM agents WHERE slug = 'video-studio')
`).Scan(&videoStudioExists))
	require.False(t, videoStudioExists)
	var seedanceSource string
	require.NoError(t, db.QueryRow(`
SELECT agentfile_source FROM agents WHERE slug = 'seedance-expert'
`).Scan(&seedanceSource))
	require.NotContains(t, seedanceSource, "/do-agent-home")
}

func requireLegacy000209Contract(t *testing.T, dsn string) {
	t.Helper()
	requireMigrationVersion(t, dsn, 209, false)
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer db.Close()

	var adapterColumnExists bool
	require.NoError(t, db.QueryRow(`
SELECT EXISTS (
  SELECT 1 FROM information_schema.columns
  WHERE table_schema = current_schema()
    AND table_name = 'agents'
    AND column_name = 'adapter_id'
)
`).Scan(&adapterColumnExists))
	require.False(t, adapterColumnExists)

	var bindingValid bool
	require.NoError(t, db.QueryRow(`
SELECT worker_spec_model_binding_is_valid(
  '{"resource_id":1,"resource_revision":1,"connection_id":1,"connection_revision":1,"provider_key":"openai","protocol_adapter":"openai-chat","model_id":"gpt"}'::JSONB
)
`).Scan(&bindingValid))
	require.False(t, bindingValid)

	var loopColumnCount int
	require.NoError(t, db.QueryRow(`
SELECT count(*)
FROM information_schema.columns
WHERE table_schema = current_schema()
  AND table_name = 'goal_loops'
  AND column_name IN (
    'current_iteration', 'no_progress_count', 'same_error_count',
    'last_progress_fingerprint', 'last_error_fingerprint',
    'retry_prompt_command_id', 'retry_prompt_created_at'
  )
`).Scan(&loopColumnCount))
	require.Equal(t, 7, loopColumnCount)
}

func breakPost229ResourceContract(t *testing.T, dsn string) {
	t.Helper()
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer db.Close()
	_, err = db.Exec(`
ALTER TABLE agents DROP CONSTRAINT IF EXISTS agents_adapter_id_check;
UPDATE agents
SET launch_command = 'cursor-agent',
    executable = 'cursor-agent',
    adapter_id = 'cursor-pty',
    supported_modes = 'pty'
WHERE slug = 'cursor-cli';
UPDATE agents
SET agentfile_source = replace(agentfile_source, '/do-agent-home', '/seedance-expert-home')
WHERE slug = 'seedance-expert';
CREATE OR REPLACE FUNCTION worker_spec_model_binding_is_valid(binding JSONB)
RETURNS BOOLEAN
LANGUAGE SQL
IMMUTABLE
AS $$
    SELECT FALSE
$$;
`)
	require.NoError(t, err)
}

func requireCurrentMigrationContract(t *testing.T, dsn string, legacyBridge bool) {
	t.Helper()
	requireMigrationVersion(t, dsn, latestMigrationVersion, false)
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer db.Close()

	var cursorAdapter, cursorModes string
	require.NoError(t, db.QueryRow(`
SELECT adapter_id, supported_modes FROM agents WHERE slug = 'cursor-cli'
`).Scan(&cursorAdapter, &cursorModes))
	require.Equal(t, "cursor-acp", cursorAdapter)
	require.Equal(t, "pty,acp", cursorModes)

	var bindingValid bool
	require.NoError(t, db.QueryRow(`
SELECT worker_spec_model_binding_is_valid(
  '{"resource_id":1,"resource_revision":1,"connection_id":1,"connection_revision":1,"provider_key":"openai","protocol_adapter":"openai-chat","model_id":"gpt"}'::JSONB
)
`).Scan(&bindingValid))
	require.True(t, bindingValid)

	var lineageMarker sql.NullString
	require.NoError(t, db.QueryRow(`
SELECT col_description('agents'::REGCLASS, attnum)
FROM pg_attribute
WHERE attrelid = 'agents'::REGCLASS
  AND attname = 'adapter_id'
  AND NOT attisdropped
`).Scan(&lineageMarker))
	require.Equal(t, legacyBridge, lineageMarker.Valid)
	if lineageMarker.Valid {
		require.Equal(
			t,
			"agentsmesh-lineage:legacy-000209-bridged-at-000210",
			lineageMarker.String,
		)
	}

	var seedanceAdapter, seedanceSource, videoAdapter string
	require.NoError(t, db.QueryRow(`
SELECT adapter_id, agentfile_source FROM agents WHERE slug = 'seedance-expert'
`).Scan(&seedanceAdapter, &seedanceSource))
	require.Equal(t, "do-agent-acp", seedanceAdapter)
	require.Contains(t, seedanceSource, "/do-agent-home")
	require.NoError(t, db.QueryRow(`
SELECT adapter_id FROM agents WHERE slug = 'video-studio'
`).Scan(&videoAdapter))
	require.Equal(t, "codex-app-server", videoAdapter)
}

func requireMigrationVersion(t *testing.T, dsn string, version int, dirty bool) {
	t.Helper()
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer db.Close()
	var gotVersion int
	var gotDirty bool
	require.NoError(t, db.QueryRow(`
SELECT version, dirty FROM schema_migrations
`).Scan(&gotVersion, &gotDirty))
	require.Equal(t, version, gotVersion)
	require.Equal(t, dirty, gotDirty)
}
