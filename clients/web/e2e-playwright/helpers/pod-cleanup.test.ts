import { afterEach, describe, expect, it, vi } from "vitest";
import {
  createE2EPodAlias,
  registerE2ECreatedPod,
  resetRegisteredE2EPodsForTest,
  terminateRegisteredE2EPods,
  terminateStaleMarkedE2EPods,
} from "./pod-cleanup";
import { E2E_ECHO_AGENT_SLUG } from "./e2e-echo-runner";

const originalFetch = globalThis.fetch;

afterEach(() => {
  globalThis.fetch = originalFetch;
  resetRegisteredE2EPodsForTest();
});

describe("registered E2E pod cleanup", () => {
  it("does not call the API when this process has not created a pod", async () => {
    const fetchMock = vi.fn<typeof fetch>();
    globalThis.fetch = fetchMock;

    await expect(terminateRegisteredE2EPods()).resolves.toBe(0);
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("verifies the exact marked E2E pod before terminating it", async () => {
    const alias = createE2EPodAlias("contract pod");
    registerE2ECreatedPod("pod-e2e-1", alias);
    const fetchMock = vi.fn<typeof fetch>()
      .mockResolvedValueOnce(jsonResponse({ token: "test-token" }))
      .mockResolvedValueOnce(jsonResponse({
        podKey: "pod-e2e-1",
        alias,
        agentSlug: E2E_ECHO_AGENT_SLUG,
        status: "running",
      }))
      .mockResolvedValueOnce(jsonResponse({ message: "terminated" }));
    globalThis.fetch = fetchMock;

    await expect(terminateRegisteredE2EPods()).resolves.toBe(1);
    expect(fetchMock).toHaveBeenCalledTimes(3);
    expect(String(fetchMock.mock.calls[1]?.[0])).toContain("PodService/GetPod");
    expect(String(fetchMock.mock.calls[2]?.[0])).toContain("PodService/TerminatePod");
    expect(fetchMock.mock.calls[2]?.[1]).toMatchObject({
      body: JSON.stringify({ orgSlug: "dev-org", podKey: "pod-e2e-1" }),
      headers: expect.objectContaining({
        "X-E2E-Caller": "terminateRegisteredE2EPods",
      }),
    });
    expect(fetchMock.mock.calls.some(([url]) => String(url).includes("ListPods"))).toBe(false);
  });

  it("throws without terminating when a registered key no longer identifies the marked E2E pod", async () => {
    const alias = createE2EPodAlias("contract pod");
    registerE2ECreatedPod("pod-e2e-1", alias);
    const fetchMock = vi.fn<typeof fetch>()
      .mockResolvedValueOnce(jsonResponse({ token: "test-token" }))
      .mockResolvedValueOnce(jsonResponse({
        podKey: "pod-e2e-1",
        alias: "real production pod",
        agentSlug: E2E_ECHO_AGENT_SLUG,
        status: "running",
      }));
    globalThis.fetch = fetchMock;

    await expect(terminateRegisteredE2EPods()).rejects.toThrow(
      "identity does not match the E2E record",
    );
    expect(fetchMock).toHaveBeenCalledTimes(2);
    expect(fetchMock.mock.calls.some(([url]) => String(url).includes("TerminatePod"))).toBe(false);
  });

  it("throws when a registered Pod cleanup cannot authenticate", async () => {
    registerE2ECreatedPod("pod-e2e-1", createE2EPodAlias("contract pod"));
    globalThis.fetch = vi.fn<typeof fetch>().mockResolvedValue(
      new Response("unavailable", { status: 503 }),
    );

    await expect(terminateRegisteredE2EPods()).rejects.toThrow(
      "registered cleanup login returned HTTP 503",
    );
  });

  it("recovers only stale Pods with the strict E2E marker and e2e-echo agent", async () => {
    const fetchMock = vi.fn<typeof fetch>()
      .mockResolvedValueOnce(jsonResponse({ token: "test-token" }))
      .mockResolvedValueOnce(jsonResponse({
        items: [
          { podKey: "pod-stale", alias: "[e2e:deadbeefcafe] stale", agentSlug: E2E_ECHO_AGENT_SLUG, status: "running" },
          { podKey: "pod-video", alias: "[e2e:deadbeefcafe] video", agentSlug: "video-studio", status: "running" },
          { podKey: "pod-seedance", alias: "Seedance production", agentSlug: E2E_ECHO_AGENT_SLUG, status: "running" },
          { podKey: "pod-pattern", alias: "[e2e:bad] pattern", agentSlug: E2E_ECHO_AGENT_SLUG, status: "running" },
          { podKey: "pod-completed", alias: "[e2e:deadbeefcafe] completed", agentSlug: E2E_ECHO_AGENT_SLUG, status: "completed" },
          { podKey: "pod-terminated", alias: "[e2e:deadbeefcafe] terminated", agentSlug: E2E_ECHO_AGENT_SLUG, status: "terminated" },
        ],
      }))
      .mockResolvedValueOnce(jsonResponse({
        podKey: "pod-stale",
        alias: "[e2e:deadbeefcafe] stale",
        agentSlug: E2E_ECHO_AGENT_SLUG,
        status: "running",
      }))
      .mockResolvedValueOnce(jsonResponse({ message: "terminated" }));
    globalThis.fetch = fetchMock;

    await expect(terminateStaleMarkedE2EPods()).resolves.toBe(1);
    expect(fetchMock).toHaveBeenCalledTimes(4);
    expect(String(fetchMock.mock.calls[1]?.[0])).toContain("PodService/ListPods");
    expect(String(fetchMock.mock.calls[1]?.[1]?.body)).toContain(
      '"status":"queued,initializing,running,paused,disconnected"',
    );
    expect(String(fetchMock.mock.calls[2]?.[1]?.body)).toContain("pod-stale");
    expect(String(fetchMock.mock.calls[3]?.[1]?.body)).toContain("pod-stale");
    expect(fetchMock.mock.calls.some(([, init]) =>
      String(init?.body).includes("pod-video") ||
      String(init?.body).includes("pod-seedance") ||
      String(init?.body).includes("pod-pattern") ||
      String(init?.body).includes("pod-completed") ||
      String(init?.body).includes("pod-terminated"),
    )).toBe(false);
  });
});

function jsonResponse(body: unknown): Response {
  return new Response(JSON.stringify(body), {
    status: 200,
    headers: { "Content-Type": "application/json" },
  });
}
