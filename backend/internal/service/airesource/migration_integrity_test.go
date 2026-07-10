package airesource

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLegacyMigratorFailsClosedOnAmbiguousAIModelOwner(t *testing.T) {
	db, cipher := migrationDB(t)
	creds := encryptJSON(t, cipher, map[string]string{"api_key": "sk-test"})
	require.NoError(t, db.Exec(
		`INSERT INTO ai_models
		 (id, organization_id, user_id, name, provider_type, model, encrypted_credentials, is_default, is_enabled)
		 VALUES (16, 1, 1, 'Ambiguous', 'openai', 'gpt-5', ?, false, true)`,
		creds,
	).Error)

	_, err := NewLegacyMigrator(db, cipher, 1).Run(context.Background())

	require.ErrorContains(t, err, "exactly one owner")
	assertNoMigratedRows(t, db)
	assertLegacyAIModelStillExists(t, db, 16)
}

func TestLegacyMigratorCheckRejectsCrossConnectionMapping(t *testing.T) {
	db, cipher := migrationDB(t)
	creds := encryptJSON(t, cipher, map[string]string{"api_key": "sk-test"})
	require.NoError(t, db.Exec(
		`INSERT INTO ai_models
		 (id, organization_id, name, provider_type, model, encrypted_credentials, is_default, is_enabled)
		 VALUES (17, 1, 'Team GPT', 'openai', 'gpt-5', ?, false, true)`,
		creds,
	).Error)
	_, err := NewLegacyMigrator(db, cipher, 1).Run(context.Background())
	require.NoError(t, err)

	var original migrationConnectionRow
	require.NoError(t, db.Where("identifier = ?", "ai-model-17").First(&original).Error)
	duplicate := original
	duplicate.ID = 0
	duplicate.Identifier = "ai-model-17-copy"
	require.NoError(t, db.Create(&duplicate).Error)
	require.NoError(t, db.Exec(
		`UPDATE ai_resource_migration_map
		    SET provider_connection_id = ?
		  WHERE source_kind = 'ai_model' AND source_id = 17`,
		duplicate.ID,
	).Error)

	report, err := NewLegacyMigrator(db, cipher, 1).Check(context.Background())

	require.NoError(t, err)
	require.False(t, report.Clean())
	require.Equal(t, 1, report.BrokenMappings)
}

func TestLegacyMigratorCheckRejectsCanonicalFieldDrift(t *testing.T) {
	db, cipher := migrationDB(t)
	creds := encryptJSON(t, cipher, map[string]string{"api_key": "sk-test"})
	require.NoError(t, db.Exec(
		`INSERT INTO ai_models
		 (id, organization_id, name, provider_type, model, encrypted_credentials, is_default, is_enabled)
		 VALUES (18, 1, 'Team GPT', 'openai', 'gpt-5', ?, true, true)`,
		creds,
	).Error)
	_, err := NewLegacyMigrator(db, cipher, 1).Run(context.Background())
	require.NoError(t, err)
	require.NoError(t, db.Exec(
		`UPDATE provider_connections
		    SET name = 'Changed', is_enabled = false
		  WHERE id = (
		    SELECT provider_connection_id FROM ai_resource_migration_map
		     WHERE source_kind = 'ai_model' AND source_id = 18
		  )`,
	).Error)
	require.NoError(t, db.Exec(
		`UPDATE model_resources
		    SET display_name = 'Changed', is_enabled = false
		  WHERE id = 18`,
	).Error)
	require.NoError(t, db.Exec(
		`DELETE FROM model_resource_defaults WHERE model_resource_id = 18`,
	).Error)

	report, err := NewLegacyMigrator(db, cipher, 1).Check(context.Background())

	require.NoError(t, err)
	require.False(t, report.Clean())
	require.Equal(t, 1, report.FieldMismatches)
}

func TestLegacyMigratorIsIdempotentAfterCutoverDropsAIModelID(t *testing.T) {
	db, cipher := migrationDB(t)
	creds := encryptJSON(t, cipher, map[string]string{"api_key": "sk-test"})
	require.NoError(t, db.Exec(
		`INSERT INTO ai_models
		 (id, organization_id, name, provider_type, model, encrypted_credentials, is_default, is_enabled)
		 VALUES (19, 1, 'Team GPT', 'openai', 'gpt-5', ?, false, true)`,
		creds,
	).Error)
	require.NoError(t, db.Exec(
		`INSERT INTO virtual_api_keys
		 (id, organization_id, user_id, ai_model_id, name, key_prefix, key_hash, status)
		 VALUES (19, 1, 1, 19, 'build', 'dwk_cutover', 'hash-cutover', 'active')`,
	).Error)
	_, err := NewLegacyMigrator(db, cipher, 1).Run(context.Background())
	require.NoError(t, err)
	require.NoError(t, db.Exec(`ALTER TABLE virtual_api_keys DROP COLUMN ai_model_id`).Error)

	report, err := NewLegacyMigrator(db, cipher, 1).Run(context.Background())

	require.NoError(t, err)
	require.Zero(t, report.AIModelsMigrated)
	require.Zero(t, report.VirtualKeysRemapped)
}
