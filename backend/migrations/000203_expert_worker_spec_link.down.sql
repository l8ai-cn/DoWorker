ALTER TABLE experts
  DROP CONSTRAINT IF EXISTS experts_worker_spec_snapshot_org_fkey;

DROP INDEX IF EXISTS idx_experts_worker_spec_snapshot_id;

ALTER TABLE experts
  DROP COLUMN IF EXISTS worker_spec_snapshot_id;
