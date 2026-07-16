ALTER TABLE orchestration_resource_revisions
  ADD CONSTRAINT orchestration_resource_revisions_org_revision_unique
  UNIQUE (organization_id, resource_id, revision);

ALTER TABLE experts
  ADD COLUMN orchestration_resource_id BIGINT,
  ADD COLUMN orchestration_resource_revision BIGINT;

ALTER TABLE experts
  ADD CONSTRAINT experts_orchestration_mode_check CHECK (
    (orchestration_resource_id IS NULL
      AND orchestration_resource_revision IS NULL)
    OR
    (orchestration_resource_id IS NOT NULL
      AND orchestration_resource_revision > 0
      AND worker_spec_snapshot_id IS NOT NULL)
  ),
  ADD CONSTRAINT experts_orchestration_revision_fkey
  FOREIGN KEY (
    organization_id,
    orchestration_resource_id,
    orchestration_resource_revision
  )
  REFERENCES orchestration_resource_revisions (
    organization_id,
    resource_id,
    revision
  )
  ON DELETE RESTRICT
  DEFERRABLE INITIALLY DEFERRED;

CREATE UNIQUE INDEX idx_experts_orchestration_resource
  ON experts (organization_id, orchestration_resource_id)
  WHERE orchestration_resource_id IS NOT NULL;

ALTER TABLE workflows
  ADD COLUMN orchestration_resource_id BIGINT,
  ADD COLUMN orchestration_resource_revision BIGINT,
  ADD COLUMN worker_spec_snapshot_id BIGINT;

ALTER TABLE workflows
  ADD CONSTRAINT workflows_orchestration_mode_check CHECK (
    (orchestration_resource_id IS NULL
      AND orchestration_resource_revision IS NULL
      AND worker_spec_snapshot_id IS NULL)
    OR
    (orchestration_resource_id IS NOT NULL
      AND orchestration_resource_revision > 0
      AND worker_spec_snapshot_id IS NOT NULL)
  ),
  ADD CONSTRAINT workflows_worker_spec_snapshot_org_fkey
  FOREIGN KEY (organization_id, worker_spec_snapshot_id)
  REFERENCES worker_spec_snapshots (organization_id, id)
  ON DELETE RESTRICT,
  ADD CONSTRAINT workflows_orchestration_revision_fkey
  FOREIGN KEY (
    organization_id,
    orchestration_resource_id,
    orchestration_resource_revision
  )
  REFERENCES orchestration_resource_revisions (
    organization_id,
    resource_id,
    revision
  )
  ON DELETE RESTRICT
  DEFERRABLE INITIALLY DEFERRED;

CREATE UNIQUE INDEX idx_workflows_orchestration_resource
  ON workflows (organization_id, orchestration_resource_id)
  WHERE orchestration_resource_id IS NOT NULL;

ALTER TABLE workflow_runs
  ADD COLUMN orchestration_resource_id BIGINT,
  ADD COLUMN orchestration_resource_revision BIGINT,
  ADD COLUMN worker_spec_snapshot_id BIGINT;

ALTER TABLE workflow_runs
  ADD CONSTRAINT workflow_runs_orchestration_mode_check CHECK (
    (orchestration_resource_id IS NULL
      AND orchestration_resource_revision IS NULL
      AND worker_spec_snapshot_id IS NULL)
    OR
    (orchestration_resource_id IS NOT NULL
      AND orchestration_resource_revision > 0
      AND worker_spec_snapshot_id IS NOT NULL)
  ),
  ADD CONSTRAINT workflow_runs_worker_spec_snapshot_org_fkey
  FOREIGN KEY (organization_id, worker_spec_snapshot_id)
  REFERENCES worker_spec_snapshots (organization_id, id)
  ON DELETE RESTRICT,
  ADD CONSTRAINT workflow_runs_orchestration_revision_fkey
  FOREIGN KEY (
    organization_id,
    orchestration_resource_id,
    orchestration_resource_revision
  )
  REFERENCES orchestration_resource_revisions (
    organization_id,
    resource_id,
    revision
  )
  ON DELETE RESTRICT
  DEFERRABLE INITIALLY DEFERRED;
