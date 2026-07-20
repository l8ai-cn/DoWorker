import { describe, expect, it } from "vitest";

import type { AgentArtifactItem } from "../agentArtifactContracts";
import { collectWorkbenchResults } from "./workbenchResults";

describe("collectWorkbenchResults", () => {
  it("includes authorized workspace files without promoting them to verified videos", () => {
    const results = collectWorkbenchResults(
      [],
      [],
      undefined,
      true,
      [workspaceVideo()],
    );

    expect(results).toHaveLength(1);
    expect(results[0]).toMatchObject({
      id: "artifact:workspace-video",
      kind: "artifact",
    });
  });
});

function workspaceVideo(): AgentArtifactItem {
  return {
    actions: [],
    artifactId: "workspace:output/final.mp4",
    filename: "final.mp4",
    grants: [],
    id: "workspace-video",
    kind: "artifact",
    manifest: null,
    mimeType: "video/mp4",
    representations: [],
    revision: 0n,
    role: "preview",
    schemaVersion: "1",
    selectedRepresentationId: null,
    status: "completed",
  };
}
