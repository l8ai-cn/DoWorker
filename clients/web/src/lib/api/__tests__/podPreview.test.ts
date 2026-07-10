import { describe, expect, it, vi } from "vitest";
import { getPodPreviewSession } from "../podPreview";
import { lightFetch } from "@/lib/light-auth/api-fetch";

vi.mock("@/lib/light-auth/api-fetch", () => ({
  lightFetch: vi.fn().mockResolvedValue({
    preview_base_url: "http://localhost:10000/preview/pod-1/",
    session_url: "http://localhost:10000/preview/pod-1/__session?token=secret",
    token: "secret",
    expires_at: "2026-07-10T00:00:00Z",
  }),
}));

describe("getPodPreviewSession", () => {
  it("requests the authenticated backend preview session endpoint", async () => {
    const session = await getPodPreviewSession("acme", "pod-1");

    expect(lightFetch).toHaveBeenCalledWith("/api/v1/orgs/acme/pods/pod-1/preview", {
      authenticated: true,
    });
    expect(session.session_url).toContain("__session");
  });
});
