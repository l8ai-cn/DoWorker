ALTER TABLE pods ADD COLUMN model_resource_id BIGINT;
ALTER TABLE pods ADD COLUMN generation BIGINT NOT NULL DEFAULT 0 CONSTRAINT pods_generation_nonnegative CHECK (generation >= 0);
ALTER TABLE pods ADD COLUMN active_config_revision_id BIGINT;
ALTER TABLE pods ADD COLUMN pending_config_revision_id BIGINT;
ALTER TABLE pods ADD COLUMN reinitialize_dispatched_at TIMESTAMPTZ;
ALTER TABLE pods ADD COLUMN archived_at TIMESTAMPTZ;
ALTER TABLE pods ADD COLUMN archived_by_id BIGINT;
ALTER TABLE pods ADD COLUMN purge_after TIMESTAMPTZ;

ALTER TABLE pods
  ADD CONSTRAINT pods_model_resource_id_fkey
  FOREIGN KEY (model_resource_id) REFERENCES model_resources(id) ON DELETE SET NULL;

ALTER TABLE pods
  ADD CONSTRAINT pods_archived_by_id_fkey
  FOREIGN KEY (archived_by_id) REFERENCES users(id) ON DELETE SET NULL;

CREATE TABLE pod_config_revisions (
  id BIGSERIAL PRIMARY KEY,
  pod_id BIGINT NOT NULL,
  revision BIGINT NOT NULL,
  agentfile_layer TEXT NOT NULL DEFAULT '',
  status VARCHAR(20) NOT NULL DEFAULT 'draft',
  config_summary JSONB NOT NULL DEFAULT '{}'::jsonb,
  model_resource_id BIGINT,
  created_by_id BIGINT NOT NULL,
  error_message TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  applied_at TIMESTAMPTZ,
  CONSTRAINT pod_config_revisions_pod_revision_key UNIQUE (pod_id, revision),
  CONSTRAINT pod_config_revisions_revision_positive CHECK (revision > 0),
  CONSTRAINT pod_config_revisions_status_check CHECK (status IN ('draft', 'applying', 'active', 'failed')),
  CONSTRAINT pod_config_revisions_config_summary_object CHECK (jsonb_typeof(config_summary) = 'object'),
  CONSTRAINT pod_config_revisions_pod_id_fkey FOREIGN KEY (pod_id) REFERENCES pods(id) ON DELETE CASCADE,
  CONSTRAINT pod_config_revisions_model_resource_id_fkey FOREIGN KEY (model_resource_id) REFERENCES model_resources(id) ON DELETE SET NULL,
  CONSTRAINT pod_config_revisions_created_by_id_fkey FOREIGN KEY (created_by_id) REFERENCES users(id) ON DELETE RESTRICT
);

CREATE INDEX idx_pods_model_resource_id ON pods (model_resource_id);
CREATE INDEX idx_pods_active_config_revision_id ON pods (active_config_revision_id);
CREATE INDEX idx_pods_pending_config_revision_id ON pods (pending_config_revision_id);
CREATE INDEX idx_pod_config_revisions_pod_id ON pod_config_revisions (pod_id);
CREATE INDEX idx_pod_config_revisions_status ON pod_config_revisions (status);
CREATE INDEX idx_pod_config_revisions_model_resource_id ON pod_config_revisions (model_resource_id);
CREATE UNIQUE INDEX idx_pod_config_revisions_one_active_per_pod ON pod_config_revisions (pod_id) WHERE status = 'active';

ALTER TABLE pods
  ADD CONSTRAINT pods_active_config_revision_id_fkey
  FOREIGN KEY (active_config_revision_id) REFERENCES pod_config_revisions(id) ON DELETE SET NULL;

ALTER TABLE pods
  ADD CONSTRAINT pods_pending_config_revision_id_fkey
  FOREIGN KEY (pending_config_revision_id) REFERENCES pod_config_revisions(id) ON DELETE SET NULL;
