BEGIN;

DELETE FROM organization_agent_configs WHERE agent_slug = 'seedance-expert';
DELETE FROM organization_agents WHERE agent_slug = 'seedance-expert';
DELETE FROM user_agent_configs WHERE agent_slug = 'seedance-expert';
DELETE FROM agents WHERE slug = 'seedance-expert';

DO $bridge$
DECLARE
  legacy_lineage BOOLEAN;
BEGIN
  SELECT col_description('agents'::REGCLASS, attnum) =
    'agentsmesh-lineage:legacy-000209-bridged-at-000210'
  FROM pg_attribute
  WHERE attrelid = 'agents'::REGCLASS
    AND attname = 'adapter_id'
    AND NOT attisdropped
  INTO legacy_lineage;

  IF COALESCE(legacy_lineage, FALSE) THEN
    UPDATE agents
    SET
      launch_command = 'cursor-agent',
      executable = 'cursor-agent',
      adapter_id = 'cursor-pty',
      supported_modes = 'pty',
      agentfile_source = E'# === Identity ===\nAGENT cursor-agent\nEXECUTABLE cursor-agent\n\n# === Mode ===\nMODE pty\n\n# === Environment ===\nENV CURSOR_API_KEY SECRET OPTIONAL\n\n# === Prompt ===\nPROMPT_POSITION prepend\n\nCAPABILITY resume none\nCAPABILITY permission acp\nCAPABILITY usage live\nCAPABILITY interrupt true\nCAPABILITY streaming true\nCAPABILITY model_family multi\n',
      updated_at = NOW()
    WHERE slug = 'cursor-cli';

    EXECUTE $function$
      CREATE OR REPLACE FUNCTION worker_spec_model_binding_is_valid(binding JSONB)
      RETURNS BOOLEAN
      LANGUAGE SQL
      IMMUTABLE
      AS $body$
        SELECT CASE
          WHEN binding IS NULL OR jsonb_typeof(binding) <> 'object' THEN FALSE
          WHEN NOT (binding ?& ARRAY[
            'resource_id', 'resource_revision', 'connection_id',
            'connection_revision', 'provider_key', 'model_id'
          ]) THEN FALSE
          WHEN binding - ARRAY[
            'resource_id', 'resource_revision', 'connection_id',
            'connection_revision', 'provider_key', 'model_id'
          ]::TEXT[] <> '{}'::JSONB THEN FALSE
          ELSE
            worker_spec_jsonb_is_positive_int64(binding->'resource_id')
            AND worker_spec_jsonb_is_positive_int64(binding->'resource_revision')
            AND worker_spec_jsonb_is_positive_int64(binding->'connection_id')
            AND worker_spec_jsonb_is_positive_int64(binding->'connection_revision')
            AND jsonb_typeof(binding->'provider_key') = 'string'
            AND char_length(binding->>'provider_key') BETWEEN 2 AND 100
            AND binding->>'provider_key' ~ '^[a-z0-9]+(-[a-z0-9]+)*$'
            AND jsonb_typeof(binding->'model_id') = 'string'
            AND btrim(binding->>'model_id') <> ''
        END
      $body$
    $function$;

    ALTER TABLE agents
      DROP CONSTRAINT agents_adapter_id_check,
      DROP COLUMN adapter_id;
  END IF;
END
$bridge$;

COMMIT;
