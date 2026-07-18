import { renderHook, waitFor } from "@testing-library/react";
import { EnvironmentBundlePurpose } from "@proto/orchestration_resource/v1/orchestration_resource_queries_pb";
import { beforeEach, describe, expect, it, vi } from "vitest";

const api = vi.hoisted(() => ({
  listResources: vi.fn(),
}));

vi.mock("@/lib/api/facade/orchestrationResource", () => api);

import { useResourceReferenceOptions } from "./use-resource-reference-options";
import { environmentBundleCatalogKey } from "./resource-reference-options";

describe("useResourceReferenceOptions", () => {
  beforeEach(() => {
    api.listResources.mockReset();
  });

  const makeResourceItems = (start: number, count: number) =>
    Array.from({ length: count }, (_, index) => ({
      identity: { target: { name: `worker-${start + index}` } },
      displayName: `Worker ${start + index}`,
      revision: BigInt(start + index),
    }));

  it("keeps successful kinds when one reference kind fails", async () => {
    api.listResources.mockImplementation((
      _orgSlug: string,
      input: { kind?: string },
    ) => {
      if (input.kind === "Prompt") {
        return Promise.reject(new Error(JSON.stringify({
          kind: "http",
          status: 403,
          code: "permission_denied",
          message: "Prompt references are not readable.",
        })));
      }
      if (input.kind === "WorkerTemplate") {
        return Promise.resolve({
          items: [{
            identity: { target: { name: "code-reviewer" } },
            displayName: "Code reviewer",
            revision: 4n,
          }],
        });
      }
      return Promise.resolve({ items: [] });
    });

    const { result } = renderHook(
      () => useResourceReferenceOptions("acme"),
    );

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.error).toBeNull();
    expect(result.current.byKind.WorkerTemplate).toEqual([{
      name: "code-reviewer",
      displayName: "Code reviewer",
      revision: 4,
    }]);
    expect(result.current.errorsByKind.Prompt).toBe(
      "Prompt references are not readable.",
    );
  });

  it("loads all pages according to API total", async () => {
    api.listResources.mockImplementation((
      _orgSlug: string,
      input: { kind?: string; offset?: number },
    ) => {
      if (input.kind !== "WorkerTemplate") {
        return Promise.resolve({ items: [] });
      }
      if (input.offset === 100) {
        return Promise.resolve({
          items: makeResourceItems(100, 1),
          total: 101n,
          offset: 100,
          limit: 100,
        });
      }
      return Promise.resolve({
        items: makeResourceItems(0, 100),
        total: 101n,
        offset: 0,
        limit: 100,
      });
    });

    const { result } = renderHook(
      () => useResourceReferenceOptions("acme"),
    );

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.byKind.WorkerTemplate).toHaveLength(101);
    expect(result.current.byKind.WorkerTemplate?.[0]).toMatchObject({
      name: "worker-0",
      revision: 0,
      displayName: "Worker 0",
    });
    expect(result.current.byKind.WorkerTemplate?.[100]).toMatchObject({
      name: "worker-100",
      revision: 100,
      displayName: "Worker 100",
    });
    expect(api.listResources).toHaveBeenCalledWith("acme", {
      kind: "WorkerTemplate",
      limit: 100,
      offset: 0,
      environmentBundleFilter: undefined,
    });
    expect(api.listResources).toHaveBeenCalledWith("acme", {
      kind: "WorkerTemplate",
      limit: 100,
      offset: 100,
      environmentBundleFilter: undefined,
    });
  });

  it("fails a catalog when a later page fails instead of returning partial options", async () => {
    api.listResources.mockImplementation((
      _orgSlug: string,
      input: { kind?: string; offset?: number },
    ) => {
      if (input.kind === "WorkerTemplate" && input.offset === 0) {
        return Promise.resolve({
          items: makeResourceItems(0, 100),
          total: 101n,
          offset: 0,
          limit: 100,
        });
      }
      if (input.kind === "WorkerTemplate" && input.offset === 100) {
        return Promise.reject(new Error("Worker catalog page unavailable."));
      }
      return Promise.resolve({ items: [] });
    });

    const { result } = renderHook(
      () => useResourceReferenceOptions("acme"),
    );

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.byKind.WorkerTemplate).toBeUndefined();
    expect(result.current.errorsByKind.WorkerTemplate).toBe(
      "Worker catalog page unavailable.",
    );
  });

  it("loads separate compatible EnvironmentBundle catalogs for a Worker type", async () => {
    api.listResources.mockImplementation((
      _orgSlug: string,
      input: {
        kind?: string;
        environmentBundleFilter?: {
          purpose: EnvironmentBundlePurpose;
          workerType: string;
          targetName?: string;
        };
      },
    ) => {
      const filter = input.environmentBundleFilter;
      if (input.kind !== "EnvironmentBundle" || !filter) {
        return Promise.resolve({ items: [] });
      }
      return Promise.resolve({
        items: [{
          identity: {
            target: {
              name: filter.targetName
                ? `bundle-${filter.targetName}`
                : `bundle-${filter.purpose}`,
            },
          },
          displayName: filter.workerType,
          revision: 2n,
        }],
        appliedEnvironmentBundleFilter: {
          ...filter,
          targetName: filter.targetName ?? "",
        },
      });
    });

    const { result } = renderHook(
      () => useResourceReferenceOptions(
        "acme",
        "cursor-cli",
        ["CURSOR_REFRESH_TOKEN", "CURSOR_API_KEY"],
      ),
    );

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.byKind[
      environmentBundleCatalogKey("runtime")
    ]).toEqual([{
      name: `bundle-${EnvironmentBundlePurpose.RUNTIME}`,
      displayName: "cursor-cli",
      revision: 2,
    }]);
    expect(result.current.byKind[
      environmentBundleCatalogKey("config")
    ]).toEqual([{
      name: `bundle-${EnvironmentBundlePurpose.CONFIG}`,
      displayName: "cursor-cli",
      revision: 2,
    }]);
    expect(result.current.byKind[
      environmentBundleCatalogKey("credential", "CURSOR_API_KEY")
    ]).toEqual([{
      name: "bundle-CURSOR_API_KEY",
      displayName: "cursor-cli",
      revision: 2,
    }]);
    expect(result.current.byKind[
      environmentBundleCatalogKey("credential", "CURSOR_REFRESH_TOKEN")
    ]).toEqual([{
      name: "bundle-CURSOR_REFRESH_TOKEN",
      displayName: "cursor-cli",
      revision: 2,
    }]);
    expect(api.listResources).toHaveBeenCalledWith("acme", {
      kind: "EnvironmentBundle",
      limit: 100,
      offset: 0,
      environmentBundleFilter: {
        purpose: EnvironmentBundlePurpose.CREDENTIAL,
        workerType: "cursor-cli",
        targetName: "CURSOR_API_KEY",
      },
    });
    expect(api.listResources).toHaveBeenCalledWith("acme", {
      kind: "EnvironmentBundle",
      limit: 100,
      offset: 0,
      environmentBundleFilter: {
        purpose: EnvironmentBundlePurpose.CREDENTIAL,
        workerType: "cursor-cli",
        targetName: "CURSOR_REFRESH_TOKEN",
      },
    });
  });

  it("isolates failures between credential target catalogs", async () => {
    api.listResources.mockImplementation((
      _orgSlug: string,
      input: {
        kind?: string;
        environmentBundleFilter?: {
          purpose: EnvironmentBundlePurpose;
          workerType: string;
          targetName?: string;
        };
      },
    ) => {
      const filter = input.environmentBundleFilter;
      if (!filter) return Promise.resolve({ items: [] });
      if (filter.targetName === "BROKEN_KEY") {
        return Promise.reject(new Error("Credential catalog unavailable."));
      }
      return Promise.resolve({
        items: filter.targetName === "WORKING_KEY"
          ? [{
              identity: { target: { name: "working-bundle" } },
              displayName: "Working bundle",
              revision: 5n,
            }]
          : [],
        appliedEnvironmentBundleFilter: {
          ...filter,
          targetName: filter.targetName ?? "",
        },
      });
    });

    const { result } = renderHook(
      () => useResourceReferenceOptions(
        "acme",
        "do-agent",
        ["WORKING_KEY", "BROKEN_KEY"],
      ),
    );

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.byKind[
      environmentBundleCatalogKey("credential", "WORKING_KEY")
    ]).toEqual([{
      name: "working-bundle",
      displayName: "Working bundle",
      revision: 5,
    }]);
    expect(result.current.errorsByKind[
      environmentBundleCatalogKey("credential", "BROKEN_KEY")
    ]).toBe("Credential catalog unavailable.");
  });

  it("ignores late catalogs from a previously selected Worker type", async () => {
    let resolveOldCatalog: ((value: { items: never[] }) => void) | undefined;
    const oldCatalog = new Promise<{ items: never[] }>((resolve) => {
      resolveOldCatalog = resolve;
    });
    api.listResources.mockImplementation((
      _orgSlug: string,
      input: {
        kind?: string;
        environmentBundleFilter?: {
          purpose: EnvironmentBundlePurpose;
          workerType: string;
          targetName?: string;
        };
      },
    ) => {
      const filter = input.environmentBundleFilter;
      if (!filter) return Promise.resolve({ items: [] });
      if (filter.workerType === "cursor-cli") return oldCatalog;
      return Promise.resolve({
        items: [{
          identity: {
            target: {
              name: `${filter.workerType}-${filter.purpose}`,
            },
          },
          displayName: filter.workerType,
          revision: 3n,
        }],
        appliedEnvironmentBundleFilter: {
          ...filter,
          targetName: filter.targetName ?? "",
        },
      });
    });

    const { result, rerender } = renderHook(
      ({ workerType }) => useResourceReferenceOptions("acme", workerType),
      { initialProps: { workerType: "cursor-cli" } },
    );

    rerender({ workerType: "do-agent" });
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.byKind[
      environmentBundleCatalogKey("config")
    ]?.[0]?.name).toBe(
      `do-agent-${EnvironmentBundlePurpose.CONFIG}`,
    );

    resolveOldCatalog?.({ items: [] });
    await Promise.resolve();
    expect(result.current.byKind[
      environmentBundleCatalogKey("config")
    ]?.[0]?.name).toBe(
      `do-agent-${EnvironmentBundlePurpose.CONFIG}`,
    );
  });

  it("rejects catalogs when the control plane does not confirm filtering", async () => {
    api.listResources.mockImplementation((
      _orgSlug: string,
      input: { kind?: string; environmentBundleFilter?: unknown },
    ) => Promise.resolve(input.environmentBundleFilter
      ? {
          items: [{
            identity: { target: { name: "unfiltered-bundle" } },
            displayName: "Unfiltered bundle",
            revision: 1n,
          }],
        }
      : { items: [] }));

    const { result } = renderHook(
      () => useResourceReferenceOptions("acme", "do-agent"),
    );

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.byKind[
      environmentBundleCatalogKey("config")
    ]).toBeUndefined();
    expect(result.current.errorsByKind[
      environmentBundleCatalogKey("config")
    ]).toBe(
      "The control plane did not apply the EnvironmentBundle reference filter.",
    );
  });
});
