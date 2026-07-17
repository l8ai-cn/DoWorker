DO $validation$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_schema = current_schema()
      AND table_name = 'agents'
      AND column_name = 'adapter_id'
      AND is_nullable = 'NO'
  ) OR NOT EXISTS (
    SELECT 1 FROM pg_constraint
    WHERE conrelid = 'agents'::REGCLASS
      AND conname = 'agents_adapter_id_check'
      AND convalidated
  ) THEN
    RAISE EXCEPTION 'agent adapter migration contract is incomplete';
  END IF;

  IF EXISTS (
    SELECT 1 FROM agents
    WHERE adapter_id IS NULL
      OR adapter_id !~ '^[a-z0-9]+(-[a-z0-9]+)*$'
      OR char_length(adapter_id) NOT BETWEEN 2 AND 100
  ) THEN
    RAISE EXCEPTION 'agent adapter data violates the migration contract';
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM agents
    WHERE slug = 'cursor-cli'
      AND launch_command = 'agent'
      AND executable = 'agent'
      AND adapter_id = 'cursor-acp'
      AND supported_modes = 'pty,acp'
  ) THEN
    RAISE EXCEPTION 'cursor-cli migration contract is incomplete';
  END IF;

  IF to_regprocedure('worker_spec_model_binding_is_valid(jsonb)') IS NULL
    OR NOT worker_spec_model_binding_is_valid(
      '{"resource_id":1,"resource_revision":1,"connection_id":1,"connection_revision":1,"provider_key":"openai","protocol_adapter":"openai-chat","model_id":"gpt"}'::JSONB
    )
  THEN
    RAISE EXCEPTION 'worker spec protocol adapter contract is incomplete';
  END IF;

  IF (
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
  ) <> 7
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
  THEN
    RAISE EXCEPTION 'goal loop migration contract is incomplete';
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM agents
    WHERE slug = 'seedance-expert'
      AND adapter_id = 'do-agent-acp'
      AND agentfile_source LIKE '%/do-agent-home%'
  ) OR NOT EXISTS (
    SELECT 1 FROM agents
    WHERE slug = 'video-studio'
      AND adapter_id = 'codex-app-server'
  ) THEN
    RAISE EXCEPTION 'Seedance or Video Studio migration contract is incomplete';
  END IF;
END
$validation$;
