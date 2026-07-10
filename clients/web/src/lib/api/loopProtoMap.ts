// LoopData (web snake_case) ↔ proto.loop.v1.Loop (camelCase) mapping.
// `fromProtoLoop` lives in connect/loopConnect.ts (used by Connect-RPC
// adapters). This module adds `loopToProtoLoop` + `loopRunToProtoLoopRun`,
// used by the store when handing payloads to the wasm bridge encoded as
// proto bytes (per proto-state contract).

import { create as protoCreate } from "@bufbuild/protobuf";
import {
  LoopSchema, LoopRunSchema,
  type Loop as ProtoLoop, type LoopRun as ProtoLoopRun,
} from "@proto/loop/v1/loop_pb";
import type { LoopData, LoopRunData } from "@/lib/viewModels/loop";

function asBigInt(v: number | undefined | null): bigint | undefined {
  return v === undefined || v === null ? undefined : BigInt(v);
}

function toJsonString(v: unknown): string {
  if (v === undefined || v === null) return "";
  if (typeof v === "string") return v;
  return JSON.stringify(v);
}

export function loopToProtoLoop(l: LoopData): ProtoLoop {
  return protoCreate(LoopSchema, {
    id: asBigInt(l.id) ?? BigInt(0),
    slug: l.slug,
    name: l.name,
    description: l.description ?? "",
    agentSlug: l.agent_slug ?? "",
    permissionMode: l.permission_mode,
    promptTemplate: l.prompt_template,
    configOverridesJson: toJsonString(l.config_overrides),
    promptVariablesJson: toJsonString(l.prompt_variables),
    executionMode: l.execution_mode,
    cronExpression: l.cron_expression ?? "",
    autopilotConfigJson: toJsonString(l.autopilot_config),
    callbackUrl: l.callback_url ?? "",
    repositoryId: asBigInt(l.repository_id),
    runnerId: asBigInt(l.runner_id),
    branchName: l.branch_name ?? "",
    ticketId: asBigInt(l.ticket_id),
    modelResourceId: asBigInt(l.model_resource_id),
    status: l.status,
    sandboxStrategy: l.sandbox_strategy,
    sessionPersistence: l.session_persistence,
    concurrencyPolicy: l.concurrency_policy,
    maxConcurrentRuns: l.max_concurrent_runs,
    maxRetainedRuns: l.max_retained_runs,
    timeoutMinutes: l.timeout_minutes,
    idleTimeoutSec: 0,
    totalRuns: BigInt(l.total_runs),
    successfulRuns: BigInt(l.successful_runs),
    failedRuns: BigInt(l.failed_runs),
    activeRunCount: BigInt(l.active_run_count),
    avgDurationSec: l.avg_duration_sec,
    lastRunAt: l.last_run_at,
    createdAt: l.created_at,
    updatedAt: l.updated_at,
    usedEnvBundles: l.used_env_bundles ?? [],
  });
}

export function loopRunToProtoLoopRun(r: LoopRunData): ProtoLoopRun {
  return protoCreate(LoopRunSchema, {
    id: BigInt(r.id),
    loopId: BigInt(r.loop_id),
    runNumber: BigInt(r.run_number),
    status: r.status,
    podKey: r.pod_key,
    startedAt: r.started_at,
    completedAt: r.finished_at,
    errorMessage: r.error_message,
    createdAt: r.created_at,
  });
}
