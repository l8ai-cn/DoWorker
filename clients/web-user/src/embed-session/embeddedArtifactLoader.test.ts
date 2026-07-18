import type { ArtifactDescriptor } from "@do-worker/proto/agent_workbench/v2/artifact_pb";
import { describe, expect, it, vi } from "vitest";

import { createEmbeddedArtifactLoader } from "./embeddedArtifactLoader";

function connection(artifact: ArtifactDescriptor) {
  return {
    getStore: () => ({
      getState: () => ({
        snapshot: { artifacts: [artifact], sessionId: "session-1" },
      }),
    }),
  };
}

function descriptor(
  transport:
    | { case: "downloadUrl"; value: string }
    | { case: "inlineBytes"; value: Uint8Array }
    | { case: "inlineText"; value: string }
    | { case: "resourceId"; value: string }
    | undefined,
): ArtifactDescriptor {
  return {
    artifactId: "artifact-1",
    filename: "result.bin",
    mediaType: "application/octet-stream",
    representations: [
      {
        representationId: "source",
        revision: 1n,
        mediaType: "text/plain",
        status: 3,
        transport: transport
          ? { $typeName: "proto.agent_workbench.v2.ArtifactTransport", transport }
          : undefined,
        $typeName: "proto.agent_workbench.v2.ArtifactRepresentation",
      },
    ],
    revision: 1n,
    status: 3,
    revisions: [],
    grants: [
      {
        actions: ["artifact.download"],
        grantId: "grant-download",
        representationIds: ["source"],
        $typeName: "proto.agent_workbench.v2.ArtifactGrant",
      },
    ],
    $typeName: "proto.agent_workbench.v2.ArtifactDescriptor",
  };
}

describe("createEmbeddedArtifactLoader", () => {
  it("直接把 inline text 和 bytes 转成精确媒体类型的 Blob", async () => {
    const resources = {
      loadDownload: vi.fn(),
      loadResource: vi.fn(),
    };
    const textLoader = createEmbeddedArtifactLoader(
      connection(descriptor({ case: "inlineText", value: "hello" })),
      resources,
    );
    const bytesLoader = createEmbeddedArtifactLoader(
      connection(
        descriptor({
          case: "inlineBytes",
          value: new TextEncoder().encode("bytes"),
        }),
      ),
      resources,
    );

    const text = await textLoader({
      artifactId: "artifact-1",
      representationId: "source",
      sessionId: "session-1",
    });
    const bytes = await bytesLoader({
      artifactId: "artifact-1",
      representationId: "source",
      sessionId: "session-1",
    });

    expect(text.type).toBe("text/plain");
    await expect(text.text()).resolves.toBe("hello");
    expect(bytes.type).toBe("text/plain");
    await expect(bytes.text()).resolves.toBe("bytes");
    expect(resources.loadDownload).not.toHaveBeenCalled();
    expect(resources.loadResource).not.toHaveBeenCalled();
  });

  it("按 raw representation transport 精确分派 resourceId 和 downloadUrl", async () => {
    const resourceBlob = new Blob(["resource"]);
    const downloadBlob = new Blob(["download"]);
    const resources = {
      loadDownload: vi.fn().mockResolvedValue(downloadBlob),
      loadResource: vi.fn().mockResolvedValue(resourceBlob),
    };
    const resourceLoader = createEmbeddedArtifactLoader(
      connection(descriptor({ case: "resourceId", value: "session-file:file_a" })),
      resources,
    );
    const downloadLoader = createEmbeddedArtifactLoader(
      connection(descriptor({ case: "downloadUrl", value: "/downloads/a.png" })),
      resources,
    );

    await expect(
      resourceLoader({
        artifactId: "artifact-1",
        representationId: "source",
        sessionId: "session-1",
      }),
    ).resolves.toBe(resourceBlob);
    await expect(
      downloadLoader({
        artifactId: "artifact-1",
        representationId: "source",
        sessionId: "session-1",
      }),
    ).resolves.toBe(downloadBlob);
    expect(resources.loadResource).toHaveBeenCalledWith(
      "session-file:file_a",
      expect.objectContaining({
        artifactId: "artifact-1",
        representationId: "source",
        sessionId: "session-1",
      }),
    );
    expect(resources.loadDownload).toHaveBeenCalledWith("/downloads/a.png");
  });

  it("representation 或 transport 缺失时硬错误", async () => {
    const resources = {
      loadDownload: vi.fn(),
      loadResource: vi.fn(),
    };
    const loader = createEmbeddedArtifactLoader(connection(descriptor(undefined)), resources);

    await expect(
      loader({
        artifactId: "artifact-1",
        sessionId: "session-1",
      }),
    ).rejects.toThrow("artifact_representation_missing");
    await expect(
      loader({
        artifactId: "artifact-1",
        representationId: "source",
        sessionId: "session-1",
      }),
    ).rejects.toThrow("artifact_transport_missing");
  });
});
