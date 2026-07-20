import { beforeEach, describe, expect, it, vi } from "vitest";
import type { ApiFixture } from "../fixtures/api.fixture";
import type { DbFixture } from "../fixtures/db.fixture";

const mocks = vi.hoisted(() => ({
  terminateRegisteredE2EPods: vi.fn(),
}));

vi.mock("./pod-cleanup", () => ({
  terminateRegisteredE2EPods: mocks.terminateRegisteredE2EPods,
}));

import { cleanupResourceWorkflowFixture } from "./resource-workflow-run-cleanup";

describe("resource Workflow run cleanup", () => {
  beforeEach(() => {
    mocks.terminateRegisteredE2EPods.mockReset().mockResolvedValue(0);
  });

  it("cancels active runs with and without a Pod before resetting the fixture", async () => {
    const cancelWorkflowRun = vi.fn().mockResolvedValue({ message: "Run cancelled" });
    const db = {
      queryValue: vi.fn().mockReturnValue("17|pod-run\n18|"),
      cleanup: vi.fn(),
    } as unknown as DbFixture;
    const api = {
      connect: vi.fn().mockResolvedValue({ workflow: { cancelWorkflowRun } }),
    } as unknown as ApiFixture;

    await cleanupResourceWorkflowFixture(db, api);

    expect(cancelWorkflowRun).toHaveBeenNthCalledWith(1, {
      orgSlug: "dev-org",
      workflowSlug: "e2e-resource-workflow-v2",
      runId: 17n,
    });
    expect(cancelWorkflowRun).toHaveBeenNthCalledWith(2, {
      orgSlug: "dev-org",
      workflowSlug: "e2e-resource-workflow-v2",
      runId: 18n,
    });
    expect(mocks.terminateRegisteredE2EPods).toHaveBeenCalledOnce();
    expect(db.cleanup).toHaveBeenCalledOnce();
  });

  it("does not cancel terminal runs filtered by the production status contract", async () => {
    const db = {
      queryValue: vi.fn().mockReturnValue(null),
      cleanup: vi.fn(),
    } as unknown as DbFixture;
    const api = { connect: vi.fn() } as unknown as ApiFixture;

    await cleanupResourceWorkflowFixture(db, api);

    expect(db.queryValue).toHaveBeenCalledWith(
      expect.stringContaining("run.status IN ('pending', 'running')"),
    );
    expect(api.connect).not.toHaveBeenCalled();
    expect(mocks.terminateRegisteredE2EPods).toHaveBeenCalledOnce();
    expect(db.cleanup).toHaveBeenCalledOnce();
  });

  it("does not terminate source Pods or reset state when cancellation fails", async () => {
    const cancelWorkflowRun = vi.fn().mockRejectedValue(new Error("cancel failed"));
    const db = {
      queryValue: vi.fn().mockReturnValue("17|pod-run"),
      cleanup: vi.fn(),
    } as unknown as DbFixture;
    const api = {
      connect: vi.fn().mockResolvedValue({ workflow: { cancelWorkflowRun } }),
    } as unknown as ApiFixture;

    await expect(cleanupResourceWorkflowFixture(db, api)).rejects.toThrow("cancel failed");

    expect(mocks.terminateRegisteredE2EPods).not.toHaveBeenCalled();
    expect(db.cleanup).not.toHaveBeenCalled();
  });
});
