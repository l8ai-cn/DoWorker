// Connect-RPC adapter for proto.workflow.v1.WorkflowService.
//
// Encodes requests via @bufbuild/protobuf .toBinary(), passes the Uint8Array
// to the wasm bridge (binary in / binary out — conventions §2.5), decodes
// responses via .fromBinary(). No JSON intermediate.
//
// Returns the existing WorkflowData / WorkflowRunData shapes (viewModels/workflow.ts)
// so call-sites in the workflow store don't need to flip off camelCase + BigInt.

import {
  CancelWorkflowRunRequestSchema,
  CancelWorkflowRunResponseSchema,
  CreateWorkflowRequestSchema,
  DeleteWorkflowRequestSchema,
  DeleteWorkflowResponseSchema,
  EnvBundleListSchema,
  GetWorkflowRequestSchema,
  ListWorkflowsRequestSchema,
  ListWorkflowsResponseSchema,
  ListWorkflowRunsRequestSchema,
  ListWorkflowRunsResponseSchema,
  WorkflowActionRequestSchema,
  WorkflowSchema,
  TriggerWorkflowRequestSchema,
  TriggerWorkflowResponseSchema,
  UpdateWorkflowRequestSchema,
} from "@proto/workflow/v1/workflow_pb";
import { create, toBinary, fromBinary } from "@bufbuild/protobuf";
// Shared proto->WorkflowData projection. Aliased to the historical fromProto*
// names used below.
import { workflowToCache as fromProtoWorkflow, workflowRunToCache as fromProtoWorkflowRun } from "@/lib/api/projections";
import { getWorkflowService } from "@/lib/wasm-core";
import type {
  CreateWorkflowRequest,
  WorkflowData,
  WorkflowRunData,
  UpdateWorkflowRequest,
} from "@/lib/viewModels/workflow";

interface ListFilters {
  status?: string;
  executionMode?: string;
  cronEnabled?: boolean;
  query?: string;
  limit?: number;
  offset?: number;
}

export async function listWorkflows(
  orgSlug: string,
  filters?: ListFilters,
): Promise<{ items: WorkflowData[]; total: number }> {
  const req = create(ListWorkflowsRequestSchema, {
    orgSlug,
    status: filters?.status ?? "",
    executionMode: filters?.executionMode ?? "",
    cronEnabled: filters?.cronEnabled,
    query: filters?.query ?? "",
    offset: filters?.offset,
    limit: filters?.limit,
  });
  const bytes = toBinary(ListWorkflowsRequestSchema, req);
  const respBytes = await getWorkflowService().listWorkflowsConnect(bytes);
  const resp = fromBinary(ListWorkflowsResponseSchema, new Uint8Array(respBytes));
  return { items: resp.items.map(fromProtoWorkflow), total: Number(resp.total) };
}

// Raw wire bytes for the fetch→state path: response → apply_fetched_workflows
// (Rust set_workflows via workflow_from_proto), no TS fromProtoWorkflow/workflowToProtoWorkflow.
export async function listWorkflowsRaw(orgSlug: string, filters?: ListFilters): Promise<Uint8Array> {
  const req = create(ListWorkflowsRequestSchema, {
    orgSlug,
    status: filters?.status ?? "",
    executionMode: filters?.executionMode ?? "",
    cronEnabled: filters?.cronEnabled,
    query: filters?.query ?? "",
    offset: filters?.offset,
    limit: filters?.limit,
  });
  return new Uint8Array(await getWorkflowService().listWorkflowsConnect(toBinary(ListWorkflowsRequestSchema, req)));
}

export async function getWorkflow(orgSlug: string, workflowSlug: string): Promise<WorkflowData> {
  const req = create(GetWorkflowRequestSchema, { orgSlug, workflowSlug });
  const bytes = toBinary(GetWorkflowRequestSchema, req);
  const respBytes = await getWorkflowService().getWorkflowConnect(bytes);
  return fromProtoWorkflow(fromBinary(WorkflowSchema, new Uint8Array(respBytes)));
}

// Raw wire bytes for the fetch→state path: response (Workflow) →
// apply_fetched_current_workflow (Rust set_current_workflow). The full wire Workflow carries
// the proto-only fields the lossy workflowToProtoWorkflow round-trip dropped.
export async function getWorkflowRaw(orgSlug: string, workflowSlug: string): Promise<Uint8Array> {
  const req = create(GetWorkflowRequestSchema, { orgSlug, workflowSlug });
  return new Uint8Array(
    await getWorkflowService().getWorkflowConnect(toBinary(GetWorkflowRequestSchema, req)),
  );
}

function toJsonString(v: unknown): string {
  if (v === undefined || v === null) return "";
  if (typeof v === "string") return v;
  return JSON.stringify(v);
}

export async function createWorkflow(orgSlug: string, data: CreateWorkflowRequest): Promise<WorkflowData> {
  const req = create(CreateWorkflowRequestSchema, {
    orgSlug,
    name: data.name,
    slug: data.slug ?? "",
    description: data.description ?? "",
    agentSlug: data.agent_slug ?? "",
    permissionMode: data.permission_mode ?? "",
    promptTemplate: data.prompt_template,
    promptVariablesJson: toJsonString(data.prompt_variables),
    configOverridesJson: toJsonString(data.config_overrides),
    autopilotConfigJson: toJsonString(data.autopilot_config),
    repositoryId: data.repository_id != null ? BigInt(data.repository_id) : undefined,
    runnerId: data.runner_id != null ? BigInt(data.runner_id) : undefined,
    branchName: data.branch_name ?? "",
    ticketId: data.ticket_id != null ? BigInt(data.ticket_id) : undefined,
    modelResourceId: data.model_resource_id != null ? BigInt(data.model_resource_id) : undefined,
    executionMode: data.execution_mode ?? "",
    cronExpression: data.cron_expression ?? "",
    callbackUrl: data.callback_url ?? "",
    sandboxStrategy: data.sandbox_strategy ?? "",
    sessionPersistence: data.session_persistence,
    concurrencyPolicy: data.concurrency_policy ?? "",
    maxConcurrentRuns: data.max_concurrent_runs,
    maxRetainedRuns: data.max_retained_runs,
    timeoutMinutes: data.timeout_minutes,
    usedEnvBundles: data.used_env_bundles ?? [],
  });
  const bytes = toBinary(CreateWorkflowRequestSchema, req);
  const respBytes = await getWorkflowService().createWorkflowConnect(bytes);
  return fromProtoWorkflow(fromBinary(WorkflowSchema, new Uint8Array(respBytes)));
}

export async function updateWorkflow(
  orgSlug: string,
  workflowSlug: string,
  data: UpdateWorkflowRequest,
): Promise<WorkflowData> {
  const req = create(UpdateWorkflowRequestSchema, {
    orgSlug,
    workflowSlug,
    name: data.name,
    description: data.description,
    agentSlug: data.agent_slug ?? "",
    permissionMode: data.permission_mode,
    promptTemplate: data.prompt_template,
    promptVariablesJson: toJsonString(data.prompt_variables),
    configOverridesJson: toJsonString(data.config_overrides),
    autopilotConfigJson: toJsonString(data.autopilot_config),
    repositoryId: data.repository_id != null ? BigInt(data.repository_id) : undefined,
    runnerId: data.runner_id != null ? BigInt(data.runner_id) : undefined,
    branchName: data.branch_name,
    ticketId: data.ticket_id != null ? BigInt(data.ticket_id) : undefined,
    modelResourceId: data.model_resource_id != null ? BigInt(data.model_resource_id) : undefined,
    executionMode: data.execution_mode,
    cronExpression: data.cron_expression,
    callbackUrl: data.callback_url,
    sandboxStrategy: data.sandbox_strategy,
    sessionPersistence: data.session_persistence,
    concurrencyPolicy: data.concurrency_policy,
    maxConcurrentRuns: data.max_concurrent_runs,
    maxRetainedRuns: data.max_retained_runs,
    timeoutMinutes: data.timeout_minutes,
    usedEnvBundles:
      data.used_env_bundles !== undefined
        ? create(EnvBundleListSchema, { names: data.used_env_bundles ?? [] })
        : undefined,
  });
  const bytes = toBinary(UpdateWorkflowRequestSchema, req);
  const respBytes = await getWorkflowService().updateWorkflowConnect(bytes);
  return fromProtoWorkflow(fromBinary(WorkflowSchema, new Uint8Array(respBytes)));
}

export async function deleteWorkflow(orgSlug: string, workflowSlug: string): Promise<void> {
  const req = create(DeleteWorkflowRequestSchema, { orgSlug, workflowSlug });
  const bytes = toBinary(DeleteWorkflowRequestSchema, req);
  const respBytes = await getWorkflowService().deleteWorkflowConnect(bytes);
  fromBinary(DeleteWorkflowResponseSchema, new Uint8Array(respBytes));
}

async function workflowAction(
  caller: (b: Uint8Array) => Promise<Uint8Array>,
  orgSlug: string,
  workflowSlug: string,
): Promise<WorkflowData> {
  const req = create(WorkflowActionRequestSchema, { orgSlug, workflowSlug });
  const bytes = toBinary(WorkflowActionRequestSchema, req);
  const respBytes = await caller(bytes);
  return fromProtoWorkflow(fromBinary(WorkflowSchema, new Uint8Array(respBytes)));
}

export async function enableWorkflow(orgSlug: string, workflowSlug: string): Promise<WorkflowData> {
  return workflowAction((b) => getWorkflowService().enableWorkflowConnect(b), orgSlug, workflowSlug);
}

export async function disableWorkflow(orgSlug: string, workflowSlug: string): Promise<WorkflowData> {
  return workflowAction((b) => getWorkflowService().disableWorkflowConnect(b), orgSlug, workflowSlug);
}

export interface TriggerWorkflowResult {
  run?: WorkflowRunData;
  skipped?: boolean;
  reason?: string;
}

export async function triggerWorkflow(
  orgSlug: string,
  workflowSlug: string,
  variables?: Record<string, unknown>,
): Promise<TriggerWorkflowResult> {
  const req = create(TriggerWorkflowRequestSchema, {
    orgSlug,
    workflowSlug,
    variablesJson: variables ? JSON.stringify(variables) : "",
  });
  const bytes = toBinary(TriggerWorkflowRequestSchema, req);
  const respBytes = await getWorkflowService().triggerWorkflowConnect(bytes);
  const resp = fromBinary(TriggerWorkflowResponseSchema, new Uint8Array(respBytes));
  if (resp.skipped) {
    return { skipped: true, reason: resp.reason };
  }
  return resp.run ? { run: fromProtoWorkflowRun(resp.run) } : {};
}

export async function listWorkflowRuns(
  orgSlug: string,
  workflowSlug: string,
  filters?: { status?: string; limit?: number; offset?: number },
): Promise<{ items: WorkflowRunData[]; total: number }> {
  const req = create(ListWorkflowRunsRequestSchema, {
    orgSlug,
    workflowSlug,
    status: filters?.status ?? "",
    offset: filters?.offset,
    limit: filters?.limit,
  });
  const bytes = toBinary(ListWorkflowRunsRequestSchema, req);
  const respBytes = await getWorkflowService().listWorkflowRunsConnect(bytes);
  const resp = fromBinary(ListWorkflowRunsResponseSchema, new Uint8Array(respBytes));
  return { items: resp.items.map(fromProtoWorkflowRun), total: Number(resp.total) };
}

// Raw wire bytes for the runs fetch→state path: response → apply_fetched_runs /
// apply_appended_runs (Rust set_runs/append_runs via run_from_proto).
export async function listWorkflowRunsRaw(
  orgSlug: string,
  workflowSlug: string,
  filters?: { status?: string; limit?: number; offset?: number },
): Promise<Uint8Array> {
  const req = create(ListWorkflowRunsRequestSchema, {
    orgSlug, workflowSlug, status: filters?.status ?? "", offset: filters?.offset, limit: filters?.limit,
  });
  return new Uint8Array(
    await getWorkflowService().listWorkflowRunsConnect(toBinary(ListWorkflowRunsRequestSchema, req)),
  );
}

export async function cancelWorkflowRun(
  orgSlug: string,
  workflowSlug: string,
  runId: number,
): Promise<void> {
  const req = create(CancelWorkflowRunRequestSchema, { orgSlug, workflowSlug, runId: BigInt(runId) });
  const bytes = toBinary(CancelWorkflowRunRequestSchema, req);
  const respBytes = await getWorkflowService().cancelWorkflowRunConnect(bytes);
  fromBinary(CancelWorkflowRunResponseSchema, new Uint8Array(respBytes));
}
