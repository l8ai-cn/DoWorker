import { act, render, screen, waitFor } from "@testing-library/react";
import { StrictMode } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { EmbeddedSessionIframe } from "./EmbeddedSessionIframe";

const renderedAccess = vi.hoisted(() => [] as unknown[]);

vi.mock("./EmbeddedAgentWorkspace", () => ({
  EmbeddedAgentWorkspace: ({ access }: { access: unknown }) => {
    renderedAccess.push(access);
    return <div>Workspace</div>;
  },
}));

afterEach(() => {
  renderedAccess.length = 0;
  vi.unstubAllGlobals();
});

describe("EmbeddedSessionIframe", () => {
  it("does not render a worker without a server-issued embed context", () => {
    window.history.replaceState({}, "", "/iframe.html");

    render(<EmbeddedSessionIframe />);

    expect(screen.getByText("无法打开嵌入会话")).toBeInTheDocument();
    expect(screen.getByText("此嵌入工作区需要有效的会话上下文。")).toBeInTheDocument();
  });

  it("keeps a stable hook order while redeeming and opening the workspace", async () => {
    const parent = { postMessage: vi.fn() } as unknown as Window;
    vi.stubGlobal("parent", parent);
    const fetcher = vi.fn(async (input: RequestInfo | URL) => {
      const path = String(input);
      if (path === "/v1/embed-contexts/inspect") {
        return new Response(
          JSON.stringify({
            expires_at: 2_000_000_000,
            parent_origins: ["http://portal.example"],
          }),
          { status: 200, headers: { "Content-Type": "application/json" } },
        );
      }
      return new Response(
        JSON.stringify({
          access_token: "session-token",
          expires_at: 2_000_000_000,
          session_id: "conv-live",
          org_slug: "acme",
          capabilities: ["read", "write"],
          parent_origins: ["http://portal.example"],
        }),
        { status: 200, headers: { "Content-Type": "application/json" } },
      );
    });
    vi.stubGlobal("fetch", fetcher);
    window.history.replaceState({}, "", "/iframe.html?embed_context=context-token");

    render(
      <StrictMode>
        <EmbeddedSessionIframe />
      </StrictMode>,
    );

    expect(screen.getByText("正在等待嵌入页面建立连接…")).toBeInTheDocument();
    await waitFor(() =>
      expect(parent.postMessage).toHaveBeenCalledWith(
        { type: "agentsmesh.embed.ready", version: 1 },
        "http://portal.example",
      ),
    );

    act(() => {
      window.dispatchEvent(
        new MessageEvent("message", {
          data: {
            type: "agentsmesh.embed.open",
            version: 1,
            redemptionProof: "parent-proof",
          },
          origin: "http://portal.example",
          source: parent,
        }),
      );
    });

    expect(await screen.findByText("Workspace")).toBeInTheDocument();
    const access = renderedAccess.at(-1);
    expect(access).toMatchObject({
      baseUrl: window.location.origin,
      orgSlug: "acme",
      sessionId: "conv-live",
    });
    expect(await (access as { getAccessToken(): Promise<string> | string }).getAccessToken()).toBe(
      "session-token",
    );
    expect(fetcher).toHaveBeenCalledTimes(2);
    expect(fetcher).toHaveBeenLastCalledWith("/v1/embed-contexts/redeem", {
      method: "POST",
      headers: {
        Authorization: "Bearer context-token",
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ redemption_proof: "parent-proof" }),
      cache: "no-store",
    });
  });
});
