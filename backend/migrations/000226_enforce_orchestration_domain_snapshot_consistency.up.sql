BEGIN;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM experts domain_row
    WHERE domain_row.orchestration_resource_id IS NOT NULL
      AND NOT EXISTS (
        SELECT 1
        FROM orchestration_resource_revisions revision
        WHERE revision.organization_id = domain_row.organization_id
          AND revision.resource_id = domain_row.orchestration_resource_id
          AND revision.revision = domain_row.orchestration_resource_revision
          AND revision.worker_spec_snapshot_id = domain_row.worker_spec_snapshot_id
      )
  ) THEN
    RAISE EXCEPTION 'experts contain orchestration revision/snapshot mismatches';
  END IF;

  IF EXISTS (
    SELECT 1
    FROM workflows domain_row
    WHERE domain_row.orchestration_resource_id IS NOT NULL
      AND NOT EXISTS (
        SELECT 1
        FROM orchestration_resource_revisions revision
        WHERE revision.organization_id = domain_row.organization_id
          AND revision.resource_id = domain_row.orchestration_resource_id
          AND revision.revision = domain_row.orchestration_resource_revision
          AND revision.worker_spec_snapshot_id = domain_row.worker_spec_snapshot_id
      )
  ) THEN
    RAISE EXCEPTION 'workflows contain orchestration revision/snapshot mismatches';
  END IF;

  IF EXISTS (
    SELECT 1
    FROM workflow_runs domain_row
    WHERE domain_row.orchestration_resource_id IS NOT NULL
      AND NOT EXISTS (
        SELECT 1
        FROM orchestration_resource_revisions revision
        WHERE revision.organization_id = domain_row.organization_id
          AND revision.resource_id = domain_row.orchestration_resource_id
          AND revision.revision = domain_row.orchestration_resource_revision
          AND revision.worker_spec_snapshot_id = domain_row.worker_spec_snapshot_id
      )
  ) THEN
    RAISE EXCEPTION 'workflow_runs contain orchestration revision/snapshot mismatches';
  END IF;

  IF EXISTS (
    SELECT 1
    FROM goal_loops domain_row
    WHERE domain_row.orchestration_resource_id IS NOT NULL
      AND NOT EXISTS (
        SELECT 1
        FROM orchestration_resource_revisions revision
        WHERE revision.organization_id = domain_row.organization_id
          AND revision.resource_id = domain_row.orchestration_resource_id
          AND revision.revision = domain_row.orchestration_resource_revision
          AND revision.worker_spec_snapshot_id = domain_row.worker_spec_snapshot_id
      )
  ) THEN
    RAISE EXCEPTION 'goal_loops contain orchestration revision/snapshot mismatches';
  END IF;

  IF EXISTS (
    SELECT 1
    FROM orchestration_worker_launches domain_row
    WHERE NOT EXISTS (
      SELECT 1
      FROM orchestration_resource_revisions revision
      WHERE revision.organization_id = domain_row.organization_id
        AND revision.resource_id = domain_row.resource_id
        AND revision.revision = domain_row.resource_revision
        AND revision.worker_spec_snapshot_id = domain_row.worker_spec_snapshot_id
    )
  ) THEN
    RAISE EXCEPTION 'orchestration_worker_launches contain revision/snapshot mismatches';
  END IF;
END
$$;

ALTER TABLE orchestration_resource_revisions
  ADD CONSTRAINT orchestration_resource_revisions_org_revision_snapshot_unique
  UNIQUE (organization_id, resource_id, revision, worker_spec_snapshot_id);

ALTER TABLE experts
  DROP CONSTRAINT experts_orchestration_revision_fkey,
  ADD CONSTRAINT experts_orchestration_revision_fkey
  FOREIGN KEY (
    organization_id,
    orchestration_resource_id,
    orchestration_resource_revision,
    worker_spec_snapshot_id
  )
  REFERENCES orchestration_resource_revisions (
    organization_id,
    resource_id,
    revision,
    worker_spec_snapshot_id
  )
  ON DELETE RESTRICT
  DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE workflows
  DROP CONSTRAINT workflows_orchestration_revision_fkey,
  ADD CONSTRAINT workflows_orchestration_revision_fkey
  FOREIGN KEY (
    organization_id,
    orchestration_resource_id,
    orchestration_resource_revision,
    worker_spec_snapshot_id
  )
  REFERENCES orchestration_resource_revisions (
    organization_id,
    resource_id,
    revision,
    worker_spec_snapshot_id
  )
  ON DELETE RESTRICT
  DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE workflow_runs
  DROP CONSTRAINT workflow_runs_orchestration_revision_fkey,
  ADD CONSTRAINT workflow_runs_orchestration_revision_fkey
  FOREIGN KEY (
    organization_id,
    orchestration_resource_id,
    orchestration_resource_revision,
    worker_spec_snapshot_id
  )
  REFERENCES orchestration_resource_revisions (
    organization_id,
    resource_id,
    revision,
    worker_spec_snapshot_id
  )
  ON DELETE RESTRICT
  DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE goal_loops
  DROP CONSTRAINT goal_loops_orchestration_revision_fkey,
  ADD CONSTRAINT goal_loops_orchestration_revision_fkey
  FOREIGN KEY (
    organization_id,
    orchestration_resource_id,
    orchestration_resource_revision,
    worker_spec_snapshot_id
  )
  REFERENCES orchestration_resource_revisions (
    organization_id,
    resource_id,
    revision,
    worker_spec_snapshot_id
  )
  ON DELETE RESTRICT
  DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE orchestration_worker_launches
  DROP CONSTRAINT orchestration_worker_launches_revision_fkey,
  ADD CONSTRAINT orchestration_worker_launches_revision_fkey
  FOREIGN KEY (
    organization_id,
    resource_id,
    resource_revision,
    worker_spec_snapshot_id
  )
  REFERENCES orchestration_resource_revisions (
    organization_id,
    resource_id,
    revision,
    worker_spec_snapshot_id
  )
  ON DELETE RESTRICT
  DEFERRABLE INITIALLY DEFERRED;

COMMIT;
