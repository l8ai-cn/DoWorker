ALTER TABLE pods
  DROP CONSTRAINT IF EXISTS pods_worker_spec_snapshot_org_fkey;

DROP INDEX IF EXISTS idx_pods_worker_spec_snapshot_id;

ALTER TABLE pods
  DROP COLUMN IF EXISTS worker_spec_snapshot_id;

ALTER TABLE worker_spec_snapshots
  DROP CONSTRAINT IF EXISTS worker_spec_snapshots_organization_id_id_key;
