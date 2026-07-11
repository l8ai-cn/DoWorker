ALTER TABLE worker_spec_snapshots
  ADD CONSTRAINT worker_spec_snapshots_organization_id_id_key
  UNIQUE (organization_id, id);

ALTER TABLE pods
  ADD COLUMN worker_spec_snapshot_id BIGINT;

ALTER TABLE pods
  ADD CONSTRAINT pods_worker_spec_snapshot_org_fkey
  FOREIGN KEY (organization_id, worker_spec_snapshot_id)
  REFERENCES worker_spec_snapshots (organization_id, id)
  ON DELETE RESTRICT;

CREATE INDEX idx_pods_worker_spec_snapshot_id
  ON pods (worker_spec_snapshot_id)
  WHERE worker_spec_snapshot_id IS NOT NULL;
