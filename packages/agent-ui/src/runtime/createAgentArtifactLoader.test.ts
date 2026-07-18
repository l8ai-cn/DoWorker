import { create } from "@bufbuild/protobuf";
import { describe, expect, it, vi } from "vitest";

import {
  ArtifactDescriptorSchema,
  ArtifactGrantSchema,
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
    grants: [
      create(ArtifactGrantSchema, {
        actions: ["artifact.download"],
        grantId: "grant-download",
        representationIds: ["representation-1"],
      }),
    ],
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

  it("rejects transport reads without an artifact download grant", async () => {
    const artifact = descriptor({
      case: "resourceId",
      value: "session-file:file_12345678",
    });
    artifact.grants = [];
    const loadResource = vi.fn();
    const loader = createAgentArtifactLoader({
      getArtifacts: () => [artifact],
      loadDownload: vi.fn(),
      loadResource,
    });

    await expect(loader({
      artifactId: "artifact-1",
      representationId: "representation-1",
      sessionId: "session-1",
    })).rejects.toThrow("artifact_download_not_authorized");
    expect(loadResource).not.toHaveBeenCalled();
  });

  it("rejects invalid or expired artifact grants", async () => {
    const artifact = descriptor({
      case: "resourceId",
      value: "session-file:file_12345678",
    });
    const loadResource = vi.fn();
    const loader = createAgentArtifactLoader({
      getArtifacts: () => [artifact],
      loadDownload: vi.fn(),
      loadResource,
    });

    artifact.grants[0].expiresAt = "invalid";
    await expect(loader({
      artifactId: "artifact-1",
      representationId: "representation-1",
      sessionId: "session-1",
    })).rejects.toThrow("artifact_download_not_authorized");

    artifact.grants[0].expiresAt = "2026-07-18T00:00:00Z";
    await expect(loader({
      artifactId: "artifact-1",
      representationId: "representation-1",
      sessionId: "session-1",
    })).rejects.toThrow("artifact_download_not_authorized");
    expect(loadResource).not.toHaveBeenCalled();
  });
});
