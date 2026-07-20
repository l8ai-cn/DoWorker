import { describe, expect, it } from "vitest";

import {
  workspaceFileArtifacts,
} from "@do-worker/agent-ui";
import {
  mergeWorkspaceArtifactDiscoveryError,
  prepareWorkspaceArtifacts,
  workspaceArtifactPath,
} from "./webAgentWorkbenchWorkspaceArtifacts";

const snapshot = {
  agentLabel: "Video Studio",
  capabilities: {
    interrupt: false,
    resolvePermission: false,
    sendMessage: false,
    terminal: false,
    updateConfiguration: false,
  },
  connection: "connected",
  error: null,
  hasOlderItems: false,
  interactionMode: "pty",
  items: [],
  permissions: [],
  plan: [],
  sessionId: "session-1",
  status: "completed",
  terminals: [],
  title: "Video",
};

describe("workspace artifact projection", () => {
  it("merges a playable workspace MP4 into the workbench", () => {
    const artifacts = workspaceFileArtifacts("workspace", [{
      path: "final.mp4",
      status: "created",
    }]);
    const [artifact] = artifacts;
    if (!artifact) throw new Error("workspace artifact fixture missing");
    const [playable] = prepareWorkspaceArtifacts([artifact]);
    if (!playable) throw new Error("playable artifact fixture missing");

    expect(playable.selectedRepresentationId).toBe("workspace-file");
    expect(playable.grants[0]?.actions).toContain("artifact.download");
  });

  it("keeps a discovery error visible to the developer timeline", () => {
    const projected = mergeWorkspaceArtifactDiscoveryError(
      snapshot,
      "workspace_artifact_request_failed",
    );

    expect(projected.items).toContainEqual({
      detail: "workspace_artifact_request_failed",
      id: "workspace-artifact-discovery",
      kind: "system",
      status: "failed",
      title: "Workspace artifact discovery failed",
    });
  });

  it("accepts only non-traversing workspace artifact paths", () => {
    expect(workspaceArtifactPath("workspace:output/final.mp4")).toBe(
      "output/final.mp4",
    );
    expect(workspaceArtifactPath("workspace:../secret.mp4")).toBeNull();
    expect(workspaceArtifactPath("artifact:video-1")).toBeNull();
  });
});
