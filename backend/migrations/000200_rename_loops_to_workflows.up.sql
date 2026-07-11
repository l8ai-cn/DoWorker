ALTER TABLE loops RENAME TO workflows;
ALTER TABLE loop_runs RENAME TO workflow_runs;
ALTER TABLE workflow_runs RENAME COLUMN loop_id TO workflow_id;

ALTER SEQUENCE loops_id_seq RENAME TO workflows_id_seq;
ALTER SEQUENCE loop_runs_id_seq RENAME TO workflow_runs_id_seq;

ALTER TABLE workflows RENAME CONSTRAINT loops_slug_format TO workflows_slug_format;
ALTER TABLE workflows RENAME CONSTRAINT loops_slug_not_reserved TO workflows_slug_not_reserved;
ALTER TABLE workflows RENAME CONSTRAINT loops_model_resource_id_fkey TO workflows_model_resource_id_fkey;

ALTER INDEX idx_loops_org_slug RENAME TO idx_workflows_org_slug;
ALTER INDEX idx_loops_cron_due RENAME TO idx_workflows_cron_due;
ALTER INDEX idx_loops_org_status RENAME TO idx_workflows_org_status;
ALTER INDEX idx_loops_agent_slug RENAME TO idx_workflows_agent_slug;
ALTER INDEX idx_loops_model_resource_id RENAME TO idx_workflows_model_resource_id;

ALTER INDEX idx_loop_runs_loop_id RENAME TO idx_workflow_runs_workflow_id;
ALTER INDEX idx_loop_runs_active RENAME TO idx_workflow_runs_active;
ALTER INDEX idx_loop_runs_loop_number RENAME TO idx_workflow_runs_workflow_number;
ALTER INDEX idx_loop_runs_pod_key RENAME TO idx_workflow_runs_pod_key;
ALTER INDEX idx_loop_runs_autopilot_key RENAME TO idx_workflow_runs_autopilot_key;

COMMENT ON COLUMN workflows.used_env_bundles IS
  'Ordered EnvBundle names attached to every Workflow run.';
