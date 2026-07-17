import { beforeEach, describe, expect, it, vi } from "vitest";
import {
  getMobilePodPreviewSession,
  replaceWithMobilePodPreview,
} from "./mobile-pod-preview";
import { apiFetch } from "./api-fetch";
import { readOrgSlug } from "./auth-store";

vi.mock("./api-fetch", () => ({ apiFetch: vi.fn() }));
vi.mock("./auth-store", () => ({ readOrgSlug: vi.fn() }));

describe("mobile Pod preview API", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(readOrgSlug).mockReturnValue("dev-org");
  });

  it("returns only the backend-issued preview session URL", async () => {
    vi.mocked(apiFetch).mockResolvedValue(
      new Response(
        JSON.stringify({
          preview_base_url: "https://relay.example/preview/pod-1/",
          session_url: "https://relay.example/preview/pod-1/__session?token=one-time",
          expires_at: "2026-07-13T00:00:00Z",
        }),
        { status: 200 },
      ),
    );

    await expect(getMobilePodPreviewSession("pod-1")).resolves.toEqual({
      sessionUrl: "https://relay.example/preview/pod-1/__session?token=one-time",
    });
    expect(apiFetch).toHaveBeenCalledWith("/api/v1/orgs/dev-org/pods/pod-1/preview");
  });

  it("rejects an invalid preview response instead of navigating to a guessed URL", async () => {
    vi.mocked(apiFetch).mockResolvedValue(
      new Response(JSON.stringify({ preview_base_url: "https://relay.example/preview/pod-1/" }), {
        status: 200,
      }),
    );

    await expect(getMobilePodPreviewSession("pod-1")).rejects.toThrow("Preview session URL 无效");
  });

  it("replaces the Preview route after obtaining the one-time session URL", async () => {
    vi.mocked(apiFetch).mockResolvedValue(
      new Response(JSON.stringify({ session_url: "https://relay.example/preview/session" }), {
        status: 200,
      }),
    );
    const replace = vi.fn();

    await replaceWithMobilePodPreview("pod-1", replace);

    expect(replace).toHaveBeenCalledWith("https://relay.example/preview/session");
  });
});
