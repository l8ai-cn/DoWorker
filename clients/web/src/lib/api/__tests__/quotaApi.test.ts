import { beforeEach, describe, expect, it, vi } from "vitest";
import { createVirtualKey } from "../quotaApi";

vi.mock("@/lib/env", () => ({
  getApiBaseUrl: () => "http://localhost:10000/api",
}));

vi.mock("@/lib/wasm-core", () => ({
  getAuthManager: () => ({ get_token: () => "test-token" }),
}));

vi.mock("@/stores/auth", () => ({
  readCurrentOrg: () => ({ slug: "acme" }),
}));

describe("quotaApi", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it("creates virtual keys with an exact model resource ID", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({
        token: "dwk_test",
        key: {
          id: 7,
          name: "Build",
          key_prefix: "dwk_test",
          model_resource_id: 42,
          status: "active",
          created_at: "2026-07-10T00:00:00Z",
        },
      }), { status: 201 }),
    );

    await createVirtualKey({
      name: "Build",
      model_resource_id: 42,
      token_budget: 1000,
    });

    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:10000/api/v1/virtual-keys",
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({
          name: "Build",
          model_resource_id: 42,
          token_budget: 1000,
        }),
      }),
    );
  });
});
