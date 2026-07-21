import { afterEach, describe, expect, it, vi } from "vitest";

import {
  getListing,
  getMarket,
  listListings,
} from "./marketplace-api";

describe("marketplace api", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.unstubAllEnvs();
  });

  it("uses the internal API URL and configured request host", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(JSON.stringify({ slug: "agent-cloud-market" })),
    );
    vi.stubGlobal("fetch", fetchMock);
    vi.stubEnv("MARKETPLACE_API_INTERNAL_URL", "http://marketplace-api:8080/");
    vi.stubEnv("MARKETPLACE_REQUEST_HOST", "market.example.cn");

    await getMarket();

    expect(fetchMock).toHaveBeenCalledWith(
      "http://marketplace-api:8080/api/marketplace/v1/markets/agent-cloud-market",
      expect.objectContaining({
        headers: { "X-Forwarded-Host": "market.example.cn" },
      }),
    );
  });

  it("sends the complete server-side filter contract to the listing collection", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(new Response(JSON.stringify({ items: [] })))
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ slug: "software-delivery-expert" })),
      );
    vi.stubGlobal("fetch", fetchMock);

    await listListings({
      q: "交付",
      scene: "software-delivery",
      industry: "enterprise-services",
      audience: "engineering-lead",
      capability: "e2e-testing",
      type: "application",
      integration: "github",
      readiness: "runner-required",
      space: "software-delivery",
      sort: "latest",
    });
    await getListing("software-delivery-expert");

    expect(fetchMock.mock.calls[0][0]).toBe(
      "http://marketplace:8080/api/marketplace/v1/markets/agent-cloud-market/listings?q=%E4%BA%A4%E4%BB%98&scene=software-delivery&industry=enterprise-services&audience=engineering-lead&capability=e2e-testing&type=application&integration=github&readiness=runner-required&space=software-delivery&sort=latest",
    );
    expect(fetchMock.mock.calls[1][0]).toMatch(
      /agent-cloud-market\/listings\/software-delivery-expert$/,
    );
  });
});
