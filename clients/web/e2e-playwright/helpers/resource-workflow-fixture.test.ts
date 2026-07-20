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

import { ensureResourceWorkflowFixture } from "./resource-workflow-fixture";

const podKey = "pod-resource-workflow-source";
const artifact = `7|sha256:${"a".repeat(64)}`;

describe("resource Workflow fixture source Pod ownership", () => {
  beforeEach(() => {
    mocks.createE2EEchoPod.mockReset().mockResolvedValue({ pod: { podKey } });
    mocks.unregisterE2ECreatedPod.mockReset();
  });

  it("keeps global cleanup ownership when persisted artifact lookup fails", async () => {
    const db = {
      queryValue: vi.fn().mockReturnValueOnce(null).mockReturnValueOnce(null),
    } as unknown as DbFixture;
    const api = { connect: vi.fn().mockResolvedValue({}) } as unknown as ApiFixture;

    await expect(ensureResourceWorkflowFixture(db, api)).rejects.toThrow(
      "source lacks a persisted dependency artifact",
    );

    expect(mocks.unregisterE2ECreatedPod).not.toHaveBeenCalled();
  });

  it("keeps global cleanup ownership when source termination fails", async () => {
    const terminatePod = vi.fn().mockRejectedValue(new Error("terminate failed"));
    const db = {
      queryValue: vi.fn().mockReturnValueOnce(null).mockReturnValueOnce(artifact),
    } as unknown as DbFixture;
    const api = {
      connect: vi.fn().mockResolvedValue({ pod: { terminatePod } }),
    } as unknown as ApiFixture;

    await expect(ensureResourceWorkflowFixture(db, api)).rejects.toThrow("terminate failed");

    expect(mocks.unregisterE2ECreatedPod).not.toHaveBeenCalled();
  });

  it("releases source Pod cleanup ownership only after termination succeeds", async () => {
    const terminatePod = vi.fn().mockResolvedValue(undefined);
    const db = {
      queryValue: vi.fn()
        .mockReturnValueOnce(null)
        .mockReturnValueOnce(artifact)
        .mockReturnValueOnce("1|2")
        .mockReturnValueOnce("3|4|7"),
    } as unknown as DbFixture;
    const api = {
      connect: vi.fn().mockResolvedValue({ pod: { terminatePod } }),
    } as unknown as ApiFixture;

    await expect(ensureResourceWorkflowFixture(db, api)).resolves.toMatchObject({
      snapshotId: "7",
      workflowId: "3",
    });

    expect(terminatePod).toHaveBeenCalledWith({
      orgSlug: "dev-org",
      podKey,
    });
    expect(mocks.unregisterE2ECreatedPod).toHaveBeenCalledWith(podKey);
  });
});
