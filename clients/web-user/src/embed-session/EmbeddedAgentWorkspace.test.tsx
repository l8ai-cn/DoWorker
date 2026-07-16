import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import type { EmbedSessionClient } from "@/embed-session-api";
import { EmbeddedAgentWorkspace } from "./EmbeddedAgentWorkspace";

function client(): EmbedSessionClient {
  return {
    getItems: vi.fn().mockResolvedValue({
      hasMore: false,
      items: [
        {
          id: "assistant-1",
          type: "message",
          response_id: "response-1",
          status: "completed",
          role: "assistant",
          content: [{ type: "output_text", text: "Workspace runtime is connected." }],
        },
      ],
    }),
    getSession: vi.fn().mockResolvedValue({
      agentLabel: "codex-cli",
      id: "session-1",
      interactionMode: "acp",
      podKey: "pod-1",
      status: "idle",
      title: "Repository review",
    }),
    openStream: vi.fn((signal: AbortSignal) =>
      Promise.resolve(
        new Response(
          new ReadableStream<Uint8Array>({
            start(controller) {
              signal.addEventListener("abort", () => controller.close(), { once: true });
            },
          }),
          { status: 200 },
        ),
      ),
    ),
    sendMessage: vi.fn().mockResolvedValue({ itemId: "user-1" }),
  };
}

describe("EmbeddedAgentWorkspace", () => {
  it("renders the shared agent workspace instead of the legacy timeline", async () => {
    render(<EmbeddedAgentWorkspace client={client()} sessionId="session-1" />);

    expect(await screen.findByText("Repository review")).toBeInTheDocument();
    expect(screen.getByText("Workspace runtime is connected.")).toBeInTheDocument();
    expect(screen.getByRole("tab", { name: "对话" })).toBeInTheDocument();
    expect(screen.getByLabelText("给智能体发送消息")).toBeEnabled();
    expect(
      screen.getByText("Repository review").closest("[data-agent-workspace]")
        ?.parentElement,
    ).toHaveClass("h-full", "min-h-0", "overflow-hidden");
  });
});
