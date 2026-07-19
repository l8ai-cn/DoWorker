import { describe, expect, it } from "vitest";
import { ensureBlockNodeId } from "../loop-node-identity";

describe("Loop block node identity", () => {
  it("removes every boundary separator from Blockly-generated ids", () => {
    const block = {
      data: "",
      id: "--nzo_ch4rdt",
      type: "loop_custom_ppt_step_v1",
      getFieldValue: () => null,
    };

    expect(ensureBlockNodeId(block as never))
      .toBe("n-loop-custom-ppt-step-v1-nzo-ch4rdt");
  });

  it("normalizes a separator exposed by suffix truncation", () => {
    const block = {
      data: "",
      id: "abcdefghijk--17k_gw01_xn",
      type: "loop_custom_ppt_step_v1",
      getFieldValue: () => null,
    };

    expect(ensureBlockNodeId(block as never))
      .toBe("n-loop-custom-ppt-step-v1-17k-gw01-xn");
  });
});
