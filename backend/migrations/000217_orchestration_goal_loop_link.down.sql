DROP INDEX IF EXISTS idx_goal_loops_orchestration_resource;

ALTER TABLE goal_loops
  DROP CONSTRAINT IF EXISTS goal_loops_orchestration_revision_fkey,
  DROP CONSTRAINT IF EXISTS goal_loops_worker_spec_snapshot_org_fkey,
  DROP CONSTRAINT IF EXISTS goal_loops_orchestration_mode_check,
  DROP COLUMN IF EXISTS orchestration_resource_revision,
  DROP COLUMN IF EXISTS orchestration_resource_id;

ALTER TABLE goal_loops
  ADD CONSTRAINT goal_loops_worker_spec_snapshot_id_fkey
  FOREIGN KEY (worker_spec_snapshot_id)
  REFERENCES worker_spec_snapshots (id)
  ON DELETE RESTRICT;
