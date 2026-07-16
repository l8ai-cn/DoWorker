import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { lightListExecutionClusters } from "./execution-clusters";
import { resolveLightBaseUrl, writeLightSession } from "@/lib/light-session";

const ORIGIN = resolveLightBaseUrl();

describe("lightListExecutionClusters", () => {
  let originalFetch: typeof fetch;

  beforeEach(() => {
    originalFetch = globalThis.fetch;
    window.localStorage.clear();
    writeLightSession({
      accessToken: "tok",
      refreshToken: "refresh",
      expiresAt: Math.floor(Date.now() / 1000) + 3600,
      baseUrl: ORIGIN,
    });
  });

  afterEach(() => {
    globalThis.fetch = originalFetch;
    window.localStorage.clear();
  });

  it("lists organization-scoped execution clusters over authenticated Connect", async () => {
    const fetchSpy = vi.fn<typeof fetch>(
      async () =>
        new Response(
          JSON.stringify({
            items: [
              {
                id: "12",
                slug: "local",
                name: "本地集群",
                kind: "local",
                status: "ready",
              },
            ],
          }),
          { status: 200, headers: { "Content-Type": "application/json" } },
        ),
    );
    globalThis.fetch = fetchSpy as typeof fetch;

    const clusters = await lightListExecutionClusters("dev-org");

    expect(clusters).toEqual([
      {
        id: "12",
        slug: "local",
        name: "本地集群",
        kind: "local",
        status: "ready",
      },
    ]);
    const [url, init] = fetchSpy.mock.calls[0];
    expect(String(url)).toBe(
      `${ORIGIN}/proto.execution_cluster.v1.ExecutionClusterService/ListExecutionClusters`,
    );
    expect((init as RequestInit).body).toBe(
      JSON.stringify({ orgSlug: "dev-org" }),
    );
    expect((init as RequestInit).headers).toMatchObject({
      Authorization: "Bearer tok",
    });
  });

  it("rejects an unsafe numeric cluster id instead of rounding it", async () => {
    globalThis.fetch = vi.fn<typeof fetch>(
      async () =>
        new Response(
          JSON.stringify({
            items: [
              {
                id: 9_007_199_254_740_992,
                slug: "remote",
                name: "线上集群",
                kind: "remote",
                status: "ready",
              },
            ],
          }),
          { status: 200, headers: { "Content-Type": "application/json" } },
        ),
    ) as typeof fetch;

    await expect(lightListExecutionClusters("dev-org")).rejects.toThrow(
      "unsafe execution cluster id",
    );
  });
});
