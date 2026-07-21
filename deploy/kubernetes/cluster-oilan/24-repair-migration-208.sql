\set ON_ERROR_STOP on

BEGIN;
SELECT pg_advisory_xact_lock(hashtext('agentcloud-schema-migration-208-repair'));
\ir repair-preconditions.sql

SELECT NOT EXISTS (
  SELECT 1 FROM information_schema.columns
  WHERE table_schema = 'public'
    AND table_name = 'agents'
    AND column_name = 'adapter_id'
) AS apply_207 \gset

\if :apply_207
-- BEGIN 000207
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
END
WHERE slug IN (
  'aider', 'claude-code', 'codex-cli', 'cursor-cli', 'do-agent',
  'gemini-cli', 'grok-build', 'loopal', 'minimax-cli', 'opencode'
);

ALTER TABLE agents
  ALTER COLUMN adapter_id SET NOT NULL,
  ADD CONSTRAINT agents_adapter_id_check
    CHECK (
      adapter_id ~ '^[a-z0-9]+(-[a-z0-9]+)*$'
      AND char_length(adapter_id) BETWEEN 2 AND 100
    );
-- END 000207

-- BEGIN 000208
UPDATE agents
SET
  launch_command = 'agent',
  executable = 'agent',
  adapter_id = 'cursor-acp',
  supported_modes = 'pty,acp',
  agentfile_source = E'# === Identity ===\nAGENT agent\nEXECUTABLE agent\n\n# === Mode ===\nMODE pty\nMODE acp "acp"\n\n# === Environment ===\nENV CURSOR_API_KEY SECRET OPTIONAL\n\n# === Prompt ===\nPROMPT_POSITION prepend\n\nCAPABILITY resume none\nCAPABILITY permission acp\nCAPABILITY usage live\nCAPABILITY interrupt true\nCAPABILITY streaming true\nCAPABILITY model_family multi\n',
  updated_at = NOW()
WHERE slug = 'cursor-cli';
-- END 000208
\endif

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_schema = 'public'
      AND table_name = 'agents'
      AND column_name = 'adapter_id'
      AND data_type = 'character varying'
      AND character_maximum_length = 100
      AND is_nullable = 'NO'
  ) THEN
    RAISE EXCEPTION 'migration 207 adapter_id column contract is incomplete';
  END IF;
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conrelid = 'agents'::regclass
      AND conname = 'agents_adapter_id_check'
      AND md5(pg_get_constraintdef(oid)) =
        '6945d5c8ae2c98789d3768284673ec6d'
  ) THEN
    RAISE EXCEPTION 'migration 207 adapter_id check contract is incomplete';
  END IF;
  IF EXISTS (
    SELECT 1
    FROM agents
    WHERE adapter_id IS DISTINCT FROM CASE slug
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
      ELSE adapter_id
    END
  ) THEN
    RAISE EXCEPTION 'migration 207 adapter mapping is incomplete';
  END IF;
  IF NOT EXISTS (
    SELECT 1
    FROM agents
    WHERE slug = 'cursor-cli'
      AND launch_command = 'agent'
      AND executable = 'agent'
      AND adapter_id = 'cursor-acp'
      AND supported_modes = 'pty,acp'
      AND agentfile_source = E'# === Identity ===\nAGENT agent\nEXECUTABLE agent\n\n# === Mode ===\nMODE pty\nMODE acp "acp"\n\n# === Environment ===\nENV CURSOR_API_KEY SECRET OPTIONAL\n\n# === Prompt ===\nPROMPT_POSITION prepend\n\nCAPABILITY resume none\nCAPABILITY permission acp\nCAPABILITY usage live\nCAPABILITY interrupt true\nCAPABILITY streaming true\nCAPABILITY model_family multi\n'
  ) THEN
    RAISE EXCEPTION 'migration 208 cursor upgrade did not match one agent';
  END IF;
END $$;

COMMIT;
