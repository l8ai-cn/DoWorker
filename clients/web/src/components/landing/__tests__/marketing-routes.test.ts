import { describe, expect, it } from "vitest";

import { isMarketingRouteActive, marketingRoutes } from "../marketing-routes";

describe("marketing routes", () => {
  it("exposes one dedicated page per first-level product menu", () => {
    expect(marketingRoutes.map(({ id, href }) => ({ id, href }))).toEqual([
      { id: "home", href: "/" },
      { id: "product", href: "/product" },
      { id: "solutions", href: "/solutions" },
      { id: "marketplace", href: "/marketplace" },
      { id: "docs", href: "/docs" },
    ]);
  });

  it("keeps nested product routes active without activating home", () => {
    expect(isMarketingRouteActive("/product/runtime", "/product")).toBe(true);
    expect(isMarketingRouteActive("/product", "/")).toBe(false);
  });
});
