DO $migration$
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

  IF NOT COALESCE(legacy_lineage, FALSE) THEN
    DROP INDEX IF EXISTS idx_goal_loops_retry_prompt_pending;

    ALTER TABLE goal_loops
      DROP CONSTRAINT IF EXISTS chk_goal_loops_retry_prompt_state,
      DROP COLUMN IF EXISTS retry_prompt_created_at,
      DROP COLUMN IF EXISTS retry_prompt_command_id;
  END IF;
END
$migration$;
