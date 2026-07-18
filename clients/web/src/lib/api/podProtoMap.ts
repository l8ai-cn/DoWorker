// PodData (web snake_case shape) ↔ proto.pod.v1.Pod (camelCase) mapping.
// `fromProtoPod` lives in podConnect.ts for historical reasons (it's used
// only by Connect-RPC adapters). This module adds the reverse direction
// `podToProtoPod`, used by the store layer when handing pod payloads to
// the wasm bridge encoded as proto bytes (per the proto-state contract).
//
// Both directions are intentionally non-symmetric:
//   * Server emits camelCase via prost; UI consumes snake_case via PodData.
//   * fromProtoPod normalises ProtoPod → PodData; podToProtoPod denormalises
//     PodData → partial ProtoPod (BigInt for int64, dropped wire-only fields).

import { create as protoCreate } from "@bufbuild/protobuf";
import {
  PodSchema,
  PodRunnerInfoSchema, PodAgentInfoSchema, PodRepositoryInfoSchema,
  PodTicketInfoSchema, PodWorkflowInfoSchema, PodCreatedByInfoSchema,
  type Pod as ProtoPod,
} from "@proto/pod/v1/pod_pb";
import type { PodData } from "@/lib/api/facade/pod";

function asBigInt(v: number | undefined | null): bigint | undefined {
  return v === undefined || v === null ? undefined : BigInt(v);
}

export function podToProtoPod(p: PodData): ProtoPod {
  return protoCreate(PodSchema, {
    id: asBigInt(p.id) ?? BigInt(0),
    podKey: p.pod_key,
    status: p.status,
    agentStatus: p.agent_status,
    alias: p.alias,
    title: p.title,
    runner: p.runner ? protoCreate(PodRunnerInfoSchema, {
      id: asBigInt(p.runner.id),
      nodeId: p.runner.node_id,
      status: p.runner.status,
    }) : undefined,
    agent: p.agent ? protoCreate(PodAgentInfoSchema, {
      name: p.agent.name, slug: p.agent.slug,
    }) : undefined,
    repository: p.repository ? protoCreate(PodRepositoryInfoSchema, {
      id: asBigInt(p.repository.id),
      name: p.repository.name,
      slug: p.repository.slug,
      providerType: p.repository.provider_type,
    }) : undefined,
    ticket: p.ticket ? protoCreate(PodTicketInfoSchema, {
      id: asBigInt(p.ticket.id),
      slug: p.ticket.slug,
      title: p.ticket.title,
    }) : undefined,
    workflow: p.workflow ? protoCreate(PodWorkflowInfoSchema, {
      id: asBigInt(p.workflow.id),
      name: p.workflow.name,
      slug: p.workflow.slug,
    }) : undefined,
    createdBy: p.created_by ? protoCreate(PodCreatedByInfoSchema, {
      id: asBigInt(p.created_by.id),
      username: p.created_by.username,
      name: p.created_by.name,
    }) : undefined,
    prompt: p.prompt,
    branchName: p.branch_name,
    sandboxPath: p.sandbox_path,
    interactionMode: p.interaction_mode,
    perpetual: p.perpetual,
    restartCount: p.restart_count,
    lastRestartAt: p.last_restart_at,
    startedAt: p.started_at,
    finishedAt: p.finished_at,
    lastActivity: p.last_activity,
    createdAt: p.created_at,
    errorCode: p.error_code,
    errorMessage: p.error_message,
    previewPort: p.preview_port,
    previewPath: p.preview_path,
    workerSpecSnapshotId: asBigInt(p.worker_spec_snapshot_id),
    workerSkillSlugs: p.worker_skill_slugs,
    sessionId: p.session_id,
  });
}

// Re-export so consumers have a single map module entry point.
export { fromProtoPod } from "@/lib/api/connect/podConnect";
