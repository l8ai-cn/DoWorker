ALTER TABLE workflow_runs
  DROP CONSTRAINT IF EXISTS workflow_runs_orchestration_revision_fkey,
  DROP CONSTRAINT IF EXISTS workflow_runs_worker_spec_snapshot_org_fkey,
  DROP CONSTRAINT IF EXISTS workflow_runs_orchestration_mode_check,
  DROP COLUMN IF EXISTS orchestration_resource_revision,
  DROP COLUMN IF EXISTS orchestration_resource_id,
  DROP COLUMN IF EXISTS worker_spec_snapshot_id;

DROP INDEX IF EXISTS idx_workflows_orchestration_resource;

ALTER TABLE workflows
  DROP CONSTRAINT IF EXISTS workflows_orchestration_revision_fkey,
  DROP CONSTRAINT IF EXISTS workflows_worker_spec_snapshot_org_fkey,
  DROP CONSTRAINT IF EXISTS workflows_orchestration_mode_check,
  DROP COLUMN IF EXISTS orchestration_resource_revision,
  DROP COLUMN IF EXISTS orchestration_resource_id,
  DROP COLUMN IF EXISTS worker_spec_snapshot_id;

DROP INDEX IF EXISTS idx_experts_orchestration_resource;

ALTER TABLE experts
  DROP CONSTRAINT IF EXISTS experts_orchestration_revision_fkey,
  DROP CONSTRAINT IF EXISTS experts_orchestration_mode_check,
  DROP COLUMN IF EXISTS orchestration_resource_revision,
  DROP COLUMN IF EXISTS orchestration_resource_id;

ALTER TABLE orchestration_resource_revisions
  DROP CONSTRAINT IF EXISTS orchestration_resource_revisions_org_revision_unique;
