import { beforeEach, describe, expect, it, vi } from "vitest";

import { sessionStorageKey } from "@/lib/light-session";
import { fetchOrganizationApplications } from "./application-api";

describe("organization applications API", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
    window.localStorage.clear();
    window.localStorage.setItem(
      sessionStorageKey(window.location.origin),
      JSON.stringify({
        access_token: "market-token",
        expires_at: Math.floor(Date.now() / 1000) + 3600,
      }),
    );
  });

  it("loads enabled applications for the real organization ID with authentication", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ applications: [] }), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      }),
    );

    await fetchOrganizationApplications(9);

    expect(fetchMock).toHaveBeenCalledWith(
      expect.stringContaining("/organizations/9/applications"),
      expect.objectContaining({
        headers: expect.objectContaining({
          Authorization: "Bearer market-token",
        }),
      }),
    );
  });
});
