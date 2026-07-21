import { beforeEach, describe, expect, it, vi } from "vitest";

import { sessionStorageKey } from "@/lib/light-session";
import { fetchMarketplaceListingDetail, fetchMarketplaceListings } from "./catalog-api";

describe("marketplace catalog API", () => {
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

  it("reads the catalog through the same-origin marketplace API", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ items: [] }), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      }),
    );

    await fetchMarketplaceListings();

    expect(fetchMock).toHaveBeenCalledWith(
      expect.stringContaining("/api/marketplace/v1/markets/agent-cloud-market/listings"),
      expect.objectContaining({
        headers: expect.objectContaining({ Authorization: "Bearer market-token" }),
      }),
    );
  });

  it("encodes a listing slug before requesting its detail", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ slug: "delivery" }), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      }),
    );

    await fetchMarketplaceListingDetail("delivery tools");

    expect(fetchMock).toHaveBeenCalledWith(
      expect.stringContaining("/listings/delivery%20tools"),
      expect.anything(),
    );
  });
});
