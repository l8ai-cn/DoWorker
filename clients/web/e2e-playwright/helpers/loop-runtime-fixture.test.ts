import { beforeEach, describe, expect, it, vi } from "vitest";
import type { ApiFixture } from "../fixtures/api.fixture";
import type { DbFixture } from "../fixtures/db.fixture";

const mocks = vi.hoisted(() => ({
  createE2EEchoPod: vi.fn(),
  unregisterE2ECreatedPod: vi.fn(),
}));

vi.mock("./e2e-worker-spec", () => ({
  createE2EEchoPod: mocks.createE2EEchoPod,
}));

vi.mock("./pod-cleanup", () => ({
  unregisterE2ECreatedPod: mocks.unregisterE2ECreatedPod,
}));

import { createLoopRuntimeFixture } from "./loop-runtime-fixture";

const fixture = {
  alias: "Loop runtime",
  goalLoopName: "loop-runtime",
  snapshotId: "7",
};
const sourcePodKey = "pod-loop-runtime";

describe("Loop runtime fixture cleanup", () => {
  beforeEach(() => {
    mocks.createE2EEchoPod.mockReset();
    mocks.unregisterE2ECreatedPod.mockReset();
  });

  it("keeps global cleanup ownership when persisted artifact lookup fails", async () => {
    mocks.createE2EEchoPod.mockResolvedValue({ pod: { podKey: sourcePodKey } });
    const db = { queryValue: vi.fn().mockReturnValue(null) } as unknown as DbFixture;
    const api = { connect: vi.fn().mockResolvedValue({}) } as unknown as ApiFixture;

    await expect(createLoopRuntimeFixture(db, api)).rejects.toThrow(
      "lacks a persisted dependency artifact",
    );
    expect(mocks.unregisterE2ECreatedPod).not.toHaveBeenCalled();
  });

  it("keeps global cleanup ownership when source termination fails", async () => {
    const terminatePod = vi.fn().mockRejectedValue(new Error("terminate failed"));
    mocks.createE2EEchoPod.mockResolvedValue({ pod: { podKey: sourcePodKey } });
    const db = {
      queryValue: vi.fn().mockReturnValue(
        `7|${fixture.alias}|sha256:${"a".repeat(64)}`,
      ),
    } as unknown as DbFixture;
    const api = {
      connect: vi.fn().mockResolvedValue({ pod: { terminatePod } }),
    } as unknown as ApiFixture;

    await expect(createLoopRuntimeFixture(db, api)).rejects.toThrow(
      "terminate failed",
    );
    expect(mocks.unregisterE2ECreatedPod).not.toHaveBeenCalled();
  });

  it("releases global cleanup ownership after source termination succeeds", async () => {
    const terminatePod = vi.fn().mockResolvedValue(undefined);
    mocks.createE2EEchoPod.mockResolvedValue({ pod: { podKey: sourcePodKey } });
    const db = {
      queryValue: vi.fn().mockReturnValue(
        `7|${fixture.alias}|sha256:${"a".repeat(64)}`,
      ),
    } as unknown as DbFixture;
    const api = {
      connect: vi.fn().mockResolvedValue({ pod: { terminatePod } }),
    } as unknown as ApiFixture;

    await createLoopRuntimeFixture(db, api);

    expect(terminatePod).toHaveBeenCalledWith({
      orgSlug: "dev-org",
      podKey: sourcePodKey,
    });
    expect(mocks.unregisterE2ECreatedPod).toHaveBeenCalledWith(sourcePodKey);
  });
});
