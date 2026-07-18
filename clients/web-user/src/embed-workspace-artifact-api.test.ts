import { describe, expect, it, vi } from "vitest";

import { loadEmbeddedArtifactRepresentation } from "./embed-workspace-artifact-api";

describe("embed artifact API", () => {
  it("loads an exact immutable artifact identity", async () => {
    const request = vi.fn(async () => new Response("artifact"));

    await loadEmbeddedArtifactRepresentation(
      request,
      "/v1/embed/sessions/session-1",
      {
        artifactId: "deck-1",
        digest: "sha256:abc",
        representationId: "preview-pdf",
        resourceId: "session-file:file_12345678",
        revision: 4n,
      },
    );

    expect(request).toHaveBeenCalledWith(
      "/v1/embed/sessions/session-1/artifacts/content?" +
        "artifact_id=deck-1&digest=sha256%3Aabc&" +
        "representation_id=preview-pdf&revision=4",
    );
  });

  it.each([
    "",
    "workspace:preview.pdf",
    "artifact-cache:preview-1",
    "session-file:",
  ])("rejects non-durable resource IDs: %j", async (resourceId) => {
    const request = vi.fn();

    await expect(
      loadEmbeddedArtifactRepresentation(
        request,
        "/v1/embed/sessions/session-1",
        {
          artifactId: "deck-1",
          digest: "sha256:abc",
          representationId: "preview-pdf",
          resourceId,
          revision: 4n,
        },
      ),
    ).rejects.toThrow(`artifact_resource_unsupported:${resourceId}`);
    expect(request).not.toHaveBeenCalled();
  });
});
