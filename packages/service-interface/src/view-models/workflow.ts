// Workflow UI ViewModels — snake_case shapes for workflow form / store / kanban.
// Kept snake_case because Rust SSOT (state/workflow_state.rs) serializes
// WorkflowData with serde snake_case JSON for set_workflows round-trip. Owned here
// (zero-dep contract layer) so Web projections share one type definition.

export type WorkflowStatus = "enabled" | "disabled" | "archived";
export type ExecutionMode = "autopilot" | "direct";
export type SandboxStrategy = "persistent" | "fresh";
export type ConcurrencyPolicy = "skip" | "queue" | "replace";
export type RunStatus = "pending" | "running" | "completed" | "failed" | "timeout" | "cancelled" | "skipped";

export interface WorkflowData {
  id: number;
  organization_id: number;
  name: string;
  slug: string;
  description?: string;
  agent_slug?: string;
  custom_agent_slug?: string;
  permission_mode: string;
  prompt_template: string;
  prompt_variables?: Record<string, unknown>;
  repository_id?: number;
  runner_id?: number;
  branch_name?: string;
  ticket_id?: number;
  model_resource_id?: number;
  /**
   * Ordered list of EnvBundle names attached to every run. Each name is
   * emitted as a `USE_ENV_BUNDLE "<name>"` line in the generated AgentFile
   * (in array order; later entries override earlier ones on conflicting
   * env keys). Empty/absent = no bundle.
   */
  used_env_bundles: string[];
  config_overrides?: Record<string, unknown>;
  execution_mode: ExecutionMode;
  cron_expression?: string;
  callback_url?: string;
  autopilot_config: Record<string, unknown>;
  status: WorkflowStatus;
  sandbox_strategy: SandboxStrategy;
  session_persistence: boolean;
  concurrency_policy: ConcurrencyPolicy;
  max_concurrent_runs: number;
  max_retained_runs: number;
  timeout_minutes: number;
  sandbox_path?: string;
  last_pod_key?: string;
  created_by_id: number;
  total_runs: number;
  successful_runs: number;
  failed_runs: number;
  active_run_count: number;
  avg_duration_sec?: number;
  last_run_at?: string;
  next_run_at?: string;
  created_at: string;
  updated_at: string;
}

export interface WorkflowRunData {
  id: number;
  organization_id: number;
  workflow_id: number;
  run_number: number;
  status: RunStatus;
  pod_key?: string;
  autopilot_controller_key?: string;
  trigger_type: string;
  trigger_source?: string;
  resolved_prompt?: string;
  started_at?: string;
  finished_at?: string;
  duration_sec?: number;
  exit_summary?: string;
  error_message?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateWorkflowRequest {
  name: string;
  slug?: string;
  description?: string;
  agent_slug?: string;
  custom_agent_slug?: string;
  permission_mode?: string;
  prompt_template: string;
  prompt_variables?: Record<string, unknown>;
  repository_id?: number;
  runner_id?: number;
  branch_name?: string;
  ticket_id?: number;
  model_resource_id?: number;
  used_env_bundles?: string[];
  config_overrides?: Record<string, unknown>;
  execution_mode?: string;
  cron_expression?: string;
  autopilot_config?: Record<string, unknown>;
  callback_url?: string;
  sandbox_strategy?: string;
  session_persistence?: boolean;
  concurrency_policy?: string;
  max_concurrent_runs?: number;
  max_retained_runs?: number;
  timeout_minutes?: number;
}

export interface UpdateWorkflowRequest {
  name?: string;
  description?: string;
  agent_slug?: string;
  custom_agent_slug?: string;
  permission_mode?: string;
  prompt_template?: string;
  prompt_variables?: Record<string, unknown>;
  repository_id?: number;
  runner_id?: number;
  branch_name?: string;
  ticket_id?: number;
  model_resource_id?: number;
  used_env_bundles?: string[];
  config_overrides?: Record<string, unknown>;
  execution_mode?: string;
  cron_expression?: string;
  autopilot_config?: Record<string, unknown>;
  callback_url?: string;
  sandbox_strategy?: string;
  session_persistence?: boolean;
  concurrency_policy?: string;
  max_concurrent_runs?: number;
  max_retained_runs?: number;
  timeout_minutes?: number;
}
