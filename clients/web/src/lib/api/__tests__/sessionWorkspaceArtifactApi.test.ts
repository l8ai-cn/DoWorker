import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("@/lib/env", () => ({
  getApiBaseUrl: () => "https://api.example.test",
}));

vi.mock("@/lib/wasm-core", () => ({
  getAuthManager: () => ({ get_token: () => "test-token" }),
}));

vi.mock("@/stores/auth", () => ({
  readCurrentOrg: () => ({ slug: "dev-org" }),
}));

import { loadSessionWorkspaceArtifactById } from "../sessionWorkspaceArtifactApi";

describe("sessionWorkspaceArtifactApi", () => {
  const originalFetch = globalThis.fetch;

  beforeEach(() => {
    globalThis.fetch = vi.fn();
  });

  afterEach(() => {
    globalThis.fetch = originalFetch;
  });

  it("loads workspace artifacts from the raw content endpoint", async () => {
    vi.mocked(globalThis.fetch).mockResolvedValueOnce(
      new Response(new Uint8Array([0, 1, 2, 3]), {
        status: 200,
        headers: { "Content-Type": "video/mp4" },
      }),
    );

    const blob = await loadSessionWorkspaceArtifactById(
      "session-1",
      "output/demo clip.mp4",
    );

    expect(blob.type).toBe("video/mp4");
    expect(blob.size).toBe(4);
    expect(globalThis.fetch).toHaveBeenCalledWith(
      "https://api.example.test/v1/sessions/session-1/resources/environments/workspace/artifacts/content/output/demo%20clip.mp4",
      {
        headers: {
          Authorization: "Bearer test-token",
          "X-Organization-Slug": "dev-org",
        },
      },
    );
  });

  it("rejects a failed raw artifact request", async () => {
    vi.mocked(globalThis.fetch).mockResolvedValueOnce(
      new Response(null, { status: 503 }),
    );

    await expect(
      loadSessionWorkspaceArtifactById("session-1", "output/demo.mp4"),
    ).rejects.toThrow("Workspace artifact request failed (503)");
  });
});
