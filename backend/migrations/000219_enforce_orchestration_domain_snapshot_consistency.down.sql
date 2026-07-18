BEGIN;

ALTER TABLE experts
  DROP CONSTRAINT IF EXISTS experts_orchestration_revision_fkey,
  ADD CONSTRAINT experts_orchestration_revision_fkey
  FOREIGN KEY (organization_id, orchestration_resource_id, orchestration_resource_revision)
  REFERENCES orchestration_resource_revisions (organization_id, resource_id, revision)
  ON DELETE RESTRICT
  DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE workflows
  DROP CONSTRAINT IF EXISTS workflows_orchestration_revision_fkey,
  ADD CONSTRAINT workflows_orchestration_revision_fkey
  FOREIGN KEY (organization_id, orchestration_resource_id, orchestration_resource_revision)
  REFERENCES orchestration_resource_revisions (organization_id, resource_id, revision)
  ON DELETE RESTRICT
  DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE workflow_runs
  DROP CONSTRAINT IF EXISTS workflow_runs_orchestration_revision_fkey,
  ADD CONSTRAINT workflow_runs_orchestration_revision_fkey
  FOREIGN KEY (organization_id, orchestration_resource_id, orchestration_resource_revision)
  REFERENCES orchestration_resource_revisions (organization_id, resource_id, revision)
  ON DELETE RESTRICT
  DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE goal_loops
  DROP CONSTRAINT IF EXISTS goal_loops_orchestration_revision_fkey,
  ADD CONSTRAINT goal_loops_orchestration_revision_fkey
  FOREIGN KEY (organization_id, orchestration_resource_id, orchestration_resource_revision)
  REFERENCES orchestration_resource_revisions (organization_id, resource_id, revision)
  ON DELETE RESTRICT
  DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE orchestration_worker_launches
  DROP CONSTRAINT IF EXISTS orchestration_worker_launches_revision_fkey,
  ADD CONSTRAINT orchestration_worker_launches_revision_fkey
  FOREIGN KEY (organization_id, resource_id, resource_revision)
  REFERENCES orchestration_resource_revisions (organization_id, resource_id, revision)
  ON DELETE RESTRICT
  DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE orchestration_resource_revisions
  DROP CONSTRAINT IF EXISTS orchestration_resource_revisions_org_revision_snapshot_unique;

COMMIT;
