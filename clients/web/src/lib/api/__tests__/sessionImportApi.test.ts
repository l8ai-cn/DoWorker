import { beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("@/lib/env", () => ({
  getApiBaseUrl: () => "http://localhost:10000/api",
}));

vi.mock("@/lib/wasm-core", () => ({
  getAuthManager: () => ({ get_token: () => "test-token" }),
}));

vi.mock("@/stores/auth", () => ({
  readCurrentOrg: () => ({ slug: "acme" }),
}));

import { fetchSessionByPodKey } from "../sessionImportApi";

describe("fetchSessionByPodKey", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it("returns null only for an absent session association", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(null, { status: 204 }),
    );

    await expect(fetchSessionByPodKey("worker-pod")).resolves.toBeNull();
  });

  it("returns the associated session metadata", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ id: "conv_123", title: "Imported" }), {
        status: 200,
      }),
    );

    await expect(fetchSessionByPodKey("worker-pod")).resolves.toEqual({
      id: "conv_123",
      title: "Imported",
    });
  });

  it("surfaces server failures instead of treating them as absence", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response("database unavailable", { status: 500 }),
    );

    await expect(fetchSessionByPodKey("worker-pod")).rejects.toThrow(
      "database unavailable",
    );
  });
});
