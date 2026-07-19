import type { AgentArtifactItem } from "./agentArtifactContracts";
import type { AgentSessionSnapshot } from "./contracts";
import { userVideoExecutionSteps } from "./userVideoExecutionTrace";

describe("userVideoExecutionSteps", () => {
  it.each([
    ["queued", undefined, "pending", "queued"],
    ["rendering", 0.42, "running", "rendering"],
    ["transcoding", undefined, "completed", "generation_ready"],
  ] as const)(
    "maps %s artifact state to a truthful execution step",
    (stage, progress, expectedStatus, expectedDetail) => {
      const steps = userVideoExecutionSteps(
        processingSnapshot(),
        [videoArtifact(stage, progress)],
      );

      expect(step(steps, "generation")).toMatchObject({
        detail: expectedDetail,
        status: expectedStatus,
      });
      if (progress !== undefined) {
        expect(step(steps, "generation").progress).toBe(progress);
      }
    },
  );

  it("shows preview preparation while transcoding", () => {
    const steps = userVideoExecutionSteps(
      processingSnapshot(),
      [videoArtifact("transcoding")],
    );

    expect(step(steps, "preview")).toMatchObject({
      detail: "preparing_preview",
      status: "running",
    });
    expect(step(steps, "verification")).toMatchObject({ status: "pending" });
  });

  it("only marks publication complete for a verified ready file", () => {
    const snapshot = processingSnapshot();
    snapshot.status = "completed";
    const steps = userVideoExecutionSteps(snapshot, [verifiedVideoArtifact()]);

    expect(step(steps, "verification")).toMatchObject({
      detail: "published",
      status: "completed",
    });
  });

  it("fails the verification step when a ready result lacks trusted metadata", () => {
    const snapshot = processingSnapshot();
    snapshot.status = "completed";
    const invalid = verifiedVideoArtifact();
    invalid.representations[0]!.byteSize = BigInt(0);
    const steps = userVideoExecutionSteps(snapshot, [invalid]);

    expect(step(steps, "generation")).toMatchObject({ status: "completed" });
    expect(step(steps, "preview")).toMatchObject({ status: "completed" });
    expect(step(steps, "verification")).toMatchObject({
      detail: "verification_failed",
      status: "failed",
    });
  });

  it("does not use an older command artifact as the current execution trace", () => {
    const snapshot = processingSnapshot();
    snapshot.latestUserCommandId = "current-command";
    const artifact = verifiedVideoArtifact();
    artifact.provenance = { ...artifact.provenance, commandId: "older-command" };

    expect(userVideoExecutionSteps(snapshot, [artifact])).toEqual([]);
  });
});

function step(
  steps: ReturnType<typeof userVideoExecutionSteps>,
  id: "generation" | "preview" | "verification",
) {
  const value = steps.find((candidate) => candidate.id === id);
  if (!value) throw new Error(`missing ${id} step`);
  return value;
}

function processingSnapshot(): AgentSessionSnapshot {
  return {
    agentLabel: "Seedance Expert",
    capabilities: {
      interrupt: true,
      resolvePermission: false,
      sendMessage: true,
      terminal: false,
      updateConfiguration: false,
    },
    connection: "connected",
    error: null,
    hasOlderItems: false,
    interactionMode: "acp",
    items: [],
    permissions: [],
    plan: [],
    sessionId: "seedance-session",
    status: "running",
    terminals: [],
    title: "Seedance",
  };
}

function videoArtifact(
  stage: "queued" | "rendering" | "transcoding",
  progress?: number,
): AgentArtifactItem {
  return {
    ...verifiedVideoArtifact(),
    manifest: {
      derivativeRepresentationIds: [],
      kind: "video",
      ...(progress === undefined ? {} : { progressFraction: progress }),
      playableRepresentationId: "playable",
      stage,
      thumbnailRepresentationIds: [],
    },
    status: "processing",
  };
}

function verifiedVideoArtifact(): AgentArtifactItem {
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
      commandId: "current-command",
      publicationToolExecutionId: "workbench-publish:1",
    },
    representations: [
      {
        byteSize: 1_700_000n,
        digest: `sha256:${"b".repeat(64)}`,
        mediaType: "video/mp4",
        representationId: "playable",
        revision: 1n,
        status: "ready",
      },
    ],
    revision: 1n,
    role: "video",
    schemaVersion: "1",
    selectedRepresentationId: "playable",
    status: "completed",
  };
}
