import { describe, expect, it } from "vitest";

import type { AgentTimelineItem } from "./contracts";
import { groupToolActivity } from "./toolActivityGrouping";

describe("groupToolActivity", () => {
  it("groups only contiguous tool activity", () => {
    const items: AgentTimelineItem[] = [
      tool("tool-1"),
      tool("tool-2"),
      {
        id: "assistant-1",
        kind: "message",
        role: "assistant",
        text: "Done",
        status: "completed",
      },
      tool("tool-3"),
    ];

    const grouped = groupToolActivity(items);

    expect(grouped).toHaveLength(3);
    expect(grouped[0]).toMatchObject({
      kind: "tool-run",
      tools: [{ id: "tool-1" }, { id: "tool-2" }],
    });
    expect(grouped[1]).toMatchObject({ id: "assistant-1", kind: "message" });
    expect(grouped[2]).toMatchObject({
      kind: "tool-run",
      tools: [{ id: "tool-3" }],
    });
  });

  it("uses every non-tool event kind as a group boundary", () => {
    const boundaries: AgentTimelineItem[] = [
      {
        id: "message",
        kind: "message",
        role: "assistant",
        text: "Done",
        status: "completed",
      },
      {
        actions: [],
        id: "artifact",
        kind: "artifact",
        artifactId: "artifact-1",
        filename: "result.png",
        grants: [],
        manifest: null,
        mimeType: "image/png",
        representations: [],
        revision: 1n,
        role: "preview",
        schemaVersion: "1",
        selectedRepresentationId: null,
        status: "completed",
      },
      {
        id: "reasoning",
        kind: "reasoning",
        title: "Considering",
        status: "completed",
      },
      {
        id: "error",
        kind: "error",
        title: "Failed",
        status: "failed",
      },
      {
        id: "system",
        kind: "system",
        title: "Connected",
        status: "completed",
      },
    ];
    const items = boundaries.flatMap((boundary, index) => [
      tool(`tool-${index}`),
      boundary,
    ]);
    items.push(tool("tool-tail"));

    const grouped = groupToolActivity(items);

    expect(grouped.filter((item) => item.kind === "tool-run")).toHaveLength(6);
    for (const item of grouped) {
      if (item.kind === "tool-run") expect(item.tools).toHaveLength(1);
    }
  });
});

function tool(id: string): AgentTimelineItem {
  return {
    id,
    identity: {
      namespace: "agentsmesh.acp",
      schemaVersion: "1",
      semanticKey: "shell",
    },
    kind: "tool",
    results: [],
    title: "shell",
    status: "completed",
  };
}
