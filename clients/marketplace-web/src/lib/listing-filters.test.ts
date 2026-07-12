import { describe, expect, it } from "vitest";

import { filterListings, parseCatalogFilters } from "./listing-filters";
import type { ListingSummary } from "./marketplace-types";

const listings: ListingSummary[] = [
  {
    listing_id: "1",
    listing_version_id: "11",
    slug: "delivery",
    resource_type: "application",
    display_name: "软件交付专家",
    tagline: "完成可验证的软件交付",
    publisher: { slug: "do-worker", display_name: "Do Worker", verified: true },
    spaces: [{ slug: "software-delivery", name: "软件交付" }],
    published_at: "2026-07-11T08:00:00Z",
  },
  {
    listing_id: "2",
    listing_version_id: "12",
    slug: "mail-skill",
    resource_type: "skill",
    display_name: "邮件整理 Skill",
    tagline: "整理重要邮件",
    publisher: { slug: "mail-lab", display_name: "Mail Lab", verified: false },
    spaces: [{ slug: "office", name: "办公效率" }],
    published_at: "2026-07-10T08:00:00Z",
  },
];

describe("catalog filters", () => {
  it("parses only supported URL values", () => {
    expect(
      parseCatalogFilters({ q: "  交付 ", type: "application", space: "software-delivery" }),
    ).toEqual({ q: "交付", type: "application", space: "software-delivery" });
    expect(parseCatalogFilters({ type: "unknown" })).toEqual({
      q: "",
      type: "",
      space: "",
    });
  });

  it("filters by query, type, and space together", () => {
    expect(
      filterListings(listings, {
        q: "do worker",
        type: "application",
        space: "software-delivery",
      }),
    ).toEqual([listings[0]]);
  });
});
