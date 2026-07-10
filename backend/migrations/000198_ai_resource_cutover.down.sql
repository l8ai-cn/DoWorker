ALTER TABLE virtual_api_keys ADD COLUMN IF NOT EXISTS ai_model_id BIGINT;

UPDATE virtual_api_keys
   SET ai_model_id = migration.source_id
  FROM ai_resource_migration_map migration
 WHERE migration.source_kind = 'ai_model'
   AND migration.model_resource_id = virtual_api_keys.model_resource_id
   AND migration.status = 'migrated'
   AND virtual_api_keys.ai_model_id IS NULL;

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM virtual_api_keys WHERE ai_model_id IS NULL) THEN
        RAISE EXCEPTION 'ai_model_id cannot be restored from AI resource migration map';
    END IF;
END $$;

ALTER TABLE virtual_api_keys ALTER COLUMN ai_model_id SET NOT NULL;

ALTER TABLE virtual_api_keys
    ADD CONSTRAINT virtual_api_keys_ai_model_id_fkey
    FOREIGN KEY (ai_model_id) REFERENCES ai_models(id) ON DELETE CASCADE
    NOT VALID;
ALTER TABLE virtual_api_keys VALIDATE CONSTRAINT virtual_api_keys_ai_model_id_fkey;

CREATE INDEX IF NOT EXISTS idx_virtual_api_keys_model
    ON virtual_api_keys(ai_model_id);

DROP INDEX IF EXISTS idx_virtual_api_keys_model_resource;
ALTER TABLE virtual_api_keys DROP CONSTRAINT IF EXISTS virtual_api_keys_model_resource_id_fkey;
ALTER TABLE virtual_api_keys DROP COLUMN IF EXISTS model_resource_id;
