import { describe, expect, it } from "vitest";

import { collectTaxonomyTags } from "./catalog-filter-options";

describe("collectTaxonomyTags", () => {
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
});
