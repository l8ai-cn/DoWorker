import { create, toBinary } from "@bufbuild/protobuf";
import { describe, expect, it, vi } from "vitest";

import {
  ArtifactDescriptorSchema,
  ArtifactRepresentationSchema,
} from "@proto/agent_workbench/v2/artifact_pb";
import { SessionSnapshotSchema } from "@proto/agent_workbench/v2/session_pb";
import { loadSessionWorkspaceArtifactById } from "@/lib/api/sessionWorkspaceArtifactApi";
import { createWebAgentWorkbenchArtifactLoader } from "./webAgentWorkbenchArtifactLoader";

vi.mock("@/lib/api/sessionWorkspaceArtifactApi", () => ({
  loadSessionWorkspaceArtifactById: vi.fn(async () =>
    new Blob(["video"], { type: "video/mp4" }),
  ),
}));

describe("createWebAgentWorkbenchArtifactLoader", () => {
  it("loads a workspace resource against the exact workbench session", async () => {
    const snapshot = create(SessionSnapshotSchema, {
      sessionId: "session-real-1",
      artifacts: [
        create(ArtifactDescriptorSchema, {
          artifactId: "video-1",
          representations: [
            create(ArtifactRepresentationSchema, {
              representationId: "playable",
              revision: 7n,
              mediaType: "video/mp4",
              transport: {
                transport: {
                  case: "resourceId",
                  value: "workspace:deliverables/demo.mp4",
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
      snapshotBytes: vi.fn(() =>
        toBinary(SessionSnapshotSchema, snapshot),
      ),
    });

    const blob = await loader({
      artifactId: "video-1",
      representationId: "playable",
      sessionId: "session-real-1",
    });

    expect(blob.type).toBe("video/mp4");
    expect(loadSessionWorkspaceArtifactById).toHaveBeenCalledWith(
      "session-real-1",
      "deliverables/demo.mp4",
      {
        artifactId: "video-1",
        representationId: "playable",
        revision: 7n,
      },
    );
  });
});
