ALTER TABLE orchestration_resource_plans
  ADD CONSTRAINT orchestration_resource_plans_org_id_unique
  UNIQUE (organization_id, id);

ALTER TABLE pods
  ADD CONSTRAINT pods_org_id_key_unique
  UNIQUE (organization_id, id, pod_key);

CREATE TABLE orchestration_worker_launches (
  id BIGSERIAL PRIMARY KEY,
  organization_id BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  actor_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  plan_id UUID NOT NULL,
  resource_id BIGINT NOT NULL,
  resource_revision BIGINT NOT NULL,
  worker_spec_snapshot_id BIGINT NOT NULL,
  prompt TEXT,
  alias VARCHAR(100) NOT NULL DEFAULT '',
  state VARCHAR(20) NOT NULL DEFAULT 'pending',
  claim_token UUID,
  lease_expires_at TIMESTAMPTZ,
  attempt_count INTEGER NOT NULL DEFAULT 0,
  last_error TEXT,
  pod_id BIGINT,
  pod_key VARCHAR(100),
  runner_id BIGINT,
  dispatched_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT orchestration_worker_launches_org_id_unique
    UNIQUE (organization_id, id),
  CONSTRAINT orchestration_worker_launches_plan_unique
    UNIQUE (organization_id, plan_id),
  CONSTRAINT orchestration_worker_launches_resource_unique
    UNIQUE (organization_id, resource_id),
  CONSTRAINT orchestration_worker_launches_plan_fkey
    FOREIGN KEY (organization_id, plan_id)
    REFERENCES orchestration_resource_plans (organization_id, id)
    ON DELETE RESTRICT,
  CONSTRAINT orchestration_worker_launches_revision_fkey
    FOREIGN KEY (organization_id, resource_id, resource_revision)
    REFERENCES orchestration_resource_revisions (
      organization_id,
      resource_id,
      revision
    )
    ON DELETE RESTRICT
    DEFERRABLE INITIALLY DEFERRED,
  CONSTRAINT orchestration_worker_launches_snapshot_fkey
    FOREIGN KEY (organization_id, worker_spec_snapshot_id)
    REFERENCES worker_spec_snapshots (organization_id, id)
    ON DELETE RESTRICT,
  CONSTRAINT orchestration_worker_launches_state_check CHECK (
    state IN ('pending', 'materializing', 'dispatched')
  ),
  CONSTRAINT orchestration_worker_launches_state_fields_check CHECK ((
    (
      state = 'pending'
      AND claim_token IS NULL
      AND lease_expires_at IS NULL
      AND pod_id IS NULL
      AND pod_key IS NULL
      AND runner_id IS NULL
      AND dispatched_at IS NULL
    )
    OR (
      state = 'materializing'
      AND claim_token IS NOT NULL
      AND lease_expires_at IS NOT NULL
      AND pod_id IS NULL
      AND pod_key IS NULL
      AND runner_id IS NULL
      AND dispatched_at IS NULL
    )
    OR (
      state = 'dispatched'
      AND claim_token IS NULL
      AND lease_expires_at IS NULL
      AND pod_id IS NOT NULL
      AND pod_key IS NOT NULL
      AND runner_id IS NOT NULL
      AND dispatched_at IS NOT NULL
    )
  ) IS TRUE),
  CONSTRAINT orchestration_worker_launches_attempt_check CHECK (
    attempt_count >= 0
  ),
  CONSTRAINT orchestration_worker_launches_timestamps_check CHECK (
    isfinite(created_at)
    AND isfinite(updated_at)
    AND updated_at >= created_at
    AND (lease_expires_at IS NULL OR isfinite(lease_expires_at))
    AND (dispatched_at IS NULL OR isfinite(dispatched_at))
  )
);

CREATE INDEX idx_orchestration_worker_launches_pending
  ON orchestration_worker_launches (
    organization_id,
    state,
    lease_expires_at,
    id
  )
  WHERE state <> 'dispatched';

ALTER TABLE pods
  ADD COLUMN orchestration_worker_launch_id BIGINT,
  ADD CONSTRAINT pods_orchestration_worker_launch_fkey
  FOREIGN KEY (organization_id, orchestration_worker_launch_id)
  REFERENCES orchestration_worker_launches (organization_id, id)
  ON DELETE RESTRICT
  DEFERRABLE INITIALLY DEFERRED;

CREATE UNIQUE INDEX idx_pods_orchestration_worker_launch
  ON pods (organization_id, orchestration_worker_launch_id)
  WHERE orchestration_worker_launch_id IS NOT NULL;

ALTER TABLE orchestration_worker_launches
  ADD CONSTRAINT orchestration_worker_launches_pod_fkey
  FOREIGN KEY (organization_id, pod_id, pod_key)
  REFERENCES pods (organization_id, id, pod_key)
  ON DELETE RESTRICT
  DEFERRABLE INITIALLY DEFERRED;
