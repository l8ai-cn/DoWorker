ALTER TABLE experts
  ADD COLUMN worker_spec_snapshot_id BIGINT;

ALTER TABLE experts
  ADD CONSTRAINT experts_worker_spec_snapshot_org_fkey
  FOREIGN KEY (organization_id, worker_spec_snapshot_id)
  REFERENCES worker_spec_snapshots (organization_id, id)
  ON DELETE RESTRICT;

CREATE INDEX idx_experts_worker_spec_snapshot_id
  ON experts (worker_spec_snapshot_id)
  WHERE worker_spec_snapshot_id IS NOT NULL;
