export interface WorkerSecretReference {
  field: string;
  kind: string;
  id: number;
}

export interface WorkerKnowledgeMount {
  knowledge_base_id: number;
  mode: string;
}

export interface WorkerConfigDocumentBinding {
  document_id: string;
  config_bundle_id: number;
}

export interface WorkerCredentialRequirement {
  id: string;
  source_kind: string;
  source_ref: string;
  target_kind: string;
  target_name: string;
}

export interface WorkerConfigDocumentRequirement {
  document_id: string;
  format: string;
  target_path: string;
  required: boolean;
}

export interface WorkerResourceRequest {
  cpu_request_millicpu: number;
  cpu_limit_millicpu: number;
  memory_request_bytes: number;
  memory_limit_bytes: number;
  storage_request_bytes: number;
  storage_limit_bytes: number;
}

export interface WorkerSpecDraft {
  model_resource_id: number;
  tool_model_resource_ids: Record<string, number>;
  worker_type_slug: string;
  runtime_image_id: number;
  placement_policy: string;
  compute_target_id: number;
  deployment_mode: string;
  resource_profile_id: number;
  custom_resources?: WorkerResourceRequest;
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
  config_document_bindings: WorkerConfigDocumentBinding[];
  instructions: string;
  initial_task: string;
  termination_policy: string;
  idle_timeout_minutes: number;
  alias: string;
  source_expert_id?: number;
  options_revision: string;
}

export interface WorkerToolModelRequirement {
  role: string;
  provider_keys: string[];
  protocol_adapters: string[];
  modality: string;
  capability: string;
}

export interface WorkerTypeOption {
  slug: string;
  name: string;
  description: string;
  schema_version: number;
  config_schema: Record<string, unknown>;
  supported_interaction_modes: string[];
  requires_model_resource: boolean;
  model_protocol_adapters: string[];
  tool_model_requirements: WorkerToolModelRequirement[];
  credential_requirements: WorkerCredentialRequirement[];
  config_document_requirements: WorkerConfigDocumentRequirement[];
  selectable: boolean;
  blocking_reason: string;
  requires_model_resource: boolean;
  model_protocol_adapters: string[];
  tool_model_requirements: WorkerToolModelRequirement[];
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
  storage_request_bytes: number;
  storage_limit_bytes: number;
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
