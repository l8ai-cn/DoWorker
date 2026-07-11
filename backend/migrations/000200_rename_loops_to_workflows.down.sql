ALTER TABLE workflow_runs RENAME COLUMN workflow_id TO loop_id;

ALTER INDEX idx_workflow_runs_workflow_id RENAME TO idx_loop_runs_loop_id;
ALTER INDEX idx_workflow_runs_active RENAME TO idx_loop_runs_active;
ALTER INDEX idx_workflow_runs_workflow_number RENAME TO idx_loop_runs_loop_number;
ALTER INDEX idx_workflow_runs_pod_key RENAME TO idx_loop_runs_pod_key;
ALTER INDEX idx_workflow_runs_autopilot_key RENAME TO idx_loop_runs_autopilot_key;

ALTER INDEX idx_workflows_org_slug RENAME TO idx_loops_org_slug;
ALTER INDEX idx_workflows_cron_due RENAME TO idx_loops_cron_due;
ALTER INDEX idx_workflows_org_status RENAME TO idx_loops_org_status;
ALTER INDEX idx_workflows_agent_slug RENAME TO idx_loops_agent_slug;
ALTER INDEX idx_workflows_model_resource_id RENAME TO idx_loops_model_resource_id;

ALTER TABLE workflows RENAME CONSTRAINT workflows_slug_format TO loops_slug_format;
ALTER TABLE workflows RENAME CONSTRAINT workflows_slug_not_reserved TO loops_slug_not_reserved;
ALTER TABLE workflows RENAME CONSTRAINT workflows_model_resource_id_fkey TO loops_model_resource_id_fkey;

ALTER SEQUENCE workflows_id_seq RENAME TO loops_id_seq;
ALTER SEQUENCE workflow_runs_id_seq RENAME TO loop_runs_id_seq;

ALTER TABLE workflow_runs RENAME TO loop_runs;
ALTER TABLE workflows RENAME TO loops;

COMMENT ON COLUMN loops.used_env_bundles IS
  'Ordered list of EnvBundle names attached to every run of this loop.';
