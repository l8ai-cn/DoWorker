DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM ai_models legacy
        WHERE NOT EXISTS (
            SELECT 1 FROM ai_resource_migration_map migration
            WHERE migration.source_kind = 'ai_model'
              AND migration.source_id = legacy.id
              AND migration.status = 'migrated'
              AND migration.model_resource_id IS NOT NULL
        )
    ) THEN
        RAISE EXCEPTION 'ai_models must be migrated before AI resource cutover';
    END IF;

    IF EXISTS (
        SELECT 1 FROM env_bundles legacy
        WHERE legacy.kind = 'credential'
          AND legacy.is_active = TRUE
          AND NOT EXISTS (
              SELECT 1 FROM ai_resource_migration_map migration
              WHERE migration.source_kind = 'env_bundle'
                AND migration.source_id = legacy.id
                AND migration.status = 'migrated'
                AND migration.model_resource_id IS NOT NULL
          )
    ) THEN
        RAISE EXCEPTION 'credential EnvBundles must be migrated before AI resource cutover';
    END IF;
END $$;

ALTER TABLE virtual_api_keys ADD COLUMN IF NOT EXISTS model_resource_id BIGINT;

UPDATE virtual_api_keys
   SET model_resource_id = migration.model_resource_id
  FROM ai_resource_migration_map migration
 WHERE migration.source_kind = 'ai_model'
   AND migration.source_id = virtual_api_keys.ai_model_id
   AND migration.status = 'migrated'
   AND migration.model_resource_id IS NOT NULL
   AND virtual_api_keys.model_resource_id IS NULL;

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM virtual_api_keys WHERE model_resource_id IS NULL) THEN
        RAISE EXCEPTION 'virtual_api_keys must have model_resource_id before AI resource cutover';
    END IF;
END $$;

ALTER TABLE virtual_api_keys ALTER COLUMN model_resource_id SET NOT NULL;

ALTER TABLE virtual_api_keys
    ADD CONSTRAINT virtual_api_keys_model_resource_id_fkey
    FOREIGN KEY (model_resource_id) REFERENCES model_resources(id) ON DELETE CASCADE
    NOT VALID;
ALTER TABLE virtual_api_keys VALIDATE CONSTRAINT virtual_api_keys_model_resource_id_fkey;

CREATE INDEX IF NOT EXISTS idx_virtual_api_keys_model_resource
    ON virtual_api_keys(model_resource_id);

DROP INDEX IF EXISTS idx_virtual_api_keys_model;
ALTER TABLE virtual_api_keys DROP CONSTRAINT IF EXISTS virtual_api_keys_ai_model_id_fkey;
ALTER TABLE virtual_api_keys DROP COLUMN ai_model_id;
