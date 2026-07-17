import { create } from "@bufbuild/protobuf";
import { describe, expect, it } from "vitest";
import { RunnerSchema } from "@proto/runner_api/v1/runner_pb";
import { runnerToCache } from "./runner";

describe("runnerToCache", () => {
  it("preserves runner tunnel state for the detail view", () => {
    const runner = create(RunnerSchema, {
      id: BigInt(19),
      nodeId: "dev-runner-gemini",
      status: "online",
      currentPods: 0,
      maxConcurrentPods: 1,
      clusterId: BigInt(3),
      tunnelState: "connected",
      tunnelLastSeenAt: "2026-07-13T05:54:14Z",
    });

    expect(runnerToCache(runner)).toMatchObject({
      cluster_id: 3,
      tunnel_state: "connected",
      tunnel_last_seen_at: "2026-07-13T05:54:14Z",
    });
  });
});
