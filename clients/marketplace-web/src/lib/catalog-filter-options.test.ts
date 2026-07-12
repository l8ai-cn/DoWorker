import { describe, expect, it } from "vitest";

import {
  collectTaxonomyTags,
  taxonomyFilterGroups,
} from "./catalog-filter-options";

describe("catalog taxonomy filters", () => {
  it("keeps the selected taxonomy value clearable when it is absent from the page", () => {
    expect(
      collectTaxonomyTags(
        [],
        "capability",
        { slug: "e2e-testing", display_name: "e2e-testing", kind: "capability" },
      ),
    ).toEqual([
      { slug: "e2e-testing", display_name: "e2e-testing", kind: "capability" },
    ]);
  });

  it("exposes capability as a catalog taxonomy filter", () => {
    expect(taxonomyFilterGroups).toContainEqual({
      key: "capability",
      label: "核心能力",
    });
  });
});
