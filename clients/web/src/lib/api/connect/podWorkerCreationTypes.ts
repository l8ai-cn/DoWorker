export interface WorkerSecretReference {
  field: string;
  kind: string;
  id: number;
}

export interface WorkerKnowledgeMount {
  knowledge_base_id: number;
  mode: string;
}

export interface WorkerSpecDraft {
  model_resource_id: number;
  worker_type_slug: string;
  runtime_image_id: number;
  placement_policy: string;
  compute_target_id: number;
  deployment_mode: string;
  resource_profile_id: number;
  type_schema_version: number;
  type_config_values: Record<string, unknown>;
  secret_refs: WorkerSecretReference[];
  interaction_mode: string;
  automation_level: string;
  repository_id?: number;
  branch: string;
  skill_ids: number[];
  knowledge_mounts: WorkerKnowledgeMount[];
  env_bundle_ids: number[];
  instructions: string;
  initial_task: string;
  termination_policy: string;
  idle_timeout_minutes: number;
  alias: string;
  source_expert_id?: number;
  options_revision: string;
}

export interface WorkerTypeOption {
  slug: string;
  name: string;
  description: string;
  schema_version: number;
  config_schema: Record<string, unknown>;
  selectable: boolean;
  blocking_reason: string;
}

export interface WorkerRuntimeImageOption {
  id: number;
  slug: string;
  name: string;
  reference: string;
  digest: string;
  worker_type_slugs: string[];
  selectable: boolean;
  blocking_reason: string;
}

export interface WorkerComputeTargetOption {
  id: number;
  slug: string;
  name: string;
  kind: string;
  supports_pooled: boolean;
  supports_dedicated: boolean;
  selectable: boolean;
  blocking_reason: string;
}

export interface WorkerDeploymentModeOption {
  value: string;
  name: string;
  selectable: boolean;
  blocking_reason: string;
}

export interface WorkerResourceProfileOption {
  id: number;
  slug: string;
  name: string;
  cpu_request_millicpu: number;
  cpu_limit_millicpu: number;
  memory_request_bytes: number;
  memory_limit_bytes: number;
  gpu_request?: number;
  gpu_limit?: number;
  selectable: boolean;
  blocking_reason: string;
}

export interface WorkerCreateOptions {
  revision: string;
  worker_types: WorkerTypeOption[];
  runtime_images: WorkerRuntimeImageOption[];
  compute_targets: WorkerComputeTargetOption[];
  deployment_modes: WorkerDeploymentModeOption[];
  resource_profiles: WorkerResourceProfileOption[];
}

export interface WorkerCreateOptionsFilter {
  worker_type_slug?: string;
  compute_target_id?: number;
  deployment_mode?: string;
}

export interface WorkerPreflightIssue {
  code: string;
  field: string;
  message: string;
  severity: string;
}

export interface WorkerPreflightResult {
  issues: WorkerPreflightIssue[];
  resolved_spec_json?: string;
  options_revision: string;
}

export interface WorkerDraftFillResult {
  draft: WorkerSpecDraft;
  issues: WorkerPreflightIssue[];
}
