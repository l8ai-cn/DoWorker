DO $bridge$
DECLARE
  has_adapter_id BOOLEAN; binding_accepts_protocol_adapter BOOLEAN;
  legacy_loop_columns INTEGER; unresolved_agents TEXT;
BEGIN
  SELECT EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_schema = current_schema()
      AND table_name = 'agents'
      AND column_name = 'adapter_id'
  ) INTO has_adapter_id;
  IF to_regprocedure('worker_spec_model_binding_is_valid(jsonb)') IS NULL THEN
    RAISE EXCEPTION
      'migration lineage is invalid: worker_spec_model_binding_is_valid(jsonb) is missing'
      USING HINT = 'Restore the database to a clean migration version before retrying.';
  END IF;
  SELECT worker_spec_model_binding_is_valid(
    '{"resource_id":1,"resource_revision":1,"connection_id":1,"connection_revision":1,"provider_key":"openai","protocol_adapter":"openai-chat","model_id":"gpt"}'::JSONB
  ) INTO binding_accepts_protocol_adapter;
  IF NOT has_adapter_id THEN
    SELECT count(*)
    FROM information_schema.columns
    WHERE table_schema = current_schema()
      AND table_name = 'goal_loops'
      AND column_name IN (
        'current_iteration',
        'no_progress_count',
        'same_error_count',
        'last_progress_fingerprint',
        'last_error_fingerprint',
        'retry_prompt_command_id',
        'retry_prompt_created_at'
      )
      AND (
        (
          column_name IN ('current_iteration', 'no_progress_count', 'same_error_count')
          AND data_type = 'integer' AND is_nullable = 'NO' AND column_default = '0'
        )
        OR (
          column_name IN ('last_progress_fingerprint', 'last_error_fingerprint', 'retry_prompt_command_id')
          AND data_type = 'character varying' AND is_nullable = 'YES'
          AND character_maximum_length = 64
        )
        OR (
          column_name = 'retry_prompt_created_at'
          AND data_type = 'timestamp with time zone' AND is_nullable = 'YES'
        )
      )
    INTO legacy_loop_columns;
    IF legacy_loop_columns <> 7
      OR to_regclass('idx_goal_loops_retry_prompt_pending') IS NULL
      OR NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conrelid = 'goal_loops'::REGCLASS
          AND conname = 'chk_goal_loops_iteration_state'
          AND convalidated
      )
      OR NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conrelid = 'goal_loops'::REGCLASS
          AND conname = 'chk_goal_loops_retry_prompt_state'
          AND convalidated
      )
      OR binding_accepts_protocol_adapter
    THEN
      RAISE EXCEPTION
        'migration lineage is neither current 000209 nor supported legacy 000209'
        USING HINT = 'Restore the database to a clean 000209 backup; do not force the migration version.';
    END IF;
    ALTER TABLE agents ADD COLUMN adapter_id VARCHAR(100);
    UPDATE agents
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
    END;
    SELECT string_agg(slug, ', ' ORDER BY slug)
    FROM agents
    WHERE adapter_id IS NULL
    INTO unresolved_agents;
    IF unresolved_agents IS NOT NULL THEN
      RAISE EXCEPTION
        'legacy 000209 bridge cannot infer adapter_id for agents: %',
        unresolved_agents
        USING HINT = 'Assign explicit adapters in an approved repair migration before retrying.';
    END IF;
    ALTER TABLE agents
      ALTER COLUMN adapter_id SET NOT NULL,
      ADD CONSTRAINT agents_adapter_id_check
        CHECK (
          adapter_id ~ '^[a-z0-9]+(-[a-z0-9]+)*$'
          AND char_length(adapter_id) BETWEEN 2 AND 100
        );
    UPDATE agents
    SET
      launch_command = 'agent',
      executable = 'agent',
      adapter_id = 'cursor-acp',
      supported_modes = 'pty,acp',
      agentfile_source = E'# === Identity ===\nAGENT agent\nEXECUTABLE agent\n\n# === Mode ===\nMODE pty\nMODE acp "acp"\n\n# === Environment ===\nENV CURSOR_API_KEY SECRET OPTIONAL\n\n# === Prompt ===\nPROMPT_POSITION prepend\n\nCAPABILITY resume none\nCAPABILITY permission acp\nCAPABILITY usage live\nCAPABILITY interrupt true\nCAPABILITY streaming true\nCAPABILITY model_family multi\n',
      updated_at = NOW()
    WHERE slug = 'cursor-cli';
    COMMENT ON COLUMN agents.adapter_id IS
      'agentsmesh-lineage:legacy-000209-bridged-at-000210';
  ELSIF NOT binding_accepts_protocol_adapter
    OR NOT EXISTS (
      SELECT 1
      FROM information_schema.columns
      WHERE table_schema = current_schema()
        AND table_name = 'agents'
        AND column_name = 'adapter_id'
        AND is_nullable = 'NO'
    )
    OR NOT EXISTS (
      SELECT 1 FROM pg_constraint
      WHERE conrelid = 'agents'::REGCLASS
        AND conname = 'agents_adapter_id_check'
        AND convalidated
    )
    OR NOT EXISTS (
      SELECT 1 FROM agents
      WHERE slug = 'cursor-cli'
        AND launch_command = 'agent'
        AND executable = 'agent'
        AND adapter_id = 'cursor-acp'
        AND supported_modes = 'pty,acp'
    )
  THEN
    RAISE EXCEPTION
      'current 000209 migration contract is incomplete'
      USING HINT = 'Restore the database to a clean migration version before retrying; do not force.';
  END IF;
END
$bridge$;
CREATE OR REPLACE FUNCTION worker_spec_model_binding_is_valid(binding JSONB)
RETURNS BOOLEAN
LANGUAGE SQL
IMMUTABLE
AS $$
  SELECT CASE
    WHEN binding IS NULL OR jsonb_typeof(binding) <> 'object' THEN FALSE
    WHEN NOT (binding ?& ARRAY['resource_id', 'resource_revision', 'connection_id',
      'connection_revision', 'provider_key', 'model_id']) THEN FALSE
    WHEN binding - ARRAY['resource_id', 'resource_revision', 'connection_id',
      'connection_revision', 'provider_key', 'protocol_adapter', 'model_id']::TEXT[]
      <> '{}'::JSONB THEN FALSE
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
INSERT INTO agents (
  slug,
  name,
  description,
  launch_command,
  executable,
  adapter_id,
  is_builtin,
  is_active,
  supported_modes,
  agentfile_source
)
VALUES (
  'seedance-expert',
  'Seedance Expert',
  'Plans, generates, resumes, and reviews Seedance video tasks.',
  'do-agent',
  'do-agent',
  'do-agent-acp',
  true,
  true,
  'pty,acp',
  E'AGENT do-agent\nEXECUTABLE do-agent\n\nMODE pty\nMODE acp "acp"\n\nCONFIG model STRING = ""\n\nENV DO_AGENT_HOME = sandbox.root + "/seedance-expert-home"\nENV DO_AGENT_SETTINGS = sandbox.root + "/seedance-expert-home/settings.json"\nENV DO_AGENT_LOG_DIR = sandbox.root + "/seedance-expert-home/logs"\nENV OPENAI_API_KEY SECRET OPTIONAL\nENV ANTHROPIC_API_KEY SECRET OPTIONAL\nENV SEEDANCE_API_KEY SECRET OPTIONAL\nENV SEEDANCE_BASE_URL TEXT OPTIONAL\nENV SEEDANCE_MODEL TEXT OPTIONAL\n\nPROMPT_POSITION prepend\nMCP ON\n\narg "--model" config.model when config.model != ""\nmkdir sandbox.root + "/seedance-expert-home"\n\nif config_json {\n  file sandbox.root + "/seedance-expert-home/settings.json" json(config_json)\n}\n\nif mcp.enabled {\n  mkdir sandbox.work_dir + "/.agent"\n  file sandbox.work_dir + "/.agent/config.json" json({ mcpServers: mcp.servers })\n}\n\nCAPABILITY resume acp\nCAPABILITY permission notification\nCAPABILITY usage live\nCAPABILITY control set_model,set_execution_mode\nCAPABILITY interrupt true\nCAPABILITY streaming true\nCAPABILITY subagents false\nCAPABILITY model_family multi\n'
);
