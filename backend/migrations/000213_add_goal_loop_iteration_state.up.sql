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
    AND column_name IN (
      'current_iteration',
      'no_progress_count',
      'same_error_count',
      'last_progress_fingerprint',
      'last_error_fingerprint'
    )
  INTO existing_columns;

  SELECT count(*)
  FROM information_schema.columns
  WHERE table_schema = current_schema()
    AND table_name = 'goal_loops'
    AND (
      (
        column_name IN ('current_iteration', 'no_progress_count', 'same_error_count')
        AND data_type = 'integer' AND is_nullable = 'NO' AND column_default = '0'
      )
      OR (
        column_name IN ('last_progress_fingerprint', 'last_error_fingerprint')
        AND data_type = 'character varying' AND is_nullable = 'YES'
        AND character_maximum_length = 64
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
      ADD COLUMN current_iteration INTEGER NOT NULL DEFAULT 0,
      ADD COLUMN no_progress_count INTEGER NOT NULL DEFAULT 0,
      ADD COLUMN same_error_count INTEGER NOT NULL DEFAULT 0,
      ADD COLUMN last_progress_fingerprint VARCHAR(64),
      ADD COLUMN last_error_fingerprint VARCHAR(64),
      ADD CONSTRAINT chk_goal_loops_iteration_state CHECK (
        current_iteration >= 0
        AND no_progress_count >= 0
        AND same_error_count >= 0
      );
  ELSIF existing_columns <> 5
    OR valid_columns <> 5
    OR NOT COALESCE(legacy_lineage, FALSE)
    OR NOT EXISTS (
      SELECT 1 FROM pg_constraint
      WHERE conrelid = 'goal_loops'::REGCLASS
        AND conname = 'chk_goal_loops_iteration_state'
        AND convalidated
    )
  THEN
    RAISE EXCEPTION
      'goal loop iteration state is partially applied or has unknown ownership'
      USING HINT = 'Restore a clean migration version; do not force past 000213.';
  END IF;
END
$migration$;
