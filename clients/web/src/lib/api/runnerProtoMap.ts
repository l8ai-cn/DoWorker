// RunnerData (web snake_case shape) ↔ proto.runner_api.v1.Runner (camelCase + BigInt).
// Mirror of podProtoMap.ts — denormalises renderer types into the proto wire
// shape so the wasm bridge can decode well-typed runner_state mutation
// requests via prost.
//
// Used by stores/runner.ts to encode mutation requests as binary bytes.

import { create as protoCreate } from "@bufbuild/protobuf";
import {
  RunnerSchema,
  type Runner as ProtoRunner,
} from "@proto/runner_api/v1/runner_pb";
import type { RunnerData } from "@/lib/viewModels/runner";

function asBigInt(v: number | undefined | null): bigint | undefined {
  return v === undefined || v === null ? undefined : BigInt(v);
}

function stringifyHost(host: RunnerData["host_info"]): string {
  if (!host) return "";
  try { return JSON.stringify(host); } catch { return ""; }
}

export function runnerToProtoRunner(r: RunnerData): ProtoRunner {
  return protoCreate(RunnerSchema, {
    id: asBigInt(r.id) ?? BigInt(0),
    nodeId: r.node_id,
    description: r.description ?? "",
    status: r.status,
    lastHeartbeat: r.last_heartbeat,
    currentPods: r.current_pods,
    maxConcurrentPods: r.max_concurrent_pods,
    runnerVersion: r.runner_version,
    isEnabled: r.is_enabled,
    visibility: r.visibility,
    registeredByUserId: asBigInt(r.registered_by_user_id),
    hostInfoJson: stringifyHost(r.host_info),
    availableAgents: r.available_agents ?? [],
    tags: r.tags ?? [],
    createdAt: r.created_at,
    updatedAt: r.updated_at,
    organizationId: BigInt(0),
    agentVersions: [],
    clusterId: asBigInt(r.cluster_id) ?? BigInt(0),
    tunnelState: r.tunnel_state ?? "",
    tunnelLastSeenAt: r.tunnel_last_seen_at,
    tunnelLastError: r.tunnel_last_error,
  });
}
