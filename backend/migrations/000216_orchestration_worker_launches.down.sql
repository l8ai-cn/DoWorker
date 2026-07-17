ALTER TABLE orchestration_worker_launches
  DROP CONSTRAINT IF EXISTS orchestration_worker_launches_pod_fkey;

DROP INDEX IF EXISTS idx_pods_orchestration_worker_launch;

ALTER TABLE pods
  DROP CONSTRAINT IF EXISTS pods_orchestration_worker_launch_fkey,
  DROP COLUMN IF EXISTS orchestration_worker_launch_id;

DROP INDEX IF EXISTS idx_orchestration_worker_launches_pending;
DROP TABLE IF EXISTS orchestration_worker_launches;

ALTER TABLE pods
  DROP CONSTRAINT IF EXISTS pods_org_id_key_unique;

ALTER TABLE orchestration_resource_plans
  DROP CONSTRAINT IF EXISTS orchestration_resource_plans_org_id_unique;
