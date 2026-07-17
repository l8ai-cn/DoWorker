DO $migration$
DECLARE
  existing_columns INTEGER;
  valid_columns INTEGER;
  legacy_lineage BOOLEAN;
BEGIN
  SELECT count(*)
  FROM information_schema.columns
  WHERE table_schema = current_schema()
    AND table_name = 'goal_loops'
    AND column_name IN ('retry_prompt_command_id', 'retry_prompt_created_at')
  INTO existing_columns;

  SELECT count(*)
  FROM information_schema.columns
  WHERE table_schema = current_schema()
    AND table_name = 'goal_loops'
    AND (
      (
        column_name = 'retry_prompt_command_id'
        AND data_type = 'character varying' AND is_nullable = 'YES'
        AND character_maximum_length = 64
      )
      OR (
        column_name = 'retry_prompt_created_at'
        AND data_type = 'timestamp with time zone' AND is_nullable = 'YES'
      )
    )
  INTO valid_columns;

  SELECT col_description('agents'::REGCLASS, attnum) =
    'agentsmesh-lineage:legacy-000209-bridged-at-000210'
  FROM pg_attribute
  WHERE attrelid = 'agents'::REGCLASS
    AND attname = 'adapter_id'
    AND NOT attisdropped
  INTO legacy_lineage;

  IF existing_columns = 0 THEN
    ALTER TABLE goal_loops
      ADD COLUMN retry_prompt_command_id VARCHAR(64),
      ADD COLUMN retry_prompt_created_at TIMESTAMPTZ,
      ADD CONSTRAINT chk_goal_loops_retry_prompt_state CHECK (
        (retry_prompt_command_id IS NULL AND retry_prompt_created_at IS NULL)
        OR (retry_prompt_command_id IS NOT NULL AND retry_prompt_created_at IS NOT NULL)
      );

    CREATE INDEX idx_goal_loops_retry_prompt_pending
      ON goal_loops (id)
      WHERE status = 'verifying' AND retry_prompt_command_id IS NOT NULL;
  ELSIF existing_columns <> 2
    OR valid_columns <> 2
    OR NOT COALESCE(legacy_lineage, FALSE)
    OR to_regclass('idx_goal_loops_retry_prompt_pending') IS NULL
    OR NOT EXISTS (
      SELECT 1 FROM pg_constraint
      WHERE conrelid = 'goal_loops'::REGCLASS
        AND conname = 'chk_goal_loops_retry_prompt_state'
        AND convalidated
    )
  THEN
    RAISE EXCEPTION
      'goal loop retry state is partially applied or has unknown ownership'
      USING HINT = 'Restore a clean migration version; do not force past 000214.';
  END IF;
END
$migration$;
