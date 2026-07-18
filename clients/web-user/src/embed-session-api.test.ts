import { describe, expect, it, vi } from "vitest";
import { create } from "@bufbuild/protobuf";
import {
  ArtifactDescriptorSchema,
  ArtifactGrantSchema,
  ArtifactRepresentationSchema,
} from "@do-worker/proto/agent_workbench/v2/artifact_pb";
import type { AgentArtifactTransportContext } from "@do-worker/agent-ui";

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
      "uploadAttachment",
    ]);
  });

  it("附件上传复用 embed bearer request 和 session 文件 endpoint", async () => {
    const fetcher = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          id: "file_12345678",
          metadata: { bytes: 7 },
          name: "notes.txt",
        }),
        { status: 200 },
      ),
    );
    const client = createEmbedSessionClient(access, fetcher);
    const file = new File(["content"], "notes.txt", { type: "text/plain" });

    await expect(client.uploadAttachment(file)).resolves.toMatchObject({
      id: "file_12345678",
      name: "notes.txt",
    });
    expect(String(fetcher.mock.calls[0]?.[0])).toBe(
      "https://api.example.test/v1/embed/sessions/conv_embed/resources/files",
    );
    expect(requestAuthorization(fetcher, 0)).toBe("Bearer session-token");
    const init = fetcher.mock.calls[0]?.[1] as RequestInit;
    expect(init.method).toBe("POST");
    expect(init.body).toBeInstanceOf(FormData);
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

  it("resourceId 通过精确成果身份 endpoint 读取", async () => {
    const fetcher = vi.fn().mockResolvedValueOnce(
      new Response("artifact-bytes", {
        status: 200,
        headers: { "Content-Type": "application/pdf" },
      }),
    );
    const client = createEmbedSessionClient(access, fetcher);

    const blob = await client.loadResource("session-file:file_12345678", artifactContext());

    await expect(blob.text()).resolves.toBe("artifact-bytes");
    expect(String(fetcher.mock.calls[0]?.[0])).toBe(
      "https://api.example.test/v1/embed/sessions/conv_embed/artifacts/content?" +
        "artifact_id=artifact-1&digest=sha256%3Aabc&" +
        "representation_id=preview-pdf&revision=2",
    );
  });

  it.each(["", "workspace:preview.pdf", "artifact-cache:preview-1", "session-file:"])(
    "非法 artifact resourceId 在读取 token 前被拒绝: %j",
    async (resourceId) => {
      const getAccessToken = vi.fn(() => "session-token");
      const fetcher = vi.fn();
      const client = createEmbedSessionClient({ ...access, getAccessToken }, fetcher);

      await expect(client.loadResource(resourceId, artifactContext())).rejects.toThrow(
        `artifact_resource_unsupported:${resourceId}`,
      );
      expect(getAccessToken).not.toHaveBeenCalled();
      expect(fetcher).not.toHaveBeenCalled();
    },
  );

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
    const client = createEmbedSessionClient({ ...access, getAccessToken }, fetcher);

    await expect(client.loadDownload("https://files.attacker.test/result.bin")).rejects.toThrow(
      "agent_workbench_download_origin_mismatch",
    );
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

function artifactContext(): AgentArtifactTransportContext {
  const representation = create(ArtifactRepresentationSchema, {
    digest: "sha256:abc",
    mediaType: "application/pdf",
    representationId: "preview-pdf",
    revision: 2n,
  });
  const descriptor = create(ArtifactDescriptorSchema, {
    artifactId: "artifact-1",
    grants: [
      create(ArtifactGrantSchema, {
        actions: ["artifact.download"],
        grantId: "grant-download",
        representationIds: ["preview-pdf"],
      }),
    ],
    representations: [representation],
    revision: 2n,
  });
  return {
    artifactId: "artifact-1",
    descriptor,
    representation,
    representationId: "preview-pdf",
    sessionId: "conv_embed",
  };
}

function requestAuthorization(fetcher: ReturnType<typeof vi.fn>, index: number): string | null {
  const init = fetcher.mock.calls[index]?.[1] as RequestInit | undefined;
  return new Headers(init?.headers).get("Authorization");
}
