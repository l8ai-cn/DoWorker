package airesource

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/l8ai-cn/agentcloud/backend/pkg/crypto"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestLegacyMigratorAdvancesPostgresModelResourceSequence(t *testing.T) {
	dsn := os.Getenv("AI_RESOURCE_MIGRATION_POSTGRES_TEST_DSN")
	if dsn == "" {
		t.Skip("AI_RESOURCE_MIGRATION_POSTGRES_TEST_DSN is not configured")
	}
	admin, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	schema := fmt.Sprintf("ai_resource_migration_%d", time.Now().UnixNano())
	require.NoError(t, admin.Exec(`CREATE SCHEMA `+schema).Error)
	t.Cleanup(func() {
		_ = admin.Exec(`DROP SCHEMA IF EXISTS ` + schema + ` CASCADE`).Error
	})

	db, err := gorm.Open(
		postgres.Open(postgresDSNWithSearchPath(dsn, schema)),
		&gorm.Config{Logger: logger.Default.LogMode(logger.Silent)},
	)
	require.NoError(t, err)
	require.NoError(t, db.Exec(postgresMigrationTestDDL).Error)
	require.NoError(t, db.Exec(`INSERT INTO users(id) VALUES (1)`).Error)
	require.NoError(t, db.Exec(`INSERT INTO organizations(id) VALUES (1)`).Error)

	cipher := crypto.NewEncryptor("migration-postgres-test-key")
	creds := encryptJSON(t, cipher, map[string]string{"api_key": "sk-test"})
	require.NoError(t, db.Exec(
		`INSERT INTO ai_models
		 (id, organization_id, name, provider_type, model, encrypted_credentials, is_default, is_enabled)
		 VALUES (1, 1, 'Team GPT', 'openai', 'gpt-5', ?, false, true)`,
		creds,
	).Error)
	bundleData, err := json.Marshal(map[string]string{
		"OPENAI_API_KEY": mustEncrypt(t, cipher, "sk-bundle"),
		"OPENAI_MODEL":   mustEncrypt(t, cipher, "gpt-5-mini"),
	})
	require.NoError(t, err)
	require.NoError(t, db.Exec(
		`INSERT INTO env_bundles
		 (id, owner_scope, owner_id, agent_slug, name, kind, kind_primary, data, is_active)
		 VALUES (1, 'user', 1, 'codex-cli', 'bundle', 'credential', false, ?, true)`,
		string(bundleData),
	).Error)

	report, err := NewLegacyMigrator(db, cipher, 1).Run(context.Background())
	require.NoError(t, err)
	require.Equal(t, 1, report.AIModelsMigrated)
	require.Equal(t, 1, report.EnvBundlesMigrated)

	var maxID int64
	require.NoError(t, db.Raw(`SELECT max(id) FROM model_resources`).Scan(&maxID).Error)
	var nextID int64
	require.NoError(t, db.Raw(
		`INSERT INTO model_resources
		 (provider_connection_id, identifier, model_id, display_name, modalities, capabilities, status, is_enabled)
		 SELECT id, 'post-migration', 'gpt-next', 'Next', '["chat"]', '["text-generation"]', 'valid', true
		   FROM provider_connections
		  ORDER BY id
		  LIMIT 1
		 RETURNING id`,
	).Scan(&nextID).Error)
	require.Greater(t, nextID, maxID)
}

func postgresDSNWithSearchPath(dsn, schema string) string {
	if strings.Contains(dsn, "://") {
		parsed, err := url.Parse(dsn)
		if err == nil {
			query := parsed.Query()
			query.Set("search_path", schema)
			parsed.RawQuery = query.Encode()
			return parsed.String()
		}
	}
	return dsn + " search_path=" + schema
}

const postgresMigrationTestDDL = `
CREATE TABLE users (id BIGINT PRIMARY KEY);
CREATE TABLE organizations (id BIGINT PRIMARY KEY);
CREATE TABLE provider_connections (
	id BIGSERIAL PRIMARY KEY, owner_scope TEXT NOT NULL, owner_id BIGINT NOT NULL,
	identifier TEXT NOT NULL, provider_key TEXT NOT NULL, name TEXT NOT NULL,
	base_url TEXT NOT NULL DEFAULT '', credentials_encrypted TEXT NOT NULL DEFAULT '',
	configured_fields JSONB NOT NULL DEFAULT '[]', status TEXT NOT NULL,
	is_enabled BOOLEAN NOT NULL, created_by BIGINT NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	UNIQUE(owner_scope, owner_id, identifier)
);
CREATE TABLE model_resources (
	id BIGSERIAL PRIMARY KEY, provider_connection_id BIGINT NOT NULL,
	identifier TEXT NOT NULL, model_id TEXT NOT NULL, display_name TEXT NOT NULL,
	modalities JSONB NOT NULL, capabilities JSONB NOT NULL, status TEXT NOT NULL,
	is_enabled BOOLEAN NOT NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	UNIQUE(provider_connection_id, identifier)
);
CREATE TABLE model_resource_defaults (
	owner_scope TEXT NOT NULL, owner_id BIGINT NOT NULL, modality TEXT NOT NULL,
	model_resource_id BIGINT NOT NULL, PRIMARY KEY(owner_scope, owner_id, modality)
);
CREATE TABLE ai_resource_migration_map (
	id BIGSERIAL PRIMARY KEY, source_kind TEXT NOT NULL, source_id BIGINT NOT NULL,
	provider_connection_id BIGINT, model_resource_id BIGINT, status TEXT NOT NULL,
	error_message TEXT NOT NULL DEFAULT '', created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), UNIQUE(source_kind, source_id)
);
CREATE TABLE ai_models (
	id BIGSERIAL PRIMARY KEY, organization_id BIGINT, user_id BIGINT,
	name TEXT NOT NULL, provider_type TEXT NOT NULL, model TEXT NOT NULL,
	base_url TEXT NOT NULL DEFAULT '', encrypted_credentials TEXT NOT NULL DEFAULT '',
	is_default BOOLEAN NOT NULL, is_enabled BOOLEAN NOT NULL
);
CREATE TABLE env_bundles (
	id BIGSERIAL PRIMARY KEY, owner_scope TEXT NOT NULL, owner_id BIGINT NOT NULL,
	agent_slug TEXT, name TEXT NOT NULL, kind TEXT NOT NULL, kind_primary BOOLEAN NOT NULL,
	data JSONB NOT NULL, is_active BOOLEAN NOT NULL
);
`
