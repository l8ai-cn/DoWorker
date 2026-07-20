import { beforeEach, describe, expect, it, vi } from "vitest";
import type { ApiFixture } from "../fixtures/api.fixture";
import type { DbFixture } from "../fixtures/db.fixture";

const mocks = vi.hoisted(() => ({
  applyE2EWorkerTemplate: vi.fn(),
  buildE2EEchoWorkerSpec: vi.fn(),
  createPod: vi.fn(),
  registerE2ECreatedPod: vi.fn(),
  unregisterE2ECreatedPod: vi.fn(),
}));

vi.mock("./e2e-worker-spec", () => ({
  buildE2EEchoWorkerSpec: mocks.buildE2EEchoWorkerSpec,
}));

vi.mock("./e2e-worker-template-resource", () => ({
  applyE2EWorkerTemplate: mocks.applyE2EWorkerTemplate,
}));

vi.mock("./pod-cleanup", () => ({
  registerE2ECreatedPod: mocks.registerE2ECreatedPod,
  unregisterE2ECreatedPod: mocks.unregisterE2ECreatedPod,
}));

import { createLoopRuntimeFixture } from "./loop-runtime-fixture";

const fixture = {
  goalLoopName: "loop-runtime",
  optionLabel: "Loop runtime · WorkerTemplate · 模板 loop-runtime",
  snapshotId: "7",
};
const sourcePodKey = "pod-loop-runtime";
const worker = {
  alias: "Loop runtime",
  automationLevel: "autonomous",
};

describe("Loop runtime fixture cleanup", () => {
  beforeEach(() => {
    mocks.applyE2EWorkerTemplate.mockReset().mockResolvedValue(fixture.snapshotId);
    mocks.buildE2EEchoWorkerSpec.mockReset().mockResolvedValue(worker);
    mocks.createPod.mockReset().mockResolvedValue({ pod: { podKey: sourcePodKey } });
    mocks.registerE2ECreatedPod.mockReset();
    mocks.unregisterE2ECreatedPod.mockReset();
  });

  it("keeps global cleanup ownership when persisted artifact lookup fails", async () => {
    const db = { queryValue: vi.fn().mockReturnValue(null) } as unknown as DbFixture;
    const api = {
      connect: vi.fn().mockResolvedValue({ pod: { createPod: mocks.createPod } }),
    } as unknown as ApiFixture;

    await expect(createLoopRuntimeFixture(db, api)).rejects.toThrow(
      "lacks a persisted dependency artifact",
    );
    expect(mocks.unregisterE2ECreatedPod).not.toHaveBeenCalled();
  });

  it("keeps global cleanup ownership when source termination fails", async () => {
    const terminatePod = vi.fn().mockRejectedValue(new Error("terminate failed"));
    const db = {
      queryValue: vi.fn()
        .mockReturnValueOnce(`7|${worker.alias}|sha256:${"a".repeat(64)}`)
        .mockReturnValueOnce(`7|sha256:${"a".repeat(64)}`),
    } as unknown as DbFixture;
    const api = {
      connect: vi.fn().mockResolvedValue({
        pod: { createPod: mocks.createPod, terminatePod },
      }),
    } as unknown as ApiFixture;

    await expect(createLoopRuntimeFixture(db, api)).rejects.toThrow(
      "terminate failed",
    );
    expect(mocks.unregisterE2ECreatedPod).not.toHaveBeenCalled();
  });

  it("releases global cleanup ownership after source termination succeeds", async () => {
    const terminatePod = vi.fn().mockResolvedValue(undefined);
    const db = {
      queryValue: vi.fn()
        .mockReturnValueOnce(`7|${worker.alias}|sha256:${"a".repeat(64)}`)
        .mockReturnValueOnce(`7|sha256:${"a".repeat(64)}`),
    } as unknown as DbFixture;
    const api = {
      connect: vi.fn().mockResolvedValue({
        pod: { createPod: mocks.createPod, terminatePod },
      }),
    } as unknown as ApiFixture;

    const created = await createLoopRuntimeFixture(db, api);

    expect(terminatePod).toHaveBeenCalledWith({
      orgSlug: "dev-org",
      podKey: sourcePodKey,
    });
    expect(mocks.applyE2EWorkerTemplate).toHaveBeenCalledWith(
      expect.anything(),
      expect.stringMatching(/^loop-runtime-[a-z]{12}$/),
      worker,
    );
    expect(mocks.registerE2ECreatedPod).toHaveBeenCalledWith(
      sourcePodKey,
      worker.alias,
    );
    expect(mocks.unregisterE2ECreatedPod).toHaveBeenCalledWith(sourcePodKey);
    expect(created.optionLabel).toMatch(
      /^Loop runtime · WorkerTemplate · 模板 loop-runtime-[a-z]{12}$/,
    );
  });
});
