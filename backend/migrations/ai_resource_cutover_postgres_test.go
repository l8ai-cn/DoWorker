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

func TestMigration000198AIResourceCutoverPostgres(t *testing.T) {
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

	schema := fmt.Sprintf("ai_resource_cutover_%d", time.Now().UnixNano())
	require.NoError(t, execSQL(ctx, conn, `CREATE SCHEMA `+schema))
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `DROP SCHEMA IF EXISTS `+schema+` CASCADE`)
	})
	require.NoError(t, execSQL(ctx, conn, `SET search_path TO `+schema))
	require.NoError(t, execMigrationSQL(ctx, conn, aiResourceCutoverBaseDDL))
	aiResources, err := FS.ReadFile("000190_ai_resources.up.sql")
	require.NoError(t, err)
	require.NoError(t, execMigrationSQL(ctx, conn, string(aiResources)))
	require.NoError(t, execMigrationSQL(ctx, conn, aiResourceCutoverSeedSQL))

	up, err := FS.ReadFile("000198_ai_resource_cutover.up.sql")
	require.NoError(t, err)
	require.NoError(t, execMigrationSQL(ctx, conn, string(up)))
	require.False(t, postgresColumnExists(ctx, t, conn, "virtual_api_keys", "ai_model_id"))
	require.True(t, postgresColumnExists(ctx, t, conn, "virtual_api_keys", "model_resource_id"))
	var resourceID int64
	require.NoError(t, conn.QueryRowContext(
		ctx,
		`SELECT model_resource_id FROM virtual_api_keys WHERE id = 7`,
	).Scan(&resourceID))
	require.Equal(t, int64(42), resourceID)

	down, err := FS.ReadFile("000198_ai_resource_cutover.down.sql")
	require.NoError(t, err)
	require.NoError(t, execMigrationSQL(ctx, conn, string(down)))
	require.True(t, postgresColumnExists(ctx, t, conn, "virtual_api_keys", "ai_model_id"))
	require.False(t, postgresColumnExists(ctx, t, conn, "virtual_api_keys", "model_resource_id"))
	var modelID int64
	require.NoError(t, conn.QueryRowContext(
		ctx,
		`SELECT ai_model_id FROM virtual_api_keys WHERE id = 7`,
	).Scan(&modelID))
	require.Equal(t, int64(42), modelID)
}

const aiResourceCutoverBaseDDL = `
CREATE TABLE users (id BIGINT PRIMARY KEY);
CREATE TABLE organizations (id BIGINT PRIMARY KEY);
CREATE TABLE ai_models (id BIGINT PRIMARY KEY);
CREATE TABLE env_bundles (
	id BIGINT PRIMARY KEY,
	kind TEXT NOT NULL,
	is_active BOOLEAN NOT NULL
);
CREATE TABLE virtual_api_keys (
	id BIGINT PRIMARY KEY,
	ai_model_id BIGINT NOT NULL REFERENCES ai_models(id) ON DELETE CASCADE
);
INSERT INTO users(id) VALUES (1);
INSERT INTO organizations(id) VALUES (1);
INSERT INTO ai_models(id) VALUES (42);
`

const aiResourceCutoverSeedSQL = `
INSERT INTO provider_connections
	(owner_scope, owner_id, identifier, provider_key, name, configured_fields, status, is_enabled, created_by)
VALUES ('org', 1, 'ai-model-42', 'openai', 'OpenAI', '["api_key"]', 'valid', true, 1);
INSERT INTO model_resources
	(id, provider_connection_id, identifier, model_id, display_name, modalities, capabilities, status, is_enabled)
SELECT 42, id, 'legacy-ai-model-42', 'gpt-5', 'GPT', '["chat"]', '["text-generation"]', 'valid', true
  FROM provider_connections
 WHERE identifier = 'ai-model-42';
INSERT INTO ai_resource_migration_map
	(source_kind, source_id, provider_connection_id, model_resource_id, status)
SELECT 'ai_model', 42, id, 42, 'migrated'
  FROM provider_connections
 WHERE identifier = 'ai-model-42';
INSERT INTO virtual_api_keys(id, ai_model_id) VALUES (7, 42);
`
