ALTER TABLE pods DROP CONSTRAINT IF EXISTS pods_pending_config_revision_id_fkey;
ALTER TABLE pods DROP CONSTRAINT IF EXISTS pods_active_config_revision_id_fkey;
ALTER TABLE pods DROP CONSTRAINT IF EXISTS pods_archived_by_id_fkey;
ALTER TABLE pods DROP CONSTRAINT IF EXISTS pods_model_resource_id_fkey;
ALTER TABLE pods DROP CONSTRAINT IF EXISTS pods_generation_nonnegative;

DROP INDEX IF EXISTS idx_pod_config_revisions_one_active_per_pod;
DROP INDEX IF EXISTS idx_pod_config_revisions_model_resource_id;
DROP INDEX IF EXISTS idx_pod_config_revisions_status;
DROP INDEX IF EXISTS idx_pod_config_revisions_pod_id;
DROP INDEX IF EXISTS idx_pods_pending_config_revision_id;
DROP INDEX IF EXISTS idx_pods_active_config_revision_id;
DROP INDEX IF EXISTS idx_pods_model_resource_id;

DROP TABLE IF EXISTS pod_config_revisions;

ALTER TABLE pods DROP COLUMN IF EXISTS purge_after;
ALTER TABLE pods DROP COLUMN IF EXISTS archived_by_id;
ALTER TABLE pods DROP COLUMN IF EXISTS archived_at;
ALTER TABLE pods DROP COLUMN IF EXISTS reinitialize_dispatched_at;
ALTER TABLE pods DROP COLUMN IF EXISTS pending_config_revision_id;
ALTER TABLE pods DROP COLUMN IF EXISTS active_config_revision_id;
ALTER TABLE pods DROP COLUMN IF EXISTS generation;
ALTER TABLE pods DROP COLUMN IF EXISTS model_resource_id;
