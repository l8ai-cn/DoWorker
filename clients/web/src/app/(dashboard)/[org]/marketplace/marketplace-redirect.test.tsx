import { describe, expect, it, vi } from "vitest";

import MarketplaceDetailRoute from "./[listingSlug]/page";
import MarketplaceRoute from "./page";

const { redirect } = vi.hoisted(() => ({ redirect: vi.fn() }));

vi.mock("next/navigation", () => ({ redirect }));

vi.mock("@/components/marketplace/MarketplaceCatalogPage", () => ({
  MarketplaceCatalogPage: () => null,
}));

vi.mock("@/components/marketplace/MarketplaceDetailPage", () => ({
  MarketplaceDetailPage: () => null,
}));

describe("organization marketplace canonical redirects", () => {
  it("sends the legacy organization market route to the public marketplace", async () => {
    await MarketplaceRoute({ params: Promise.resolve({ org: "dev-org" }) });

    expect(redirect).toHaveBeenCalledWith("https://market.l8ai.cn");
  });

  it("sends a legacy organization listing route to the public app detail", async () => {
    await MarketplaceDetailRoute({
      params: Promise.resolve({ org: "dev-org", listingSlug: "software-delivery-expert" }),
    });

    expect(redirect).toHaveBeenCalledWith(
      "https://market.l8ai.cn/apps/software-delivery-expert",
    );
  });
});
