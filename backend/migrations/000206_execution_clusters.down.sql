DROP INDEX IF EXISTS idx_runner_pending_auths_cluster_id;
ALTER TABLE runner_pending_auths
  DROP CONSTRAINT IF EXISTS runner_pending_auths_cluster_ownership_check,
  DROP CONSTRAINT IF EXISTS runner_pending_auths_cluster_organization_fkey,
  DROP COLUMN IF EXISTS cluster_id,
  ALTER COLUMN authorized DROP NOT NULL;

DROP INDEX IF EXISTS idx_runner_grpc_registration_tokens_cluster_id;
ALTER TABLE runner_grpc_registration_tokens
  DROP CONSTRAINT IF EXISTS runner_grpc_registration_tokens_cluster_organization_fkey,
  DROP COLUMN IF EXISTS cluster_id;

DROP INDEX IF EXISTS idx_pods_cluster_id;
ALTER TABLE pods
  DROP CONSTRAINT IF EXISTS pods_execution_cluster_organization_fkey,
  DROP COLUMN IF EXISTS cluster_id;

DROP INDEX IF EXISTS idx_runners_cluster_id;
ALTER TABLE runners
  DROP CONSTRAINT IF EXISTS runners_execution_cluster_organization_fkey,
  DROP COLUMN IF EXISTS tunnel_last_error,
  DROP COLUMN IF EXISTS tunnel_last_seen_at,
  DROP COLUMN IF EXISTS tunnel_state,
  DROP COLUMN IF EXISTS cluster_id;

DROP TABLE IF EXISTS execution_clusters;
