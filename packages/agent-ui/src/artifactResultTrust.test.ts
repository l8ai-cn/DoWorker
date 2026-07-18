import type { AgentArtifactItem } from "./agentArtifactContracts";
import {
  isUserVisibleArtifact,
  isVerifiedReadyVideoArtifact,
} from "./artifactResultTrust";

const digest = `sha256:${"a".repeat(64)}`;

describe("video artifact trust", () => {
  it("accepts only a ready, non-empty, digest-backed playable video", () => {
    const artifact = videoArtifact();

    expect(isVerifiedReadyVideoArtifact(artifact)).toBe(true);
    expect(isUserVisibleArtifact(artifact)).toBe(true);
  });

  it.each([
    ["missing manifest", { manifest: null }],
    ["zero bytes", { representations: [playable({ byteSize: 0n })] }],
    ["missing digest", { representations: [playable({ digest: undefined })] }],
    ["failed representation", {
      representations: [playable({ status: "failed" })],
    }],
    ["missing Runner publication execution", {
      provenance: {
        producerId: "seedance-task-1",
        producerNamespace: "seedance",
        producerType: "video.generate",
      },
    }],
  ])("rejects completed video with %s", (_label, override) => {
    const artifact = { ...videoArtifact(), ...override };

    expect(isVerifiedReadyVideoArtifact(artifact)).toBe(false);
    expect(isUserVisibleArtifact(artifact)).toBe(false);
  });

  it("keeps an in-progress video visible without claiming completion", () => {
    const artifact = {
      ...videoArtifact(),
      status: "processing" as const,
      manifest: {
        ...videoArtifact().manifest!,
        stage: "rendering" as const,
      },
    };

    expect(isVerifiedReadyVideoArtifact(artifact)).toBe(false);
    expect(isUserVisibleArtifact(artifact)).toBe(true);
  });

  it("hides a failed video whose manifest still claims it is ready", () => {
    const artifact = {
      ...videoArtifact(),
      status: "failed" as const,
    };

    expect(isUserVisibleArtifact(artifact)).toBe(false);
  });

  it("hides non-video artifacts in user presentation", () => {
    const artifact = {
      ...videoArtifact(),
      manifest: null,
      mimeType: "text/csv",
      role: "data",
    };

    expect(isUserVisibleArtifact(artifact)).toBe(false);
  });
});

function videoArtifact(): AgentArtifactItem {
  return {
    actions: [],
    artifactId: "seedance-video",
    filename: "seedance-video.mp4",
    grants: [],
    id: "artifact-video",
    kind: "artifact",
    manifest: {
      derivativeRepresentationIds: [],
      kind: "video",
      playableRepresentationId: "playable",
      stage: "ready",
      thumbnailRepresentationIds: [],
    },
    mimeType: "video/mp4",
    provenance: {
      publicationToolExecutionId: "workbench-publish:1",
      producerId: "seedance-task-1",
      producerNamespace: "seedance",
      producerType: "video.generate",
    },
    representations: [playable()],
    revision: 1n,
    role: "video",
    schemaVersion: "1",
    selectedRepresentationId: "playable",
    status: "completed",
  };
}

function playable(
  override: Partial<AgentArtifactItem["representations"][number]> = {},
): AgentArtifactItem["representations"][number] {
  return {
    byteSize: 1024n,
    digest,
    mediaType: "video/mp4",
    representationId: "playable",
    revision: 1n,
    role: "playable",
    status: "ready",
    ...override,
  };
}
