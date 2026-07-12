CREATE TABLE execution_clusters (
  id BIGSERIAL PRIMARY KEY,
  organization_id BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  slug VARCHAR(100) NOT NULL,
  name VARCHAR(255) NOT NULL,
  kind VARCHAR(32) NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'pending',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT execution_clusters_org_slug_unique UNIQUE (organization_id, slug),
  CONSTRAINT execution_clusters_id_organization_unique UNIQUE (id, organization_id),
  CONSTRAINT execution_clusters_slug_check
    CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$' AND char_length(slug) BETWEEN 2 AND 100),
  CONSTRAINT execution_clusters_kind_check CHECK (kind IN ('online', 'local')),
  CONSTRAINT execution_clusters_status_check CHECK (status IN ('ready', 'pending', 'offline'))
);

INSERT INTO execution_clusters (organization_id, slug, name, kind, status)
SELECT id, 'online', 'Online cluster', 'online', 'pending'
FROM organizations
ON CONFLICT (organization_id, slug) DO NOTHING;

INSERT INTO execution_clusters (organization_id, slug, name, kind, status)
SELECT id, 'local', 'Local cluster', 'local', 'pending'
FROM organizations
ON CONFLICT (organization_id, slug) DO NOTHING;

ALTER TABLE runners
  ADD COLUMN cluster_id BIGINT,
  ADD COLUMN tunnel_state VARCHAR(32) NOT NULL DEFAULT 'disconnected',
  ADD COLUMN tunnel_last_seen_at TIMESTAMPTZ,
  ADD COLUMN tunnel_last_error VARCHAR(255);

UPDATE runners AS runner
SET cluster_id = cluster.id
FROM execution_clusters AS cluster
WHERE cluster.organization_id = runner.organization_id
  AND cluster.slug = 'online';

ALTER TABLE runners
  ALTER COLUMN cluster_id SET NOT NULL,
  ADD CONSTRAINT runners_execution_cluster_organization_fkey
    FOREIGN KEY (cluster_id, organization_id)
    REFERENCES execution_clusters(id, organization_id);

CREATE INDEX idx_runners_cluster_id ON runners(cluster_id);

ALTER TABLE pods ADD COLUMN cluster_id BIGINT;

UPDATE pods AS pod
SET cluster_id = runner.cluster_id
FROM runners AS runner
WHERE runner.id = pod.runner_id;

ALTER TABLE pods
  ALTER COLUMN cluster_id SET NOT NULL,
  ADD CONSTRAINT pods_execution_cluster_organization_fkey
    FOREIGN KEY (cluster_id, organization_id)
    REFERENCES execution_clusters(id, organization_id);

CREATE INDEX idx_pods_cluster_id ON pods(cluster_id);

ALTER TABLE runner_grpc_registration_tokens ADD COLUMN cluster_id BIGINT;

UPDATE runner_grpc_registration_tokens AS token
SET cluster_id = cluster.id
FROM execution_clusters AS cluster
WHERE cluster.organization_id = token.organization_id
  AND cluster.slug = 'local';

ALTER TABLE runner_grpc_registration_tokens
  ALTER COLUMN cluster_id SET NOT NULL,
  ADD CONSTRAINT runner_grpc_registration_tokens_cluster_organization_fkey
    FOREIGN KEY (cluster_id, organization_id)
    REFERENCES execution_clusters(id, organization_id);

CREATE INDEX idx_runner_grpc_registration_tokens_cluster_id
  ON runner_grpc_registration_tokens(cluster_id);

ALTER TABLE runner_pending_auths ADD COLUMN cluster_id BIGINT;

UPDATE runner_pending_auths SET authorized = FALSE WHERE authorized IS NULL;
ALTER TABLE runner_pending_auths ALTER COLUMN authorized SET NOT NULL;

UPDATE runner_pending_auths AS pending
SET cluster_id = runner.cluster_id
FROM runners AS runner
WHERE pending.runner_id = runner.id;

UPDATE runner_pending_auths AS pending
SET organization_id = runner.organization_id
FROM runners AS runner
WHERE pending.runner_id = runner.id
  AND pending.organization_id IS NULL;

UPDATE runner_pending_auths AS pending
SET cluster_id = cluster.id
FROM execution_clusters AS cluster
WHERE pending.cluster_id IS NULL
  AND pending.organization_id = cluster.organization_id
  AND cluster.slug = 'local';

ALTER TABLE runner_pending_auths
  ADD CONSTRAINT runner_pending_auths_cluster_organization_fkey
    FOREIGN KEY (cluster_id, organization_id)
    REFERENCES execution_clusters(id, organization_id),
  ADD CONSTRAINT runner_pending_auths_cluster_ownership_check CHECK (
    (organization_id IS NULL AND cluster_id IS NULL AND authorized = FALSE)
    OR (organization_id IS NOT NULL AND cluster_id IS NOT NULL)
  );

CREATE INDEX idx_runner_pending_auths_cluster_id ON runner_pending_auths(cluster_id);
