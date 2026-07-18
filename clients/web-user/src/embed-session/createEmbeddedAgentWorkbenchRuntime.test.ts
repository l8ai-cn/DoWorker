import {
  AgentSessionConnection,
  AgentSessionRuntimeV2,
  AgentWorkbenchConnectTransport,
} from "@do-worker/agent-ui";
import { describe, expect, it, vi } from "vitest";

import { EmbeddedTerminalRuntime } from "./EmbeddedTerminalRuntime";
import { EmbeddedAgentSessionRuntime } from "./EmbeddedAgentSessionRuntime";
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

    expect(result.runtime).toBeInstanceOf(EmbeddedAgentSessionRuntime);
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

  it("runtime exposes session-bound attachment upload", async () => {
    const fetcher = vi.fn(async (input: RequestInfo | URL) => {
      if (String(input).endsWith("/resources/files")) {
        return new Response(
          JSON.stringify({
            id: "file_12345678",
            metadata: { bytes: 7 },
            name: "notes.txt",
          }),
          { status: 200 },
        );
      }
      return new Response(
        JSON.stringify({
          agent_name: "codex-cli",
          interaction_mode: "acp",
          title: "Repository review",
        }),
        { status: 200 },
      );
    });
    const result = await createEmbeddedAgentWorkbenchRuntime(
      {
        baseUrl: "https://api.example.test",
        getAccessToken: () => "session-token",
        orgSlug: "acme",
        sessionId: "session-1",
      },
      { fetch: fetcher },
    );
    const file = new File(["content"], "notes.txt", { type: "text/plain" });

    await expect(result.runtime.uploadAttachment("session-1", file)).resolves.toMatchObject({
      id: "file_12345678",
      name: "notes.txt",
    });
    expect(() => result.runtime.uploadAttachment("session-2", file)).toThrow(
      "agent_workbench_runtime_session_mismatch",
    );
  });
});
