ALTER TABLE coordinator_projects
  ADD COLUMN worker_spec_snapshot_id BIGINT;

ALTER TABLE coordinator_projects
  ADD CONSTRAINT coordinator_projects_worker_spec_snapshot_id_fkey
  FOREIGN KEY (organization_id, worker_spec_snapshot_id)
  REFERENCES worker_spec_snapshots (organization_id, id)
  ON DELETE RESTRICT;

ALTER TABLE coordinator_projects
  ADD CONSTRAINT coordinator_projects_worker_spec_snapshot_positive
  CHECK (worker_spec_snapshot_id IS NULL OR worker_spec_snapshot_id > 0);

CREATE INDEX idx_coordinator_projects_worker_spec_snapshot_id
  ON coordinator_projects (worker_spec_snapshot_id)
  WHERE worker_spec_snapshot_id IS NOT NULL;

COMMENT ON COLUMN coordinator_projects.worker_spec_snapshot_id IS
  'Required immutable worker spec snapshot. NULL legacy projects are disabled until an audited binding is applied.';
