import { describe, expect, it } from "vitest";
import {
  hasBlockCustomDefinition,
  registerBlockCustomDefinition,
} from "../block-custom-definition-registry";

describe("block custom definition registry", () => {
  const definitions = [{ slug: "ppt-step", label: "Professional PPT" }];

  it("registers a new definition without changing existing entries", () => {
    expect(registerBlockCustomDefinition(definitions, {
      slug: "report-step",
      label: "Report",
    })).toEqual({
      definitions: [
        { slug: "ppt-step", label: "Professional PPT" },
        { slug: "report-step", label: "Report" },
      ],
      registered: true,
    });
  });

  it("rejects duplicate slugs instead of silently replacing a definition", () => {
    expect(hasBlockCustomDefinition(definitions, "ppt-step")).toBe(true);
    expect(registerBlockCustomDefinition(definitions, {
      slug: "ppt-step",
      label: "Changed PPT",
    })).toEqual({
      definitions,
      registered: false,
    });
  });
});
