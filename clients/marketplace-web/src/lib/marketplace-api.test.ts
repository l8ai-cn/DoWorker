import { afterEach, describe, expect, it, vi } from "vitest";

import { getListing, getMarket, listListings } from "./marketplace-api";

describe("marketplace api", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.unstubAllEnvs();
  });

  it("uses the internal API URL and configured request host", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(JSON.stringify({ slug: "do-worker-market" })),
    );
    vi.stubGlobal("fetch", fetchMock);
    vi.stubEnv("MARKETPLACE_API_INTERNAL_URL", "http://marketplace-api:8080/");
    vi.stubEnv("MARKETPLACE_REQUEST_HOST", "market.example.cn");

    await getMarket();

    expect(fetchMock).toHaveBeenCalledWith(
      "http://marketplace-api:8080/api/marketplace/v1/markets/do-worker-market",
      expect.objectContaining({
        headers: { "X-Forwarded-Host": "market.example.cn" },
      }),
    );
  });

  it("requests listing collection and detail from the configured market", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(new Response(JSON.stringify({ items: [] })))
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ slug: "software-delivery-expert" })),
      );
    vi.stubGlobal("fetch", fetchMock);

    await listListings();
    await getListing("software-delivery-expert");

    expect(fetchMock.mock.calls[0][0]).toMatch(/do-worker-market\/listings$/);
    expect(fetchMock.mock.calls[1][0]).toMatch(
      /do-worker-market\/listings\/software-delivery-expert$/,
    );
  });
});
