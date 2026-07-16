import { describe, expect, it, vi } from "vitest";

import {
  clearEmbedContextFromLocation,
  inspectEmbedContext,
  redeemEmbedContextOnce,
  readEmbedContext,
  redeemEmbedContext,
} from "./embed-context";

describe("embed context", () => {
  it("requires exactly one context query value", () => {
    expect(() => readEmbedContext("")).toThrow("embed_context is required");
    expect(() => readEmbedContext("?embed_context=one&embed_context=two")).toThrow(
      "embed_context must appear exactly once",
    );
    expect(readEmbedContext("?embed_context=signed-value")).toBe("signed-value");
  });

  it("inspects allowed parent origins without redeeming the context", async () => {
    const fetcher = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          expires_at: 123,
          parent_origins: ["https://portal.example"],
        }),
        { status: 200 },
      ),
    );

    await expect(inspectEmbedContext("context-token", fetcher)).resolves.toEqual({
      expiresAt: 123,
      parentOrigins: ["https://portal.example"],
    });
    expect(fetcher).toHaveBeenCalledWith("/v1/embed-contexts/inspect", {
      method: "POST",
      headers: { Authorization: "Bearer context-token" },
      cache: "no-store",
    });
  });

  it("redeems the context with its parent-held proof", async () => {
    const fetcher = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          access_token: "session-token",
          expires_at: 123,
          session_id: "conv_embed",
          capabilities: ["read", "write"],
          parent_origins: ["https://portal.example"],
        }),
        { status: 200 },
      ),
    );

    await expect(redeemEmbedContext("context-token", "redemption-proof", fetcher)).resolves.toEqual(
      {
        accessToken: "session-token",
        expiresAt: 123,
        sessionId: "conv_embed",
        capabilities: ["read", "write"],
        parentOrigins: ["https://portal.example"],
      },
    );
    expect(fetcher).toHaveBeenCalledWith("/v1/embed-contexts/redeem", {
      method: "POST",
      headers: {
        Authorization: "Bearer context-token",
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ redemption_proof: "redemption-proof" }),
      cache: "no-store",
    });
  });

  it("accepts the explicit Agent Workspace capability set", async () => {
    const fetcher = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          access_token: "session-token",
          expires_at: 123,
          session_id: "conv_embed",
          capabilities: ["read", "write", "approve", "terminal", "control"],
          parent_origins: ["https://portal.example"],
        }),
        { status: 200 },
      ),
    );

    const access = await redeemEmbedContext("context-token", "redemption-proof", fetcher);

    expect(access.capabilities).toEqual(["read", "write", "approve", "terminal", "control"]);
  });

  it("deduplicates only concurrent redemption and releases bearer material after settlement", async () => {
    let resolveFirst!: (response: Response) => void;
    const firstResponse = new Promise<Response>((resolve) => {
      resolveFirst = resolve;
    });
    const accessBody = JSON.stringify({
      access_token: "session-token",
      expires_at: 123,
      session_id: "conv_embed",
      capabilities: ["read"],
      parent_origins: ["https://portal.example"],
    });
    const fetcher = vi
      .fn()
      .mockImplementationOnce(() => firstResponse)
      .mockResolvedValueOnce(new Response(accessBody, { status: 200 }));

    const first = redeemEmbedContextOnce("one-shot-context", "proof", fetcher);
    const concurrent = redeemEmbedContextOnce("one-shot-context", "proof", fetcher);
    expect(concurrent).toBe(first);
    expect(fetcher).toHaveBeenCalledTimes(1);

    resolveFirst(new Response(accessBody, { status: 200 }));
    await first;
    await redeemEmbedContextOnce("one-shot-context", "proof", fetcher);

    expect(fetcher).toHaveBeenCalledTimes(2);
  });

  it("removes the redeemed context from the address bar", () => {
    window.history.replaceState({}, "", "/iframe.html?embed_context=signed-value&theme=dark");

    clearEmbedContextFromLocation();

    expect(window.location.pathname + window.location.search).toBe("/iframe.html?theme=dark");
  });
});
