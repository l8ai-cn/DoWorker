import { describe, expect, it } from "vitest";
import { podToProtoPod } from "./podProtoMap";

describe("podToProtoPod", () => {
  it("preserves the Worker snapshot association during cache replacement", () => {
    const pod = podToProtoPod({
      id: 1,
      pod_key: "video-worker-1",
      status: "running",
      worker_spec_snapshot_id: 91,
    });

    expect(pod.workerSpecSnapshotId).toBe(BigInt(91));
  });
});
