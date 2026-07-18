BEGIN;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM workflow_runs
    WHERE orchestration_resource_id IS NOT NULL
      AND finished_at IS NULL
  ) THEN
    RAISE EXCEPTION
      'workflow_runs contain active resource-native runs; drain them before rollback';
  END IF;
END
$$;

ALTER TABLE workflow_runs
  DROP CONSTRAINT IF EXISTS workflow_runs_execution_manifest_check,
  DROP COLUMN IF EXISTS execution_manifest;

COMMIT;
