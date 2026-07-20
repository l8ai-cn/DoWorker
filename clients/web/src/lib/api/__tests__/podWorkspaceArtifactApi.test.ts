import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("@/lib/env", () => ({
  getApiBaseUrl: () => "",
}));

vi.mock("@/lib/wasm-core", () => ({
  getAuthManager: () => ({ get_token: () => "test-token" }),
}));

vi.mock("@/stores/auth", () => ({
  readCurrentOrg: () => ({ slug: "dev-org" }),
}));

import {
  listPodWorkspaceArtifacts,
  loadPodWorkspaceArtifact,
} from "../podWorkspaceArtifactApi";

describe("podWorkspaceArtifactApi", () => {
  const originalFetch = globalThis.fetch;

  beforeEach(() => {
    globalThis.fetch = vi.fn();
  });

  afterEach(() => {
    globalThis.fetch = originalFetch;
  });

  it("projects supported Worker files into workspace artifacts", async () => {
    vi.mocked(globalThis.fetch).mockResolvedValueOnce(
      new Response(
        JSON.stringify({
          object: "list",
          data: [
            { path: "output/demo.mp4", status: "created" },
            { path: "src/main.ts", status: "created" },
          ],
          has_more: false,
        }),
        { status: 200, headers: { "Content-Type": "application/json" } },
      ),
    );

    await expect(listPodWorkspaceArtifacts("worker-1")).resolves.toEqual([
      {
        id: "workspace-discovery:artifact:0",
          kind: "artifact",
          artifactId: "workspace:output/demo.mp4",
          filename: "demo.mp4",
          actions: [],
          grants: [],
          manifest: null,
          mimeType: "video/mp4",
          representations: [],
          revision: BigInt(0),
          role: "preview",
          schemaVersion: "1",
          selectedRepresentationId: null,
          status: "completed",
        },
    ]);
    expect(globalThis.fetch).toHaveBeenCalledWith(
      "/api/v1/orgs/dev-org/pods/worker-1/resources/workspace/changes",
      {
        cache: "no-store",
        headers: { Authorization: "Bearer test-token" },
      },
    );
  });

  it("loads a binary Worker artifact without resolving a session", async () => {
    vi.mocked(globalThis.fetch).mockResolvedValueOnce(
      new Response(
        JSON.stringify({
          content: "AAAA",
          content_type: "video/mp4",
          encoding: "base64",
          truncated: false,
        }),
        { status: 200, headers: { "Content-Type": "application/json" } },
      ),
    );

    const blob = await loadPodWorkspaceArtifact(
      "worker-1",
      "output/demo clip.mp4",
    );

    expect(blob.type).toBe("video/mp4");
    expect(blob.size).toBe(3);
    expect(globalThis.fetch).toHaveBeenCalledWith(
      "/api/v1/orgs/dev-org/pods/worker-1/resources/workspace/filesystem/output/demo%20clip.mp4",
      {
        cache: "no-store",
        headers: { Authorization: "Bearer test-token" },
      },
    );
  });

  it("rejects truncated Worker artifact content", async () => {
    vi.mocked(globalThis.fetch).mockResolvedValueOnce(
      new Response(
        JSON.stringify({
          content: "AAAA",
          content_type: "video/mp4",
          encoding: "base64",
          truncated: true,
        }),
        { status: 200, headers: { "Content-Type": "application/json" } },
      ),
    );

    await expect(
      loadPodWorkspaceArtifact("worker-1", "output/demo.mp4"),
    ).rejects.toThrow("exceeds the preview size limit");
  });
});
