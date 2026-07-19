BEGIN;

LOCK TABLE coordinator_projects IN ACCESS EXCLUSIVE MODE;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM coordinator_projects
    WHERE worker_spec_snapshot_id IS NOT NULL
  ) THEN
    RAISE EXCEPTION 'coordinator projects have worker spec snapshot bindings; clear them before rollback';
  END IF;
END $$;

DROP INDEX IF EXISTS idx_coordinator_projects_worker_spec_snapshot_id;

ALTER TABLE coordinator_projects
  DROP CONSTRAINT IF EXISTS coordinator_projects_worker_spec_snapshot_positive,
  DROP CONSTRAINT IF EXISTS coordinator_projects_worker_spec_snapshot_id_fkey,
  DROP COLUMN IF EXISTS worker_spec_snapshot_id;

COMMIT;
