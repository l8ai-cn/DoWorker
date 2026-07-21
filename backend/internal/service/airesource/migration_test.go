package airesource

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/l8ai-cn/agentcloud/backend/internal/testkit"
	"github.com/l8ai-cn/agentcloud/backend/pkg/crypto"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestLegacyMigratorMigratesAIModelsAndVirtualKeys(t *testing.T) {
	db, cipher := migrationDB(t)
	creds := encryptJSON(t, cipher, map[string]string{"api_key": "sk-test"})
	require.NoError(t, db.Exec(
		`INSERT INTO ai_models
		 (id, organization_id, name, provider_type, model, base_url, encrypted_credentials, is_default, is_enabled)
		 VALUES (42, 1, 'Team GPT', 'openai', 'gpt-5', 'https://api.openai.com/v1', ?, true, true)`,
		creds,
	).Error)
	require.NoError(t, db.Exec(
		`INSERT INTO virtual_api_keys
		 (id, organization_id, user_id, ai_model_id, name, key_prefix, key_hash, status)
		 VALUES (7, 1, 1, 42, 'build', 'dwk_test', 'hash', 'active')`,
	).Error)

	report, err := NewLegacyMigrator(db, cipher, 1).Run(context.Background())
	require.NoError(t, err)
	require.Equal(t, 1, report.AIModelsMigrated)
	require.Equal(t, 1, report.VirtualKeysRemapped)

	var modelID, mappedID int64
	require.NoError(t, db.Raw(`SELECT id FROM model_resources WHERE id = 42`).Scan(&modelID).Error)
	require.Equal(t, int64(42), modelID)
	require.NoError(t, db.Raw(`SELECT model_resource_id FROM virtual_api_keys WHERE id = 7`).Scan(&mappedID).Error)
	require.Equal(t, int64(42), mappedID)

	var connection migrationConnectionRow
	require.NoError(t, db.Where("provider_key = ?", "openai").First(&connection).Error)
	require.Equal(t, "org", connection.OwnerScope)
	require.Equal(t, int64(1), connection.OwnerID)
	require.Equal(t, "https://api.openai.com/v1", connection.BaseURL)
	require.Equal(t, migrationStringList{"api_key"}, connection.ConfiguredFields)

	plain, err := cipher.Decrypt(connection.CredentialsEncrypted)
	require.NoError(t, err)
	require.JSONEq(t, `{"api_key":"sk-test"}`, plain)

	var defaultResourceID int64
	require.NoError(t, db.Raw(
		`SELECT model_resource_id FROM model_resource_defaults
		  WHERE owner_scope = 'org' AND owner_id = 1 AND modality = 'chat'`,
	).Scan(&defaultResourceID).Error)
	require.Equal(t, int64(42), defaultResourceID)

	report, err = NewLegacyMigrator(db, cipher, 1).Run(context.Background())
	require.NoError(t, err)
	require.Zero(t, report.AIModelsMigrated)
	require.Zero(t, report.VirtualKeysRemapped)
}

func TestLegacyMigratorMigratesCredentialEnvBundle(t *testing.T) {
	db, cipher := migrationDB(t)
	data := map[string]string{
		"ANTHROPIC_API_KEY": mustEncrypt(t, cipher, "sk-ant"),
		"ANTHROPIC_MODEL":   mustEncrypt(t, cipher, "claude-sonnet"),
	}
	encoded, err := json.Marshal(data)
	require.NoError(t, err)
	require.NoError(t, db.Exec(
		`INSERT INTO env_bundles
		 (id, owner_scope, owner_id, agent_slug, name, kind, kind_primary, data, is_active)
		 VALUES (9, 'user', 1, 'claude-code', 'claude-work', 'credential', true, ?, true)`,
		string(encoded),
	).Error)

	report, err := NewLegacyMigrator(db, cipher, 1).Run(context.Background())
	require.NoError(t, err)
	require.Equal(t, 1, report.EnvBundlesMigrated)

	var modelID string
	require.NoError(t, db.Raw(
		`SELECT model_id FROM model_resources
		  WHERE id = (SELECT model_resource_id FROM ai_resource_migration_map
		              WHERE source_kind = 'env_bundle' AND source_id = 9)`,
	).Scan(&modelID).Error)
	require.Equal(t, "claude-sonnet", modelID)

	var connection migrationConnectionRow
	require.NoError(t, db.Where("provider_key = ?", "anthropic").First(&connection).Error)
	require.Equal(t, "user", connection.OwnerScope)
	require.Equal(t, int64(1), connection.OwnerID)
	require.Equal(t, "https://api.anthropic.com", connection.BaseURL)
	require.Equal(t, migrationStringList{"api_key"}, connection.ConfiguredFields)

	var defaultResourceID int64
	require.NoError(t, db.Raw(
		`SELECT model_resource_id FROM model_resource_defaults
		  WHERE owner_scope = 'user' AND owner_id = 1 AND modality = 'chat'`,
	).Scan(&defaultResourceID).Error)
	require.NotZero(t, defaultResourceID)
}

func TestLegacyMigratorFailsClosedOnUnknownAIModelProvider(t *testing.T) {
	db, cipher := migrationDB(t)
	creds := encryptJSON(t, cipher, map[string]string{"api_key": "sk-test"})
	require.NoError(t, db.Exec(
		`INSERT INTO ai_models
		 (id, user_id, name, provider_type, model, encrypted_credentials, is_default, is_enabled)
		 VALUES (6, 1, 'Unknown', 'unknown-provider', 'model-x', ?, false, true)`,
		creds,
	).Error)

	_, err := NewLegacyMigrator(db, cipher, 1).Run(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "unknown provider")
	assertNoMigratedRows(t, db)
	assertLegacyAIModelStillExists(t, db, 6)
}

func TestLegacyMigratorFailsClosedOnUnknownCredentialBundleAgent(t *testing.T) {
	db, cipher := migrationDB(t)
	data := map[string]string{
		"OPENAI_API_KEY": mustEncrypt(t, cipher, "sk-openai"),
		"OPENAI_MODEL":   mustEncrypt(t, cipher, "gpt-5"),
	}
	encoded, err := json.Marshal(data)
	require.NoError(t, err)
	require.NoError(t, db.Exec(
		`INSERT INTO env_bundles
		 (id, owner_scope, owner_id, agent_slug, name, kind, kind_primary, data, is_active)
		 VALUES (10, 'user', 1, 'unknown-agent', 'unknown-agent-key', 'credential', false, ?, true)`,
		string(encoded),
	).Error)

	_, err = NewLegacyMigrator(db, cipher, 1).Run(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported agent")
	assertNoMigratedRows(t, db)
	assertLegacyEnvBundleStillExists(t, db, 10)
}

func TestLegacyMigratorFailsClosedOnCorruptCredentialBundleCiphertext(t *testing.T) {
	db, cipher := migrationDB(t)
	data := map[string]string{
		"OPENAI_API_KEY": "not-ciphertext",
		"OPENAI_MODEL":   mustEncrypt(t, cipher, "gpt-5"),
	}
	encoded, err := json.Marshal(data)
	require.NoError(t, err)
	require.NoError(t, db.Exec(
		`INSERT INTO env_bundles
		 (id, owner_scope, owner_id, agent_slug, name, kind, kind_primary, data, is_active)
		 VALUES (11, 'user', 1, 'codex-cli', 'broken-key', 'credential', false, ?, true)`,
		string(encoded),
	).Error)

	_, err = NewLegacyMigrator(db, cipher, 1).Run(context.Background())
	require.Error(t, err)
	assertNoMigratedRows(t, db)
	assertLegacyEnvBundleStillExists(t, db, 11)
}

func TestLegacyMigratorFailsClosedOnCorruptCiphertext(t *testing.T) {
	db, cipher := migrationDB(t)
	require.NoError(t, db.Exec(
		`INSERT INTO ai_models
		 (id, user_id, name, provider_type, model, encrypted_credentials, is_default, is_enabled)
		 VALUES (5, 1, 'Broken', 'openai', 'gpt-5', 'not-base64', false, true)`,
	).Error)

	_, err := NewLegacyMigrator(db, cipher, 1).Run(context.Background())
	require.Error(t, err)

	assertNoMigratedRows(t, db)
	assertLegacyAIModelStillExists(t, db, 5)
}

func TestLegacyMigratorCheckReportsDirtyAndCleanStates(t *testing.T) {
	db, cipher := migrationDB(t)
	creds := encryptJSON(t, cipher, map[string]string{"api_key": "sk-test"})
	require.NoError(t, db.Exec(
		`INSERT INTO ai_models
		 (id, user_id, name, provider_type, model, encrypted_credentials, is_default, is_enabled)
		 VALUES (11, 1, 'User GPT', 'openai', 'gpt-5', ?, false, true)`,
		creds,
	).Error)

	report, err := NewLegacyMigrator(db, cipher, 1).Check(context.Background())
	require.NoError(t, err)
	require.False(t, report.Clean())
	require.Equal(t, 1, report.UnmigratedAIModels)

	_, err = NewLegacyMigrator(db, cipher, 1).Run(context.Background())
	require.NoError(t, err)
	report, err = NewLegacyMigrator(db, cipher, 1).Check(context.Background())
	require.NoError(t, err)
	require.True(t, report.Clean())
}

func TestLegacyMigratorCheckReportsUnmappedVirtualKeys(t *testing.T) {
	db, cipher := migrationDB(t)
	require.NoError(t, db.Exec(
		`INSERT INTO virtual_api_keys
		 (id, organization_id, user_id, ai_model_id, name, key_prefix, key_hash, status)
		 VALUES (8, 1, 1, 99, 'build', 'dwk_test2', 'hash2', 'active')`,
	).Error)

	report, err := NewLegacyMigrator(db, cipher, 1).Check(context.Background())
	require.NoError(t, err)
	require.False(t, report.Clean())
	require.Equal(t, 1, report.UnmappedVirtualKeys)
}

func TestLegacyMigratorCheckReportsUnmappedVirtualKeysBeforeCutoverColumn(t *testing.T) {
	db, cipher := migrationDB(t)
	require.NoError(t, db.Exec(`ALTER TABLE virtual_api_keys DROP COLUMN model_resource_id`).Error)
	require.NoError(t, db.Exec(
		`INSERT INTO virtual_api_keys
		 (id, organization_id, user_id, ai_model_id, name, key_prefix, key_hash, status)
		 VALUES (9, 1, 1, 99, 'build', 'dwk_test3', 'hash3', 'active')`,
	).Error)

	report, err := NewLegacyMigrator(db, cipher, 1).Check(context.Background())
	require.NoError(t, err)
	require.False(t, report.Clean())
	require.Equal(t, 1, report.UnmappedVirtualKeys)
}

func TestLegacyMigratorCheckReportsScopeAndFieldMismatches(t *testing.T) {
	db, cipher := migrationDB(t)
	creds := encryptJSON(t, cipher, map[string]string{"api_key": "sk-test"})
	require.NoError(t, db.Exec(
		`INSERT INTO ai_models
		 (id, organization_id, name, provider_type, model, base_url, encrypted_credentials, is_default, is_enabled)
		 VALUES (13, 1, 'Team GPT', 'openai', 'gpt-5', 'https://api.openai.com/v1', ?, false, true)`,
		creds,
	).Error)
	_, err := NewLegacyMigrator(db, cipher, 1).Run(context.Background())
	require.NoError(t, err)
	require.NoError(t, db.Exec(`INSERT INTO users(id, email, username) VALUES (99, 'other@example.com', 'other-user')`).Error)
	wrongCreds, err := cipher.Encrypt(`{"api_key":"sk-test"}`)
	require.NoError(t, err)
	require.NoError(t, db.Exec(
		`INSERT INTO provider_connections
		 (owner_scope, owner_id, identifier, provider_key, name, base_url, credentials_encrypted, configured_fields, status, is_enabled, created_by)
		 VALUES ('user', 99, 'wrong-scope', 'openai', 'Wrong', 'https://wrong.example', ?, '["api_key"]', 'valid', true, 1)`,
		wrongCreds,
	).Error)
	var wrongConnectionID int64
	require.NoError(t, db.Raw(`SELECT id FROM provider_connections WHERE identifier = 'wrong-scope'`).Scan(&wrongConnectionID).Error)
	require.NoError(t, db.Exec(
		`UPDATE ai_resource_migration_map SET provider_connection_id = ?
		  WHERE source_kind = 'ai_model' AND source_id = 13`,
		wrongConnectionID,
	).Error)
	require.NoError(t, db.Exec(`UPDATE model_resources SET model_id = 'wrong-model' WHERE id = 13`).Error)

	report, err := NewLegacyMigrator(db, cipher, 1).Check(context.Background())
	require.NoError(t, err)
	require.False(t, report.Clean())
	require.Equal(t, 1, report.ScopeMismatches)
	require.Equal(t, 1, report.FieldMismatches)
}

func TestLegacyMigratorCheckReportsBrokenMappingsWithoutOverwritingIDMismatch(t *testing.T) {
	db, cipher := migrationDB(t)
	creds := encryptJSON(t, cipher, map[string]string{"api_key": "sk-test"})
	require.NoError(t, db.Exec(
		`INSERT INTO ai_models
		 (id, user_id, name, provider_type, model, encrypted_credentials, is_default, is_enabled)
		 VALUES (14, 1, 'User GPT', 'openai', 'gpt-5', ?, false, true)`,
		creds,
	).Error)
	_, err := NewLegacyMigrator(db, cipher, 1).Run(context.Background())
	require.NoError(t, err)
	require.NoError(t, db.Exec(
		`INSERT INTO ai_resource_migration_map
		 (source_kind, source_id, provider_connection_id, model_resource_id, status)
		 VALUES ('ai_model', 15, 9999, 9999, 'migrated')`,
	).Error)
	require.NoError(t, db.Exec(
		`UPDATE ai_resource_migration_map SET model_resource_id = 140
		  WHERE source_kind = 'ai_model' AND source_id = 14`,
	).Error)

	report, err := NewLegacyMigrator(db, cipher, 1).Check(context.Background())
	require.NoError(t, err)
	require.False(t, report.Clean())
	require.Equal(t, 3, report.BrokenMappings)
}

func TestLegacyMigratorCheckReportsDecryptFailures(t *testing.T) {
	db, cipher := migrationDB(t)
	require.NoError(t, db.Exec(
		`INSERT INTO ai_models
		 (id, user_id, name, provider_type, model, encrypted_credentials, is_default, is_enabled)
		 VALUES (12, 1, 'Broken', 'openai', 'gpt-5', 'not-base64', false, true)`,
	).Error)

	report, err := NewLegacyMigrator(db, cipher, 1).Check(context.Background())
	require.NoError(t, err)
	require.False(t, report.Clean())
	require.Equal(t, 1, report.DecryptFailures)
}

func migrationDB(t *testing.T) (*gorm.DB, *crypto.Encryptor) {
	t.Helper()
	db := testkit.SetupTestDB(t)
	require.NoError(t, db.Exec(`INSERT INTO users(id, email, username) VALUES (1, 'u@example.com', 'user-one')`).Error)
	require.NoError(t, db.Exec(`INSERT INTO organizations(id, name, slug) VALUES (1, 'Acme', 'acme')`).Error)
	require.NoError(t, db.Exec(legacyAIModelsDDL).Error)
	require.NoError(t, db.Exec(legacyVirtualKeysDDL).Error)
	require.NoError(t, db.Exec(`ALTER TABLE virtual_api_keys ADD COLUMN model_resource_id INTEGER`).Error)
	return db, crypto.NewEncryptor("migration-test-key")
}

const legacyAIModelsDDL = `CREATE TABLE IF NOT EXISTS ai_models (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	organization_id INTEGER, user_id INTEGER,
	name TEXT NOT NULL, provider_type TEXT NOT NULL, model TEXT NOT NULL,
	base_url TEXT NOT NULL DEFAULT '', encrypted_credentials TEXT NOT NULL DEFAULT '',
	is_default INTEGER NOT NULL DEFAULT 0, is_enabled INTEGER NOT NULL DEFAULT 1
)`

const legacyVirtualKeysDDL = `CREATE TABLE IF NOT EXISTS virtual_api_keys (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	organization_id INTEGER NOT NULL, user_id INTEGER NOT NULL, ai_model_id INTEGER NOT NULL,
	name TEXT NOT NULL, key_prefix TEXT NOT NULL, key_hash TEXT NOT NULL, status TEXT NOT NULL
)`

func encryptJSON(t *testing.T, cipher *crypto.Encryptor, values map[string]string) string {
	t.Helper()
	encoded, err := json.Marshal(values)
	require.NoError(t, err)
	return mustEncrypt(t, cipher, string(encoded))
}

func mustEncrypt(t *testing.T, cipher *crypto.Encryptor, value string) string {
	t.Helper()
	encrypted, err := cipher.Encrypt(value)
	require.NoError(t, err)
	return encrypted
}

func assertNoMigratedRows(t *testing.T, db *gorm.DB) {
	t.Helper()
	for _, table := range []string{"provider_connections", "model_resources", "ai_resource_migration_map"} {
		var count int64
		require.NoError(t, db.Table(table).Count(&count).Error)
		require.Zero(t, count, table)
	}
}

func assertLegacyAIModelStillExists(t *testing.T, db *gorm.DB, id int64) {
	t.Helper()
	var count int64
	require.NoError(t, db.Table("ai_models").Where("id = ?", id).Count(&count).Error)
	require.Equal(t, int64(1), count)
}

func assertLegacyEnvBundleStillExists(t *testing.T, db *gorm.DB, id int64) {
	t.Helper()
	var count int64
	require.NoError(t, db.Table("env_bundles").Where("id = ?", id).Count(&count).Error)
	require.Equal(t, int64(1), count)
}
