import { afterEach, describe, expect, it, vi } from "vitest";
import type { Browser } from "@playwright/test";
import { authenticateE2ETestUser } from "./test-user-auth";
import { E2E_ECHO_AGENT_SLUG } from "./e2e-echo-runner";

const originalFetch = globalThis.fetch;

afterEach(() => {
  globalThis.fetch = originalFetch;
});

describe("E2E test-user authentication", () => {
  it("creates browser auth state without enumerating or terminating Pods", async () => {
    const fetchMock = vi.fn<typeof fetch>()
      .mockResolvedValueOnce(jsonResponse({ token: "cleanup-token" }))
      .mockResolvedValueOnce(jsonResponse({
        items: [
          { podKey: "production-video", alias: "Video", agentSlug: "video-studio" },
          { podKey: "production-pattern", alias: "[e2e:bad] Pattern", agentSlug: E2E_ECHO_AGENT_SLUG },
        ],
      }))
      .mockResolvedValueOnce(jsonResponse({
        token: "test-token",
        refreshToken: "refresh-token",
        expiresIn: 3600,
      }));
    globalThis.fetch = fetchMock;
    const page = { goto: vi.fn().mockResolvedValue(undefined) };
    const context = {
      addInitScript: vi.fn().mockResolvedValue(undefined),
      newPage: vi.fn().mockResolvedValue(page),
      storageState: vi.fn().mockResolvedValue(undefined),
      close: vi.fn().mockResolvedValue(undefined),
    };
    const browser = {
      newContext: vi.fn().mockResolvedValue(context),
    } as unknown as Browser;

    await authenticateE2ETestUser(browser);

    expect(fetchMock).toHaveBeenCalledTimes(3);
    expect(String(fetchMock.mock.calls[0]?.[0])).toContain("AuthService/Login");
    expect(String(fetchMock.mock.calls[1]?.[0])).toContain("PodService/ListPods");
    expect(String(fetchMock.mock.calls[2]?.[0])).toContain("AuthService/Login");
    expect(fetchMock.mock.calls.some(([url]) => String(url).includes("TerminatePod"))).toBe(false);
    expect(context.storageState).toHaveBeenCalledWith({ path: ".auth/user.json" });
    expect(context.close).toHaveBeenCalledOnce();
  });
});

function jsonResponse(body: unknown): Response {
  return new Response(JSON.stringify(body), {
    status: 200,
    headers: { "Content-Type": "application/json" },
  });
}
