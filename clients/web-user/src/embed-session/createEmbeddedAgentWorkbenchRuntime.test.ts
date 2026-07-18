import {
  AgentSessionConnection,
  AgentSessionRuntimeV2,
  AgentWorkbenchConnectTransport,
} from "@do-worker/agent-ui";
import { describe, expect, it, vi } from "vitest";

import { EmbeddedTerminalRuntime } from "./EmbeddedTerminalRuntime";
import { createEmbeddedAgentWorkbenchRuntime } from "./createEmbeddedAgentWorkbenchRuntime";

describe("createEmbeddedAgentWorkbenchRuntime", () => {
  it("用同一 access scope 组装官方 Connect V2 runtime", async () => {
    const getAccessToken = vi.fn(() => "session-token");
    const fetcher = vi.fn(async (input: RequestInfo | URL) => {
      expect(String(input)).toBe("https://api.example.test/v1/embed/sessions/session-1");
      return new Response(
        JSON.stringify({
          agent_name: "codex-cli",
          interaction_mode: "acp",
          title: "Repository review",
        }),
        { status: 200, headers: { "Content-Type": "application/json" } },
      );
    });

    const result = await createEmbeddedAgentWorkbenchRuntime(
      {
        baseUrl: "https://api.example.test",
        getAccessToken,
        orgSlug: "acme",
        sessionId: "session-1",
      },
      { fetch: fetcher },
    );

    expect(result.runtime).toBeInstanceOf(AgentSessionRuntimeV2);
    expect(result.terminalRuntime).toBeInstanceOf(EmbeddedTerminalRuntime);
    const connection = (result.runtime as unknown as { connection: AgentSessionConnection })
      .connection;
    expect(connection).toBeInstanceOf(AgentSessionConnection);
    const transport = (connection as unknown as { transport: AgentWorkbenchConnectTransport })
      .transport;
    expect(transport).toBeInstanceOf(AgentWorkbenchConnectTransport);
    expect(transport as unknown as { orgSlug: string; sessionId: string }).toMatchObject({
      orgSlug: "acme",
      sessionId: "session-1",
    });
    expect(result.runtime.getSnapshot("session-1")).toMatchObject({
      agentLabel: "codex-cli",
      interactionMode: "acp",
      sessionId: "session-1",
      title: "Repository review",
    });
    expect(getAccessToken).toHaveBeenCalledTimes(1);
  });
});
