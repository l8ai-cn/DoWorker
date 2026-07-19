BEGIN;
DO $migration$
DECLARE
  agents_table REGCLASS := to_regclass(format('%I.agents', current_schema()));
  invalid_adapter_data BOOLEAN;
BEGIN
  IF agents_table IS NULL THEN
    RAISE EXCEPTION 'current schema agents table is required before resource lineage repair';
  END IF;
  IF NOT EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_schema = current_schema()
      AND table_name = 'agents'
      AND column_name = 'adapter_id'
  ) THEN
    EXECUTE format('ALTER TABLE %s ADD COLUMN adapter_id VARCHAR(100)', agents_table);
  END IF;
  EXECUTE format($sql$
    UPDATE %s
    SET adapter_id = CASE slug
      WHEN 'aider' THEN 'aider-pty'
      WHEN 'claude-code' THEN 'claude-stream-json'
      WHEN 'codex-cli' THEN 'codex-app-server'
      WHEN 'cursor-cli' THEN 'cursor-acp'
      WHEN 'do-agent' THEN 'do-agent-acp'
      WHEN 'gemini-cli' THEN 'gemini-acp'
      WHEN 'grok-build' THEN 'grok-build-acp'
      WHEN 'loopal' THEN 'loopal-acp'
      WHEN 'minimax-cli' THEN 'minimax-pty'
      WHEN 'opencode' THEN 'opencode-acp'
      WHEN 'seedance-expert' THEN 'do-agent-acp'
      WHEN 'video-studio' THEN 'codex-app-server'
    END
    WHERE slug IN (
      'aider', 'claude-code', 'codex-cli', 'cursor-cli', 'do-agent',
      'gemini-cli', 'grok-build', 'loopal', 'minimax-cli', 'opencode',
      'seedance-expert', 'video-studio'
    )
  $sql$, agents_table);
  EXECUTE format($sql$
    UPDATE %s
    SET
      launch_command = 'agent',
      executable = 'agent',
      supported_modes = 'pty,acp',
      agentfile_source = E'# === Identity ===\nAGENT agent\nEXECUTABLE agent\n\n# === Mode ===\nMODE pty\nMODE acp "acp"\n\n# === Environment ===\nENV CURSOR_API_KEY SECRET OPTIONAL\n\n# === Prompt ===\nPROMPT_POSITION prepend\n\nCAPABILITY resume none\nCAPABILITY permission acp\nCAPABILITY usage live\nCAPABILITY interrupt true\nCAPABILITY streaming true\nCAPABILITY model_family multi\n',
      updated_at = NOW()
    WHERE slug = 'cursor-cli'
  $sql$, agents_table);
  EXECUTE format($sql$
    UPDATE %s
    SET agentfile_source = replace(agentfile_source, '/seedance-expert-home', '/do-agent-home'),
        updated_at = NOW()
    WHERE slug = 'seedance-expert'
      AND agentfile_source LIKE '%%/seedance-expert-home%%'
  $sql$, agents_table);
  EXECUTE format($sql$
    SELECT EXISTS (
      SELECT 1
      FROM ONLY %s
      WHERE adapter_id IS NULL
        OR adapter_id !~ '^[a-z0-9]+(-[a-z0-9]+)*$'
        OR char_length(adapter_id) NOT BETWEEN 2 AND 100
    )
  $sql$, agents_table)
  INTO invalid_adapter_data;
  IF invalid_adapter_data THEN
    RAISE EXCEPTION 'agent adapter data cannot be repaired deterministically';
  END IF;
  EXECUTE format('ALTER TABLE %s DROP CONSTRAINT IF EXISTS agents_adapter_id_check', agents_table);
  EXECUTE format('ALTER TABLE %s ALTER COLUMN adapter_id SET NOT NULL', agents_table);
  EXECUTE format($sql$
    ALTER TABLE %s
    ADD CONSTRAINT agents_adapter_id_check
    CHECK (
      adapter_id ~ '^[a-z0-9]+(-[a-z0-9]+)*$'
      AND char_length(adapter_id) BETWEEN 2 AND 100
    )
  $sql$, agents_table);
END
$migration$;
CREATE OR REPLACE FUNCTION worker_spec_model_binding_is_valid(binding JSONB) RETURNS BOOLEAN LANGUAGE SQL IMMUTABLE AS $$
    SELECT CASE
        WHEN binding = '{}'::JSONB THEN TRUE
        WHEN binding IS NULL OR jsonb_typeof(binding) <> 'object' THEN FALSE
        WHEN NOT (
            binding ?& ARRAY[
                'resource_id',
                'resource_revision',
                'connection_id',
                'connection_revision',
                'provider_key',
                'model_id'
            ]
        ) THEN FALSE
        WHEN binding - ARRAY[
            'resource_id',
            'resource_revision',
            'connection_id',
            'connection_revision',
            'provider_key',
            'protocol_adapter',
            'model_id'
        ]::TEXT[] <> '{}'::JSONB THEN FALSE
        ELSE
            worker_spec_jsonb_is_positive_int64(binding->'resource_id')
            AND worker_spec_jsonb_is_positive_int64(binding->'resource_revision')
            AND worker_spec_jsonb_is_positive_int64(binding->'connection_id')
            AND worker_spec_jsonb_is_positive_int64(binding->'connection_revision')
            AND jsonb_typeof(binding->'provider_key') = 'string'
            AND char_length(binding->>'provider_key') BETWEEN 2 AND 100
            AND binding->>'provider_key' ~ '^[a-z0-9]+(-[a-z0-9]+)*$'
            AND (
                NOT binding ? 'protocol_adapter'
                OR (
                    jsonb_typeof(binding->'protocol_adapter') = 'string'
                    AND char_length(binding->>'protocol_adapter') BETWEEN 2 AND 100
                    AND binding->>'protocol_adapter' ~ '^[a-z0-9]+(-[a-z0-9]+)*$'
                )
            )
            AND jsonb_typeof(binding->'model_id') = 'string'
            AND btrim(binding->>'model_id') <> ''
    END
$$;
DO $validation$
DECLARE
  agents_table REGCLASS := to_regclass(format('%I.agents', current_schema()));
  constraint_ready BOOLEAN;
  cursor_repaired BOOLEAN;
  seedance_repaired BOOLEAN;
  video_repaired BOOLEAN;
BEGIN
  IF agents_table IS NULL THEN
    RAISE EXCEPTION 'current schema agents table is required before resource lineage validation';
  END IF;
  EXECUTE format($sql$
    SELECT EXISTS (
      SELECT 1
      FROM pg_constraint
      WHERE conrelid = %L::regclass
        AND conname = 'agents_adapter_id_check'
        AND convalidated
    )
  $sql$, agents_table::TEXT)
  INTO constraint_ready;
  IF NOT constraint_ready THEN
    RAISE EXCEPTION 'agent adapter constraint repair is incomplete';
  END IF;
  EXECUTE format($sql$
    SELECT EXISTS (
      SELECT 1
      FROM ONLY %s
      WHERE slug = 'cursor-cli'
        AND launch_command = 'agent'
        AND executable = 'agent'
        AND adapter_id = 'cursor-acp'
        AND supported_modes = 'pty,acp'
    )
  $sql$, agents_table)
  INTO cursor_repaired;
  IF NOT cursor_repaired THEN
    RAISE EXCEPTION 'cursor-cli resource lineage repair is incomplete';
  END IF;
  IF NOT worker_spec_model_binding_is_valid(
    '{"resource_id":1,"resource_revision":1,"connection_id":1,"connection_revision":1,"provider_key":"openai","protocol_adapter":"openai-chat","model_id":"gpt"}'::JSONB
  ) OR NOT worker_spec_model_binding_is_valid('{}'::JSONB) THEN
    RAISE EXCEPTION 'worker spec model binding validator repair is incomplete';
  END IF;
  EXECUTE format($sql$
    SELECT NOT EXISTS (
      SELECT 1
      FROM ONLY %s
      WHERE slug = 'seedance-expert'
        AND adapter_id = 'do-agent-acp'
        AND agentfile_source LIKE '%%/do-agent-home%%'
    )
  $sql$, agents_table)
  INTO seedance_repaired;
  IF seedance_repaired THEN
    RAISE EXCEPTION 'seedance resource lineage repair is incomplete';
  END IF;
  EXECUTE format($sql$
    SELECT NOT EXISTS (
      SELECT 1
      FROM ONLY %s
      WHERE slug = 'video-studio'
        AND adapter_id = 'codex-app-server'
    )
  $sql$, agents_table)
  INTO video_repaired;
  IF video_repaired THEN
    RAISE EXCEPTION 'video studio resource lineage repair is incomplete';
  END IF;
END
$validation$;
COMMIT;
