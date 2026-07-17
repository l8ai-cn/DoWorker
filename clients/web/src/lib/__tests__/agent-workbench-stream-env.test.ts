import { afterEach, describe, expect, it, vi } from "vitest";

afterEach(() => {
  vi.unstubAllEnvs();
  vi.resetModules();
});

describe("getAgentWorkbenchStreamBaseUrl", () => {
  it("uses the explicit Workbench stream URL", async () => {
    vi.stubEnv(
      "NEXT_PUBLIC_AGENT_WORKBENCH_STREAM_URL",
      "https://stream.example.test",
    );
    vi.stubEnv("NEXT_PUBLIC_API_URL", "");

    const { getAgentWorkbenchStreamBaseUrl } = await import("../env");

    expect(getAgentWorkbenchStreamBaseUrl()).toBe(
      "https://stream.example.test",
    );
  });

  it("derives the direct primary-domain URL when API calls use the Next proxy", async () => {
    vi.stubEnv("NEXT_PUBLIC_AGENT_WORKBENCH_STREAM_URL", "");
    vi.stubEnv("NEXT_PUBLIC_API_URL", "");
    vi.stubEnv("NEXT_PUBLIC_PRIMARY_DOMAIN", "localhost:29950");
    vi.stubEnv("NEXT_PUBLIC_USE_HTTPS", "false");

    const { getAgentWorkbenchStreamBaseUrl } = await import("../env");

    expect(getAgentWorkbenchStreamBaseUrl()).toBe("http://localhost:29950");
  });
});
