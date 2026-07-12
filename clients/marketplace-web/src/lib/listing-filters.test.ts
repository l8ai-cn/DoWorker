import { describe, expect, it } from "vitest";

import { parseCatalogFilters } from "./listing-filters";

describe("catalog filters", () => {
  it("parses every supported marketplace filter from the URL", () => {
    expect(
      parseCatalogFilters({
        q: "  交付 ",
        scene: "software-delivery",
        industry: "enterprise-services",
        audience: "engineering-lead",
        capability: "e2e-testing",
        type: "application",
        integration: "github",
        readiness: "runner-required",
        space: "software-delivery",
        sort: "latest",
      }),
    ).toEqual({
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
    expect(parseCatalogFilters({ type: "unknown" })).toEqual({
      q: "",
      scene: "",
      industry: "",
      audience: "",
      capability: "",
      type: "",
      integration: "",
      readiness: "",
      space: "",
      sort: "featured",
    });
  });
});
