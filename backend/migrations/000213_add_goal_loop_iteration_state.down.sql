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
    ALTER TABLE goal_loops
      DROP CONSTRAINT IF EXISTS chk_goal_loops_iteration_state,
      DROP COLUMN IF EXISTS last_error_fingerprint,
      DROP COLUMN IF EXISTS last_progress_fingerprint,
      DROP COLUMN IF EXISTS same_error_count,
      DROP COLUMN IF EXISTS no_progress_count,
      DROP COLUMN IF EXISTS current_iteration;
  END IF;
END
$migration$;
