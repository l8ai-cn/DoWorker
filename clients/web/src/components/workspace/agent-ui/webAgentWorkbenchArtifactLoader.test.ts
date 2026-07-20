import { create, toBinary } from "@bufbuild/protobuf";
import { describe, expect, it, vi } from "vitest";

import {
  ArtifactDescriptorSchema,
  ArtifactGrantSchema,
  ArtifactRepresentationSchema,
} from "@proto/agent_workbench/v2/artifact_pb";
import { SessionSnapshotSchema } from "@proto/agent_workbench/v2/session_pb";
import { loadSessionArtifactRepresentation } from "@/lib/api/sessionWorkspaceArtifactApi";
import { createWebAgentWorkbenchArtifactLoader } from "./webAgentWorkbenchArtifactLoader";

vi.mock("@/lib/api/podWorkspaceArtifactApi", () => ({
  loadPodWorkspaceArtifact: vi.fn(async () => new Blob(["workspace-video"], {
    type: "video/mp4",
  })),
}));

vi.mock("@/lib/api/sessionWorkspaceArtifactApi", () => ({
  loadSessionArtifactRepresentation: vi.fn(async (input: { resourceId: string }) => {
    if (!input.resourceId.startsWith("session-file:")) {
      throw new Error(`artifact_resource_unsupported:${input.resourceId}`);
    }
    return new Blob(["video"], { type: "video/mp4" });
  }),
}));

import { loadPodWorkspaceArtifact } from "@/lib/api/podWorkspaceArtifactApi";

describe("createWebAgentWorkbenchArtifactLoader", () => {
  it("loads an immutable session file against the exact artifact identity", async () => {
    const snapshot = create(SessionSnapshotSchema, {
      sessionId: "session-real-1",
      artifacts: [
        create(ArtifactDescriptorSchema, {
          artifactId: "video-1",
          revision: 3n,
          grants: [
            create(ArtifactGrantSchema, {
              actions: ["artifact.download"],
              grantId: "grant-download",
              representationIds: ["playable"],
            }),
          ],
          representations: [
            create(ArtifactRepresentationSchema, {
              representationId: "playable",
              revision: 3n,
              mediaType: "video/mp4",
              digest: "sha256:video",
              transport: {
                transport: {
                  case: "resourceId",
                  value: "session-file:file_12345678",
                },
              },
            }),
          ],
        }),
      ],
    });
    const loader = createWebAgentWorkbenchArtifactLoader({
      projectionStatus: vi.fn(),
      resyncReason: vi.fn(),
      revision: vi.fn(),
      snapshotBytes: vi.fn(() => toBinary(SessionSnapshotSchema, snapshot)),
    });

    const blob = await loader({
      artifactId: "video-1",
      representationId: "playable",
      sessionId: "session-real-1",
    });

    expect(blob.type).toBe("video/mp4");
    expect(loadSessionArtifactRepresentation).toHaveBeenCalledWith({
      artifactId: "video-1",
      digest: "sha256:video",
      representationId: "playable",
      resourceId: "session-file:file_12345678",
      revision: 3n,
      sessionId: "session-real-1",
    });
  });

  it("rejects non-durable resource transports", async () => {
    const snapshot = create(SessionSnapshotSchema, {
      sessionId: "session-real-1",
      artifacts: [
        create(ArtifactDescriptorSchema, {
          artifactId: "preview-1",
          revision: 1n,
          grants: [
            create(ArtifactGrantSchema, {
              actions: ["artifact.download"],
              grantId: "grant-download",
              representationIds: ["pdf"],
            }),
          ],
          representations: [
            create(ArtifactRepresentationSchema, {
              representationId: "pdf",
              revision: 1n,
              mediaType: "application/pdf",
              digest: "sha256:pdf",
              transport: {
                transport: {
                  case: "resourceId",
                  value: "workspace:preview.pdf",
                },
              },
            }),
          ],
        }),
      ],
    });
    const loader = createWebAgentWorkbenchArtifactLoader({
      projectionStatus: vi.fn(),
      resyncReason: vi.fn(),
      revision: vi.fn(),
      snapshotBytes: vi.fn(() => toBinary(SessionSnapshotSchema, snapshot)),
    });

    await expect(loader({
      artifactId: "preview-1",
      representationId: "pdf",
      sessionId: "session-real-1",
    })).rejects.toThrow("artifact_resource_unsupported:workspace:preview.pdf");
  });

  it("loads a workspace video through the Pod filesystem resource", async () => {
    const loader = createWebAgentWorkbenchArtifactLoader({
      projectionStatus: vi.fn(),
      resyncReason: vi.fn(),
      revision: vi.fn(),
      snapshotBytes: vi.fn(),
    }, { podKey: "pod-video-1" });

    const blob = await loader({
      artifactId: "workspace:output/final.mp4",
      representationId: "workspace-file",
      sessionId: "session-real-1",
    });

    expect(blob.type).toBe("video/mp4");
    expect(loadPodWorkspaceArtifact).toHaveBeenCalledWith(
      "pod-video-1",
      "output/final.mp4",
    );
  });

  it("rejects an unregistered workspace representation", async () => {
    const loader = createWebAgentWorkbenchArtifactLoader({
      projectionStatus: vi.fn(),
      resyncReason: vi.fn(),
      revision: vi.fn(),
      snapshotBytes: vi.fn(),
    }, { podKey: "pod-video-1" });

    await expect(loader({
      artifactId: "workspace:output/final.mp4",
      representationId: "original",
      sessionId: "session-real-1",
    })).rejects.toThrow("workspace_artifact_representation_invalid");
  });
});
