import { renderHook, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { WorkerCreateOptions } from "@/lib/api/facade/podConnect";
import { useWorkerCreateOptions } from "./useWorkerCreateOptions";

const listWorkerCreateOptions = vi.hoisted(() => vi.fn());

vi.mock("@/lib/api/facade/podConnect", () => ({
  listWorkerCreateOptions,
}));

describe("useWorkerCreateOptions", () => {
  beforeEach(() => {
    listWorkerCreateOptions.mockReset();
    listWorkerCreateOptions.mockImplementation(async (orgSlug: string) =>
      options(orgSlug)
    );
  });

  it("isolates loaded options by organization", async () => {
    const selection = {
      workerTypeSlug: "",
      computeTargetId: 0,
      deploymentMode: "",
    };
    const { result, rerender } = renderHook(
      ({ orgSlug }) => useWorkerCreateOptions(true, orgSlug, selection),
      { initialProps: { orgSlug: "acme" } },
    );

    await waitFor(() => {
      expect(result.current.status).toBe("ready");
    });
    expect(result.current.status === "ready" && result.current.data.revision)
      .toBe("acme-revision");

    rerender({ orgSlug: "globex" });
    expect(result.current.status).toBe("loading");
    await waitFor(() => {
      expect(result.current.status === "ready" && result.current.data.revision)
        .toBe("globex-revision");
    });
    expect(listWorkerCreateOptions.mock.calls.map(([orgSlug]) => orgSlug))
      .toEqual(["acme", "globex"]);
  });
});

function options(orgSlug: string): WorkerCreateOptions {
  return {
    revision: `${orgSlug}-revision`,
    worker_types: [],
    runtime_images: [],
    compute_targets: [],
    deployment_modes: [],
    resource_profiles: [],
  };
}
