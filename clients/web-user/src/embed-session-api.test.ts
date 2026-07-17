import { describe, expect, it, vi } from "vitest";

import { createEmbedSessionClient } from "./embed-session-api";

const access = {
  baseUrl: "https://api.example.test",
  getAccessToken: vi.fn(() => "session-token"),
  orgSlug: "acme",
  sessionId: "conv_embed",
};

describe("embedded session API", () => {
  it("只暴露宿主资源，不暴露旧对话命令", () => {
    const client = createEmbedSessionClient(access);

    expect(Object.keys(client).sort()).toEqual([
      "getRelayConnection",
      "getSession",
      "getTerminals",
      "loadDownload",
      "loadResource",
    ]);
  });

  it("一次性读取 session 元数据并刷新 bearer token", async () => {
    const getAccessToken = vi.fn().mockReturnValueOnce("token-1").mockReturnValueOnce("token-2");
    const fetcher = vi.fn().mockImplementation(
      async () =>
        new Response(
          JSON.stringify({
            agent_name: "codex-cli",
            interaction_mode: "acp",
            title: "Embedded task",
          }),
          { status: 200 },
        ),
    );
    const client = createEmbedSessionClient({ ...access, getAccessToken }, fetcher);

    await expect(client.getSession()).resolves.toEqual({
      agentLabel: "codex-cli",
      interactionMode: "acp",
      title: "Embedded task",
    });
    await client.getSession();

    expect(getAccessToken).toHaveBeenCalledTimes(2);
    expect(String(fetcher.mock.calls[0]?.[0])).toBe(
      "https://api.example.test/v1/embed/sessions/conv_embed",
    );
    expect(requestAuthorization(fetcher, 0)).toBe("Bearer token-1");
    expect(requestAuthorization(fetcher, 1)).toBe("Bearer token-2");
  });

  it("resourceId 使用 session-bound 文件和 workspace endpoint", async () => {
    const fetcher = vi
      .fn()
      .mockResolvedValueOnce(
        new Response("image-bytes", {
          status: 200,
          headers: { "Content-Type": "image/png" },
        }),
      )
      .mockResolvedValueOnce(
        new Response("workspace-bytes", {
          status: 200,
          headers: { "Content-Type": "image/png" },
        }),
      );
    const client = createEmbedSessionClient(access, fetcher);

    await expect(client.loadResource("file-1")).resolves.toEqual(expect.any(Blob));
    const workspace = await client.loadResource("workspace:deliverables/preview.png");

    await expect(workspace.text()).resolves.toBe("workspace-bytes");
    expect(String(fetcher.mock.calls[0]?.[0])).toBe(
      "https://api.example.test/v1/embed/sessions/conv_embed/resources/files/file-1/content",
    );
    expect(String(fetcher.mock.calls[1]?.[0])).toBe(
      "https://api.example.test/v1/embed/sessions/conv_embed/resources/environments/workspace/artifacts/content/deliverables/preview.png",
    );
  });

  it("downloadUrl 走带 access token 的明确 fetch", async () => {
    const fetcher = vi.fn().mockResolvedValue(
      new Response("download", {
        status: 200,
        headers: { "Content-Type": "application/octet-stream" },
      }),
    );
    const client = createEmbedSessionClient(access, fetcher);

    const blob = await client.loadDownload("/downloads/result.bin");

    await expect(blob.text()).resolves.toBe("download");
    expect(String(fetcher.mock.calls[0]?.[0])).toBe(
      "https://api.example.test/downloads/result.bin",
    );
    expect(requestAuthorization(fetcher, 0)).toBe("Bearer session-token");
  });

  it("跨源 downloadUrl 在读取 token 前被拒绝", async () => {
    const getAccessToken = vi.fn(() => "session-token");
    const fetcher = vi.fn();
    const client = createEmbedSessionClient(
      { ...access, getAccessToken },
      fetcher,
    );

    await expect(
      client.loadDownload("https://files.attacker.test/result.bin"),
    ).rejects.toThrow("agent_workbench_download_origin_mismatch");
    expect(getAccessToken).not.toHaveBeenCalled();
    expect(fetcher).not.toHaveBeenCalled();
  });

  it("保留 terminal inventory 和 Relay data plane", async () => {
    const fetcher = vi.fn(async (input: RequestInfo | URL) => {
      if (String(input).endsWith("/resources/terminals")) {
        return new Response(
          JSON.stringify({
            data: [
              {
                id: "terminal-main",
                name: "main",
                metadata: { running: true },
              },
            ],
          }),
          { status: 200 },
        );
      }
      return new Response(
        JSON.stringify({
          relay_url: "wss://relay.example.test",
          token: "relay-token",
          pod_key: "pod-1",
        }),
        { status: 200 },
      );
    });
    const client = createEmbedSessionClient(access, fetcher);

    await expect(client.getTerminals()).resolves.toEqual([
      {
        id: "terminal-main",
        label: "main",
        status: "connected",
        writable: true,
      },
    ]);
    await expect(client.getRelayConnection()).resolves.toEqual({
      relayUrl: "wss://relay.example.test",
      token: "relay-token",
      podKey: "pod-1",
    });
  });
});

function requestAuthorization(fetcher: ReturnType<typeof vi.fn>, index: number): string | null {
  const init = fetcher.mock.calls[index]?.[1] as RequestInit | undefined;
  return new Headers(init?.headers).get("Authorization");
}
