import { describe, expect, it, vi } from "vitest";

import { createEmbedSessionClient } from "./embed-session-api";

describe("embedded session API", () => {
  it("uses the restricted session routes with the redeemed bearer token", async () => {
    const fetcher = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          agent_name: "codex-cli",
          id: "conv_embed",
          interaction_mode: "acp",
          pod_key: "pod-1",
          title: "Embedded task",
          status: "idle",
        }),
        { status: 200 },
      ),
    );
    const client = createEmbedSessionClient(
      {
        accessToken: "session-token",
        expiresAt: 123,
        sessionId: "conv_embed",
        capabilities: ["read"],
        parentOrigins: ["https://portal.example"],
      },
      fetcher,
    );

    await expect(client.getSession()).resolves.toEqual({
      agentLabel: "codex-cli",
      id: "conv_embed",
      interactionMode: "acp",
      podKey: "pod-1",
      runnerId: null,
      title: "Embedded task",
      totalCostUsd: null,
      status: "idle",
    });
    expect(fetcher).toHaveBeenCalledWith("/v1/embed/sessions/conv_embed", {
      headers: { Authorization: "Bearer session-token" },
      cache: "no-store",
    });
  });

  it("does not create a mutation client for a read-only context", () => {
    const client = createEmbedSessionClient({
      accessToken: "session-token",
      expiresAt: 123,
      sessionId: "conv_embed",
      capabilities: ["read"],
      parentOrigins: ["https://portal.example"],
    });

    expect(client.sendMessage).toBeUndefined();
  });

  it("separates prompt submission from session control capabilities", () => {
    const writeClient = createEmbedSessionClient({
      accessToken: "session-token",
      expiresAt: 123,
      sessionId: "conv_embed",
      capabilities: ["read", "write"],
      parentOrigins: ["https://portal.example"],
    });
    const controlClient = createEmbedSessionClient({
      accessToken: "session-token",
      expiresAt: 123,
      sessionId: "conv_embed",
      capabilities: ["read", "control"],
      parentOrigins: ["https://portal.example"],
    });

    expect(writeClient.sendMessage).toBeDefined();
    expect(writeClient.interrupt).toBeUndefined();
    expect(writeClient.getAcpRelayConnection).toBeUndefined();
    expect(controlClient.sendMessage).toBeUndefined();
    expect(controlClient.interrupt).toBeDefined();
    expect(controlClient.getAcpRelayConnection).toBeDefined();
  });

  it("loads an artifact blob through the scoped read route", async () => {
    const fetcher = vi.fn().mockResolvedValue(
      new Response("image-bytes", {
        status: 200,
        headers: { "Content-Type": "image/png" },
      }),
    );
    const client = createEmbedSessionClient(
      {
        accessToken: "session-token",
        expiresAt: 123,
        sessionId: "conv_embed",
        capabilities: ["read"],
        parentOrigins: ["https://portal.example"],
      },
      fetcher,
    );

    const blob = await client.loadArtifact!("file-1");

    expect(blob.type).toBe("image/png");
    await expect(blob.text()).resolves.toBe("image-bytes");
    expect(fetcher).toHaveBeenCalledWith(
      "/v1/embed/sessions/conv_embed/resources/files/file-1/content",
      {
        headers: { Authorization: "Bearer session-token" },
        cache: "no-store",
      },
    );
  });

  it("loads a generated workspace artifact from the session-bound Runner", async () => {
    const fetcher = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          object: "session.environment.filesystem.file_content",
          path: "deliverables/preview.png",
          content_type: "image/png",
          encoding: "base64",
          content: "aW1hZ2UtYnl0ZXM=",
          bytes: 11,
          truncated: false,
        }),
        { status: 200 },
      ),
    );
    const client = createEmbedSessionClient(
      {
        accessToken: "session-token",
        expiresAt: 123,
        sessionId: "conv_embed",
        capabilities: ["read"],
        parentOrigins: ["https://portal.example"],
      },
      fetcher,
    );

    const blob = await client.loadArtifact!("workspace:deliverables/preview.png");

    expect(blob.type).toBe("image/png");
    await expect(blob.text()).resolves.toBe("image-bytes");
    expect(fetcher).toHaveBeenCalledWith(
      "/v1/embed/sessions/conv_embed/resources/environments/workspace/filesystem/deliverables/preview.png",
      {
        headers: { Authorization: "Bearer session-token" },
        cache: "no-store",
      },
    );
  });

  it("discovers generated deliverables from Runner workspace changes", async () => {
    const fetcher = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          object: "list",
          data: [
            { path: "deliverables/demo.mp4", status: "untracked" },
            { path: "src/main.ts", status: "modified" },
          ],
          has_more: false,
        }),
        { status: 200 },
      ),
    );
    const client = createEmbedSessionClient(
      {
        accessToken: "session-token",
        expiresAt: 123,
        sessionId: "conv_embed",
        capabilities: ["read"],
        parentOrigins: ["https://portal.example"],
      },
      fetcher,
    );

    await expect(client.listWorkspaceArtifacts?.()).resolves.toEqual([
      expect.objectContaining({
        artifactId: "workspace:deliverables/demo.mp4",
        filename: "demo.mp4",
        mimeType: "video/mp4",
      }),
    ]);
    expect(fetcher).toHaveBeenCalledWith(
      "/v1/embed/sessions/conv_embed/resources/environments/workspace/changes",
      {
        headers: { Authorization: "Bearer session-token" },
        cache: "no-store",
      },
    );
  });

  it("posts a text message only to the scoped embed event route", async () => {
    const fetcher = vi.fn().mockResolvedValue(
      new Response(JSON.stringify({ queued: true, item_id: "item_1" }), { status: 202 }),
    );
    const client = createEmbedSessionClient(
      {
        accessToken: "session-token",
        expiresAt: 123,
        sessionId: "conv_embed",
        capabilities: ["read", "write"],
        parentOrigins: ["https://portal.example"],
      },
      fetcher,
    );

    await expect(client.sendMessage?.("hello")).resolves.toEqual({ itemId: "item_1" });
    expect(fetcher).toHaveBeenCalledWith("/v1/embed/sessions/conv_embed/events", {
      method: "POST",
      headers: {
        Authorization: "Bearer session-token",
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        type: "message",
        data: { role: "user", content: [{ type: "input_text", text: "hello" }] },
      }),
      cache: "no-store",
    });
  });

  it("exposes interrupt, approval, terminal inventory, and Relay connection by capability", async () => {
    const fetcher = vi.fn(async (input: RequestInfo | URL) => {
      const path = input.toString();
      if (path.endsWith("/resources/terminals")) {
        return new Response(
          JSON.stringify({
            data: [
              {
                id: "terminal_tui_main",
                name: "main:tui",
                metadata: { running: true },
              },
            ],
          }),
          { status: 200 },
        );
      }
      if (path.endsWith("relay-connection")) {
        return new Response(
          JSON.stringify({
            relay_url: "ws://relay.example",
            token: "relay-token",
            pod_key: "pod-1",
          }),
          { status: 200 },
        );
      }
      return new Response(JSON.stringify({ queued: true }), { status: 202 });
    });
    const client = createEmbedSessionClient(
      {
        accessToken: "session-token",
        expiresAt: 123,
        sessionId: "conv_embed",
        capabilities: ["read", "write", "approve", "terminal", "control"],
        parentOrigins: ["https://portal.example"],
      },
      fetcher as typeof fetch,
    );

    await client.interrupt?.();
    await client.resolvePermission?.("permission-1", {
      action: "accept",
      content: { answers: { confirmation: ["yes"] } },
    });
    await expect(client.getTerminals?.()).resolves.toEqual([
      {
        id: "terminal_tui_main",
        label: "main:tui",
        status: "connected",
        writable: true,
      },
    ]);
    await expect(client.getRelayConnection?.()).resolves.toEqual({
      relayUrl: "ws://relay.example",
      token: "relay-token",
      podKey: "pod-1",
    });
    await expect(client.getAcpRelayConnection?.()).resolves.toEqual({
      relayUrl: "ws://relay.example",
      token: "relay-token",
      podKey: "pod-1",
    });

    expect(fetcher).toHaveBeenCalledWith(
      "/v1/embed/sessions/conv_embed/events",
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({ type: "interrupt", data: {} }),
      }),
    );
    expect(fetcher).toHaveBeenCalledWith(
      "/v1/embed/sessions/conv_embed/elicitations/permission-1/resolve",
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({
          action: "accept",
          content: { answers: { confirmation: ["yes"] } },
        }),
      }),
    );
    expect(fetcher).toHaveBeenCalledWith(
      "/v1/embed/sessions/conv_embed/acp-relay-connection",
      expect.objectContaining({
        headers: { Authorization: "Bearer session-token" },
      }),
    );
  });
});
