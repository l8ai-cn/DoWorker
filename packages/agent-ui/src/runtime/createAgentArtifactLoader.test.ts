import { create } from "@bufbuild/protobuf";
import { describe, expect, it, vi } from "vitest";

import {
  ArtifactDescriptorSchema,
  ArtifactRepresentationSchema,
} from "@do-worker/proto/agent_workbench/v2/artifact_pb";
import { createAgentArtifactLoader } from "./createAgentArtifactLoader";

function descriptor(
  transport:
    | { case: "inlineBytes"; value: Uint8Array }
    | { case: "inlineText"; value: string }
    | { case: "resourceId"; value: string }
    | { case: "downloadUrl"; value: string },
) {
  return create(ArtifactDescriptorSchema, {
    artifactId: "artifact-1",
    representations: [
      create(ArtifactRepresentationSchema, {
        mediaType: "video/mp4",
        representationId: "representation-1",
        transport: { transport },
      }),
    ],
  });
}

describe("createAgentArtifactLoader", () => {
  it("dispatches the exact protobuf transport without inferring from filenames", async () => {
    const loadResource = vi.fn(async () => new Blob(["resource"]));
    const loadDownload = vi.fn(async () => new Blob(["download"]));
    const loader = createAgentArtifactLoader({
      getArtifacts: () => [descriptor({
        case: "resourceId",
        value: "workspace:deliverables/demo.mp4",
      })],
      loadDownload,
      loadResource,
    });

    await loader({
      artifactId: "artifact-1",
      representationId: "representation-1",
      sessionId: "session-1",
    });

    expect(loadResource).toHaveBeenCalledWith(
      "workspace:deliverables/demo.mp4",
      expect.objectContaining({
        representationId: "representation-1",
        sessionId: "session-1",
      }),
    );
    expect(loadDownload).not.toHaveBeenCalled();
  });

  it("hard-fails when the requested representation or transport is missing", async () => {
    const loader = createAgentArtifactLoader({
      getArtifacts: () => [descriptor({
        case: "inlineText",
        value: "<html></html>",
      })],
      loadDownload: vi.fn(),
      loadResource: vi.fn(),
    });

    await expect(loader({
      artifactId: "artifact-1",
      sessionId: "session-1",
    })).rejects.toThrow("artifact_representation_missing");
    await expect(loader({
      artifactId: "artifact-1",
      representationId: "missing",
      sessionId: "session-1",
    })).rejects.toThrow("artifact_representation_missing");
  });
});
