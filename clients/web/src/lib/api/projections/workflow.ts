import type { Workflow as ProtoWorkflow, WorkflowRun as ProtoWorkflowRun } from "@agent-cloud/proto/workflow/v1/workflow_pb";
import type { WorkflowData, WorkflowRunData } from "@agent-cloud/service-interface";
import { parseJSONObject } from "./parse-json";

// Single source of truth for the proto.workflow.v1 → WorkflowData projection.

export function workflowToCache(p: ProtoWorkflow): WorkflowData {
  return {
    id: Number(p.id), organization_id: 0, name: p.name, slug: p.slug,
    description: p.description || undefined,
    agent_slug: p.agentSlug || undefined,
    permission_mode: p.permissionMode,
    prompt_template: p.promptTemplate,
    prompt_variables: parseJSONObject(p.promptVariablesJson),
    repository_id: p.repositoryId != null ? Number(p.repositoryId) : undefined,
    runner_id: p.runnerId != null ? Number(p.runnerId) : undefined,
    branch_name: p.branchName || undefined,
    ticket_id: p.ticketId != null ? Number(p.ticketId) : undefined,
    model_resource_id: p.modelResourceId != null ? Number(p.modelResourceId) : undefined,
    used_env_bundles: p.usedEnvBundles ?? [],
    config_overrides: parseJSONObject(p.configOverridesJson),
    execution_mode: p.executionMode as WorkflowData["execution_mode"],
    cron_expression: p.cronExpression || undefined,
    callback_url: p.callbackUrl || undefined,
    autopilot_config: parseJSONObject(p.autopilotConfigJson) ?? {},
    status: p.status as WorkflowData["status"],
    sandbox_strategy: p.sandboxStrategy as WorkflowData["sandbox_strategy"],
    session_persistence: p.sessionPersistence,
    concurrency_policy: p.concurrencyPolicy as WorkflowData["concurrency_policy"],
    max_concurrent_runs: p.maxConcurrentRuns,
    max_retained_runs: p.maxRetainedRuns,
    timeout_minutes: p.timeoutMinutes,
    created_by_id: 0,
    total_runs: Number(p.totalRuns), successful_runs: Number(p.successfulRuns),
    failed_runs: Number(p.failedRuns), active_run_count: Number(p.activeRunCount),
    avg_duration_sec: p.avgDurationSec ?? undefined,
    last_run_at: p.lastRunAt, created_at: p.createdAt, updated_at: p.updatedAt,
  };
}

export function workflowRunToCache(p: ProtoWorkflowRun): WorkflowRunData {
  return {
    id: Number(p.id), organization_id: 0, workflow_id: Number(p.workflowId),
    run_number: Number(p.runNumber), status: p.status as WorkflowRunData["status"],
    pod_key: p.podKey, trigger_type: "", started_at: p.startedAt,
    // proto WorkflowRun.completed_at maps to the viewModel's finished_at.
    finished_at: p.completedAt,
    error_message: p.errorMessage, created_at: p.createdAt, updated_at: p.createdAt,
  };
}
