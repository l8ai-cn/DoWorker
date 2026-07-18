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

import { loadSessionArtifactRepresentation } from "../sessionWorkspaceArtifactApi";

describe("session artifact API", () => {
  const originalFetch = globalThis.fetch;

  beforeEach(() => {
    globalThis.fetch = vi.fn();
  });

  afterEach(() => {
    globalThis.fetch = originalFetch;
  });

  it("loads an immutable representation by exact artifact identity", async () => {
    vi.mocked(globalThis.fetch).mockResolvedValueOnce(
      new Response(new Uint8Array([0, 1, 2, 3]), {
        status: 200,
        headers: { "Content-Type": "video/mp4" },
      }),
    );

    const blob = await loadSessionArtifactRepresentation({
      artifactId: "video-1",
      digest: "sha256:abc",
      representationId: "playable",
      resourceId: "session-file:file_12345678",
      revision: 3n,
      sessionId: "session-1",
    });

    expect(blob.type).toBe("video/mp4");
    expect(globalThis.fetch).toHaveBeenCalledWith(
      "https://api.example.test/v1/sessions/session-1/artifacts/content?" +
        "artifact_id=video-1&digest=sha256%3Aabc&" +
        "representation_id=playable&revision=3",
      {
        headers: {
          Authorization: "Bearer test-token",
          "X-Organization-Slug": "dev-org",
        },
      },
    );
  });

  it.each([
    "",
    "workspace:output/demo.mp4",
    "artifact-cache:preview-1",
    "session-file:",
  ])("rejects non-durable resource IDs: %j", async (resourceId) => {
    await expect(
      loadSessionArtifactRepresentation({
        artifactId: "video-1",
        digest: "sha256:abc",
        representationId: "playable",
        resourceId,
        revision: 3n,
        sessionId: "session-1",
      }),
    ).rejects.toThrow(`artifact_resource_unsupported:${resourceId}`);
    expect(globalThis.fetch).not.toHaveBeenCalled();
  });

  it("rejects incomplete artifact identity before fetching", async () => {
    await expect(
      loadSessionArtifactRepresentation({
        artifactId: "video-1",
        digest: "",
        representationId: "playable",
        resourceId: "session-file:file_12345678",
        revision: 3n,
        sessionId: "session-1",
      }),
    ).rejects.toThrow("artifact_identity_missing");
    expect(globalThis.fetch).not.toHaveBeenCalled();
  });

  it("reports an authorized artifact request failure", async () => {
    vi.mocked(globalThis.fetch).mockResolvedValueOnce(
      new Response(null, { status: 403 }),
    );

    await expect(
      loadSessionArtifactRepresentation({
        artifactId: "video-1",
        digest: "sha256:abc",
        representationId: "playable",
        resourceId: "session-file:file_12345678",
        revision: 3n,
        sessionId: "session-1",
      }),
    ).rejects.toThrow("Artifact request failed (403)");
  });
});
