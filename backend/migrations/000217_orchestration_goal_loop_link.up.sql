ALTER TABLE goal_loops
  DROP CONSTRAINT goal_loops_worker_spec_snapshot_id_fkey,
  ADD COLUMN orchestration_resource_id BIGINT,
  ADD COLUMN orchestration_resource_revision BIGINT;

ALTER TABLE goal_loops
  ADD CONSTRAINT goal_loops_orchestration_mode_check CHECK (
    (orchestration_resource_id IS NULL
      AND orchestration_resource_revision IS NULL)
    OR
    (orchestration_resource_id IS NOT NULL
      AND orchestration_resource_revision > 0
      AND worker_spec_snapshot_id IS NOT NULL)
  ),
  ADD CONSTRAINT goal_loops_worker_spec_snapshot_org_fkey
  FOREIGN KEY (organization_id, worker_spec_snapshot_id)
  REFERENCES worker_spec_snapshots (organization_id, id)
  ON DELETE RESTRICT,
  ADD CONSTRAINT goal_loops_orchestration_revision_fkey
  FOREIGN KEY (
    organization_id,
    orchestration_resource_id,
    orchestration_resource_revision
  )
  REFERENCES orchestration_resource_revisions (
    organization_id,
    resource_id,
    revision
  )
  ON DELETE RESTRICT
  DEFERRABLE INITIALLY DEFERRED;

CREATE UNIQUE INDEX idx_goal_loops_orchestration_resource
  ON goal_loops (organization_id, orchestration_resource_id)
  WHERE orchestration_resource_id IS NOT NULL;
