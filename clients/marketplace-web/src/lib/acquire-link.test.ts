import { describe, expect, it } from "vitest";

import { buildAcquireLink } from "./acquire-link";

describe("acquire link", () => {
  it("builds a Core Web acquisition URL with stable identifiers", () => {
    expect(
      buildAcquireLink("https://app.l8ai.cn/", {
        market: "do-worker-market",
        listing: "software-delivery-expert",
        version: "301",
      }),
    ).toBe(
      "https://app.l8ai.cn/marketplace/acquire?market=do-worker-market&listing=software-delivery-expert&version=301",
    );
  });

  it("returns null when Core Web is not configured", () => {
    expect(
      buildAcquireLink(undefined, {
        market: "do-worker-market",
        listing: "software-delivery-expert",
        version: "301",
      }),
    ).toBeNull();
  });
});
