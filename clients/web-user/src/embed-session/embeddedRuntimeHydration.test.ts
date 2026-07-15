import { describe, expect, it } from "vitest";

import type { AnyBlock } from "@/lib/blocks";
import {
  applyEmbeddedRuntimeHydration,
  loadEmbeddedRuntimeHydration,
  type EmbeddedRuntimeHydration,
} from "./embeddedRuntimeHydration";
import { createRuntimeState } from "./embeddedRuntimeState";

const session = {
  agentLabel: "codex-cli",
  id: "session-1",
  interactionMode: "acp" as const,
  podKey: "pod-1",
  status: "idle" as const,
  title: "Embedded task",
};

describe("applyEmbeddedRuntimeHydration", () => {
  it("keeps the conversation available when workspace artifact discovery fails", async () => {
    const hydration = await loadEmbeddedRuntimeHydration({
      getItems: async () => ({
        hasMore: false,
        items: [
          {
            id: "assistant-1",
            type: "message",
            response_id: "response-1",
            status: "completed",
            role: "assistant",
            content: [{ type: "output_text", text: "Conversation remains available." }],
          },
        ],
      }),
      getSession: async () => session,
      listWorkspaceArtifacts: async () => {
        throw new Error("runner unavailable");
      },
      openStream: async () => new Response(),
    });

    expect(hydration.page.items).toHaveLength(1);
    expect(hydration.session).toBe(session);
    expect(hydration.workspaceArtifacts).toEqual([]);
    expect(hydration.resourceError).toBe(
      "工作区成果暂不可用：runner unavailable",
    );
  });

  it("replaces an idless streamed assistant block with its durable item", () => {
    const context = {
      agent: null,
      depth: 0,
      itemId: null,
      responseId: "response-1",
      timestamp: 1,
      turn: 1,
    };
    const streamed = [
      {
        type: "text_chunk",
        ctx: context,
        text: "Embedded dialog ",
      },
      {
        type: "text_done",
        ctx: { ...context, itemId: "assistant-1" },
        fullText: "Embedded dialog interaction verified.",
        hasCodeBlocks: false,
      },
    ] satisfies AnyBlock[];
    const state = {
      ...createRuntimeState({
        getItems: async () => ({ hasMore: false, items: [] }),
        getSession: async () => session,
        openStream: async () => new Response(),
      }, "session-1"),
      blocks: streamed,
      session,
    };
    const hydration: EmbeddedRuntimeHydration = {
      page: {
        hasMore: false,
        items: [
          {
            id: "assistant-1",
            type: "message",
            response_id: "response-1",
            status: "completed",
            role: "assistant",
            content: [
              {
                type: "output_text",
                text: "Embedded dialog interaction verified.",
              },
            ],
          },
        ],
      },
      session,
      terminals: [],
      workspaceArtifacts: [],
      resourceError: null,
    };

    const hydrated = applyEmbeddedRuntimeHydration(state, hydration);
    const assistantBlocks = hydrated.blocks.filter(
      (block) => block.type === "text_done",
    );

    expect(assistantBlocks).toHaveLength(1);
    expect(assistantBlocks[0]?.ctx.itemId).toBe("assistant-1");
    expect(hydrated.blocks.some((block) => block.type === "text_chunk")).toBe(false);
  });
});
