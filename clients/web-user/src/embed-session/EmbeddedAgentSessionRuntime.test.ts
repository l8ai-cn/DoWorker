import { describe, expect, it, vi } from "vitest";

import type { EmbedSessionClient } from "@/embed-session-api";
import { EmbeddedAgentSessionRuntime } from "./EmbeddedAgentSessionRuntime";

function openStream(signal: AbortSignal): Promise<Response> {
  return Promise.resolve(
    new Response(
      new ReadableStream<Uint8Array>({
        start(controller) {
          signal.addEventListener("abort", () => controller.close(), { once: true });
        },
      }),
      { status: 200 },
    ),
  );
}

function createClient(): EmbedSessionClient {
  const artifact = new Blob(["artifact"], { type: "text/plain" });
  return {
    getItems: vi.fn().mockResolvedValue({
      hasMore: true,
      items: [
        {
          id: "user-1",
          type: "message",
          response_id: "response-1",
          status: "completed",
          role: "user",
          content: [{ type: "input_text", text: "Review auth" }],
        },
        {
          id: "assistant-1",
          type: "message",
          response_id: "response-1",
          status: "completed",
          role: "assistant",
          content: [{ type: "output_text", text: "I found duplicated session logic." }],
        },
      ],
    }),
    getSession: vi.fn().mockResolvedValue({
      agentLabel: "codex-cli",
      id: "session-1",
      interactionMode: "pty",
      podKey: "pod-1",
      status: "idle",
      title: "Auth review",
    }),
    getTerminals: vi.fn().mockResolvedValue([
      {
        id: "terminal-main",
        label: "main",
        status: "connected",
        writable: true,
      },
    ]),
    interrupt: vi.fn().mockResolvedValue(undefined),
    loadArtifact: vi.fn().mockResolvedValue(artifact),
    listWorkspaceArtifacts: vi.fn().mockResolvedValue([
      {
        id: "workspace-discovery:artifact:0",
        kind: "artifact",
        artifactId: "workspace:deliverables/demo.mp4",
        filename: "demo.mp4",
        mimeType: "video/mp4",
        status: "completed",
      },
    ]),
    openStream: vi.fn(openStream),
    resolvePermission: vi.fn().mockResolvedValue(undefined),
    sendMessage: vi.fn().mockResolvedValue({ itemId: "user-2" }),
  };
}

describe("EmbeddedAgentSessionRuntime", () => {
  it("hydrates the shared workspace and delegates scoped commands", async () => {
    const client = createClient();
    const runtime = new EmbeddedAgentSessionRuntime(client);
    const listener = vi.fn();
    runtime.subscribe("session-1", listener);

    expect(runtime.getSnapshot("session-1")).toBe(
      runtime.getSnapshot("session-1"),
    );
    await runtime.open("session-1");

    expect(runtime.getSnapshot("session-1")).toMatchObject({
      sessionId: "session-1",
      title: "Auth review",
      connection: "connected",
      capabilities: {
        interrupt: true,
        resolvePermission: true,
        sendMessage: true,
        terminal: true,
        updateConfiguration: false,
      },
      hasOlderItems: true,
      terminals: [{ id: "terminal-main" }],
    });
    expect(runtime.getSnapshot("session-1").items).toEqual(
      expect.arrayContaining([
        expect.objectContaining({ kind: "message", role: "user", text: "Review auth" }),
        expect.objectContaining({
          kind: "message",
          role: "assistant",
          text: "I found duplicated session logic.",
        }),
        expect.objectContaining({
          kind: "artifact",
          artifactId: "workspace:deliverables/demo.mp4",
        }),
      ]),
    );

    await runtime.sendMessage("session-1", "command-1", { text: "Fix it" });
    await runtime.interrupt("session-1", "command-2");
    await runtime.resolvePermission(
      "session-1",
      "command-3",
      "permission-1",
      { action: "accept", content: { answers: {} } },
    );
    const artifact = await runtime.loadArtifact("session-1", "file-1");
    await expect(artifact.text()).resolves.toBe("artifact");

    expect(client.sendMessage).toHaveBeenCalledWith("Fix it");
    expect(client.interrupt).toHaveBeenCalledOnce();
    expect(client.resolvePermission).toHaveBeenCalledWith("permission-1", {
      action: "accept",
      content: { answers: {} },
    });
    expect(client.loadArtifact).toHaveBeenCalledOnce();
    expect(client.loadArtifact).toHaveBeenCalledWith("file-1");
    expect(runtime.getSnapshot("session-1").items).toEqual(
      expect.arrayContaining([
        expect.objectContaining({ kind: "message", role: "user", text: "Fix it" }),
      ]),
    );
    expect(listener).toHaveBeenCalled();
    runtime.close("session-1");
  });

  it("surfaces an open failure in the snapshot without rejecting the React effect", async () => {
    const client = createClient();
    vi.mocked(client.getSession).mockRejectedValueOnce(new Error("session unavailable"));
    const runtime = new EmbeddedAgentSessionRuntime(client);

    await expect(runtime.open("session-1")).resolves.toBeUndefined();

    expect(runtime.getSnapshot("session-1")).toMatchObject({
      connection: "connected",
      error: "session unavailable",
    });
    runtime.close("session-1");
  });

  it("reconciles durable items after the session stream reconnects", async () => {
    const client = createClient();
    vi.mocked(client.openStream)
      .mockResolvedValueOnce(
        new Response(
          new ReadableStream<Uint8Array>({
            start(controller) {
              controller.close();
            },
          }),
          { status: 200 },
        ),
      )
      .mockImplementation(openStream);
    vi.mocked(client.getItems).mockResolvedValueOnce({
      hasMore: false,
      items: [],
    });
    vi.mocked(client.getItems).mockResolvedValue({
      hasMore: false,
      items: [
        {
          id: "assistant-recovered",
          type: "message",
          response_id: "response-recovered",
          status: "completed",
          role: "assistant",
          content: [{ type: "output_text", text: "Recovered after reconnect" }],
        },
      ],
    });
    const runtime = new EmbeddedAgentSessionRuntime(client);

    await runtime.open("session-1");
    await vi.waitFor(
      () => expect(client.getItems).toHaveBeenCalledTimes(2),
      { timeout: 2_000 },
    );

    expect(runtime.getSnapshot("session-1").items).toEqual(
      expect.arrayContaining([
        expect.objectContaining({
          kind: "message",
          role: "assistant",
          text: "Recovered after reconnect",
        }),
      ]),
    );
    runtime.close("session-1");
  });

  it("rehydrates after response completion because idle precedes the durable item", async () => {
    const client = createClient();
    vi.mocked(client.getItems)
      .mockResolvedValueOnce({ hasMore: false, items: [] })
      .mockResolvedValue({
        hasMore: false,
        items: [
          {
            id: "assistant-final",
            type: "message",
            response_id: "response-final",
            status: "completed",
            role: "assistant",
            content: [{ type: "output_text", text: "Final answer" }],
          },
        ],
      });
    vi.mocked(client.openStream).mockResolvedValue(
      new Response(
        new ReadableStream<Uint8Array>({
          start(controller) {
            controller.enqueue(
              new TextEncoder().encode(
                [
                  "event: turn.started",
                  'data: {"id":"response-final","status":"in_progress"}',
                  "",
                  "event: turn.text.delta",
                  'data: {"delta":"Final ","response_id":"response-final"}',
                  "",
                  "event: turn.text.delta",
                  'data: {"delta":"answer","response_id":"response-final"}',
                  "",
                  "event: session.status",
                  'data: {"conversation_id":"session-1","status":"idle"}',
                  "",
                  "event: turn.item.done",
                  'data: {"item":{"id":"assistant-final","response_id":"response-final","type":"message","role":"assistant","status":"completed","content":[{"type":"output_text","text":"Final answer"}]}}',
                  "",
                  "event: turn.completed",
                  'data: {"id":"response-final","status":"completed"}',
                  "",
                ].join("\n"),
              ),
            );
          },
        }),
        { status: 200 },
      ),
    );
    const runtime = new EmbeddedAgentSessionRuntime(client);

    await runtime.open("session-1");
    await vi.waitFor(() => expect(client.getItems).toHaveBeenCalledTimes(3));

    const finalAnswers = runtime
      .getSnapshot("session-1")
      .items.filter(
        (item) =>
          item.kind === "message" &&
          item.role === "assistant" &&
          item.text === "Final answer",
      );
    expect(finalAnswers).toEqual([
      expect.objectContaining({
        kind: "message",
        role: "assistant",
        status: "completed",
        text: "Final answer",
      }),
    ]);
    runtime.close("session-1");
  });

  it("does not let a stale hydration overwrite a live running status", async () => {
    let resolveSession = (_session: {
      agentLabel: string;
      id: string;
      interactionMode: "pty";
      podKey: string;
      status: "idle";
      title: string;
    }) => {};
    const session = new Promise<{
      agentLabel: string;
      id: string;
      interactionMode: "pty";
      podKey: string;
      status: "idle";
      title: string;
    }>((resolve) => {
      resolveSession = resolve;
    });
    const client = createClient();
    vi.mocked(client.getSession).mockReturnValue(session);
    vi.mocked(client.openStream).mockImplementation((signal) =>
      Promise.resolve(
        new Response(
          new ReadableStream<Uint8Array>({
            start(controller) {
              controller.enqueue(
                new TextEncoder().encode(
                  'event: session.status\ndata: {"conversation_id":"session-1","status":"running"}\n\n',
                ),
              );
              signal.addEventListener(
                "abort",
                () => controller.close(),
                { once: true },
              );
            },
          }),
          { status: 200 },
        ),
      ),
    );
    const runtime = new EmbeddedAgentSessionRuntime(client);
    const opening = runtime.open("session-1");

    await vi.waitFor(() =>
      expect(runtime.getSnapshot("session-1").status).toBe("running"),
    );
    resolveSession({
      agentLabel: "codex-cli",
      id: "session-1",
      interactionMode: "pty",
      podKey: "pod-1",
      status: "idle",
      title: "Auth review",
    });
    await opening;

    expect(runtime.getSnapshot("session-1").status).toBe("running");
    runtime.close("session-1");
  });
});
