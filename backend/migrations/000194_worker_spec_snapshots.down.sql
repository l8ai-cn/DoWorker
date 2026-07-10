DROP TRIGGER IF EXISTS worker_spec_snapshots_immutable ON worker_spec_snapshots;
DROP FUNCTION IF EXISTS prevent_worker_spec_snapshot_update();
DROP TABLE IF EXISTS worker_spec_snapshots;
