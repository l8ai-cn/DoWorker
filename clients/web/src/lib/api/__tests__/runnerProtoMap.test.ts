// runnerProtoMap converter test — RunnerData (snake_case view-model) →
// proto.runner_api.v1.Runner (camelCase + BigInt). Verifies field-name
// translation, optional-field handling, and the host_info JSON-stringify
// boundary.

import { describe, it, expect } from "vitest";
import { runnerToProtoRunner } from "../runnerProtoMap";
import type { RunnerData } from "@/lib/viewModels/runner";

const fixture: RunnerData = {
  id: 42,
  node_id: "node-abc",
  description: "primary runner",
  status: "online",
  last_heartbeat: "2026-05-27T10:00:00Z",
  current_pods: 3,
  max_concurrent_pods: 8,
  runner_version: "1.2.3",
  is_enabled: true,
  visibility: "organization",
  registered_by_user_id: 99,
  host_info: { os: "linux", arch: "x86_64", cpu_cores: 8, memory: 16 * 1024 * 1024 * 1024 },
  available_agents: ["claude-code", "codex"],
  tags: ["prod", "us-east"],
  created_at: "2026-05-01T00:00:00Z",
  updated_at: "2026-05-27T10:00:00Z",
};

describe("runnerToProtoRunner", () => {
  it("translates snake_case scalars to camelCase + BigInt for int64s", () => {
    const proto = runnerToProtoRunner(fixture);
    expect(proto.id).toBe(BigInt(42));
    expect(proto.nodeId).toBe("node-abc");
    expect(proto.description).toBe("primary runner");
    expect(proto.status).toBe("online");
    expect(proto.lastHeartbeat).toBe("2026-05-27T10:00:00Z");
    expect(proto.currentPods).toBe(3);
    expect(proto.maxConcurrentPods).toBe(8);
    expect(proto.runnerVersion).toBe("1.2.3");
    expect(proto.isEnabled).toBe(true);
    expect(proto.visibility).toBe("organization");
    expect(proto.registeredByUserId).toBe(BigInt(99));
    expect(proto.availableAgents).toEqual(["claude-code", "codex"]);
    expect(proto.tags).toEqual(["prod", "us-east"]);
    expect(proto.createdAt).toBe("2026-05-01T00:00:00Z");
    expect(proto.updatedAt).toBe("2026-05-27T10:00:00Z");
  });

  it("stringifies host_info into hostInfoJson (UI-owned schema boundary)", () => {
    const proto = runnerToProtoRunner(fixture);
    expect(typeof proto.hostInfoJson).toBe("string");
    const parsed = JSON.parse(proto.hostInfoJson);
    expect(parsed).toEqual(fixture.host_info);
  });

  it("defaults host_info to empty string when absent", () => {
    const minimal: RunnerData = { ...fixture, host_info: undefined };
    const proto = runnerToProtoRunner(minimal);
    expect(proto.hostInfoJson).toBe("");
  });

  it("handles optional numeric fields — id 0 when undefined", () => {
    // RunnerData.id is required `number`, but Rust treats id=0 as sentinel
    // for unsaved; the proto map must keep BigInt(0) rather than mis-typing.
    const zeroId: RunnerData = { ...fixture, id: 0 };
    const proto = runnerToProtoRunner(zeroId);
    expect(proto.id).toBe(BigInt(0));
  });

  it("registeredByUserId stays undefined when source field absent", () => {
    const noUser: RunnerData = { ...fixture, registered_by_user_id: undefined };
    const proto = runnerToProtoRunner(noUser);
    expect(proto.registeredByUserId).toBeUndefined();
  });

  it("defaults available_agents/tags to [] when source absent", () => {
    const noAgents: RunnerData = { ...fixture, available_agents: undefined, tags: undefined };
    const proto = runnerToProtoRunner(noAgents);
    expect(proto.availableAgents).toEqual([]);
    expect(proto.tags).toEqual([]);
  });

  it("organizationId hard-coded to 0 — backend infers from session, not denormalized here", () => {
    // The view-model doesn't carry organization_id (snake_case shape only has
    // it on the wire); the proto map respects that by sending 0 and letting
    // the Rust state pick it up from the auth context.
    const proto = runnerToProtoRunner(fixture);
    expect(proto.organizationId).toBe(BigInt(0));
  });

  it("preserves the runner tunnel evidence used by the detail view", () => {
    const runner = {
      ...fixture,
      cluster_id: 19,
      tunnel_state: "connected",
      tunnel_last_seen_at: "2026-07-13T05:54:14Z",
      tunnel_last_error: undefined,
    };
    const proto = runnerToProtoRunner(runner);
    const tunnelProto = proto as unknown as {
      clusterId?: bigint;
      tunnelState?: string;
      tunnelLastSeenAt?: string;
      tunnelLastError?: string;
    };

    expect(tunnelProto.clusterId).toBe(BigInt(19));
    expect(tunnelProto.tunnelState).toBe("connected");
    expect(tunnelProto.tunnelLastSeenAt).toBe("2026-07-13T05:54:14Z");
    expect(tunnelProto.tunnelLastError).toBeUndefined();
  });
});
