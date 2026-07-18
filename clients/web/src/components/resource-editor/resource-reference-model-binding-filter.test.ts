import { beforeEach, describe, expect, it, vi } from "vitest";

const api = vi.hoisted(() => ({
  listResources: vi.fn(),
}));

vi.mock("@/lib/api/facade/orchestrationResource", () => api);

import { loadResourceReferenceCatalog } from "./resource-reference-catalog-loader";

describe("loadResourceReferenceCatalog ModelBinding filter", () => {
  beforeEach(() => {
    api.listResources.mockReset();
  });

  it("uses the server-owned MiniMax protocol filter", async () => {
    api.listResources.mockImplementation((
      _orgSlug: string,
      input: { kind?: string; modelBindingFilter?: { workerType: string } },
    ) => {
      if (input.kind !== "ModelBinding") return Promise.resolve({ items: [] });
      expect(input.modelBindingFilter).toEqual({ workerType: "minimax-cli" });
      return Promise.resolve({
        items: [{
          identity: { target: { name: "minimax-chat" } },
          displayName: "MiniMax Chat",
          revision: 8n,
        }],
        appliedModelBindingFilter: {
          workerType: "minimax-cli",
          protocolAdapters: ["minimax"],
        },
      });
    });

    const catalog = await loadResourceReferenceCatalog(
      "acme",
      "minimax-cli",
      ["minimax"],
      [],
    );

    expect(catalog.byKind.ModelBinding).toEqual([{
      name: "minimax-chat",
      displayName: "MiniMax Chat",
      revision: 8,
    }]);
  });

  it("rejects a ModelBinding response without the server filter confirmation", async () => {
    api.listResources.mockImplementation((
      _orgSlug: string,
      input: { kind?: string },
    ) => Promise.resolve(input.kind === "ModelBinding"
      ? {
          items: [{
            identity: { target: { name: "unfiltered-model" } },
            displayName: "Unfiltered model",
            revision: 1n,
          }],
        }
      : { items: [] }));

    const catalog = await loadResourceReferenceCatalog(
      "acme",
      "minimax-cli",
      ["minimax"],
      [],
    );

    expect(catalog.byKind.ModelBinding).toBeUndefined();
    expect(catalog.errorsByKind.ModelBinding).toBe(
      "The control plane did not apply the ModelBinding protocol filter.",
    );
  });
});
