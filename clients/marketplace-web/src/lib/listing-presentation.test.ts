import { describe, expect, it } from "vitest";

import { formatQuotaSummary } from "./listing-presentation";

describe("formatQuotaSummary", () => {
  it("formats micro credits without losing precision", () => {
    expect(
      formatQuotaSummary({
        mode: "per_install",
        estimated_credits_micro: "20500000",
      }),
    ).toBe("启用需 20.5 市场额度");
  });

  it("does not invent quota information for invalid data", () => {
    expect(
      formatQuotaSummary({
        mode: "per_install",
        estimated_credits_micro: "invalid",
      }),
    ).toBeNull();
  });
});
