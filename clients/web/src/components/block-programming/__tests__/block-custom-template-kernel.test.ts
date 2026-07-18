import { describe, expect, it } from "vitest";
import {
  expandBlockTemplate,
  extractBlockTemplateParameters,
  matchBlockTemplate,
} from "../block-custom-template-kernel";

describe("block custom template kernel", () => {
  it("extracts ordered unique parameters across templates", () => {
    expect(extractBlockTemplateParameters([
      "制作 {{topic}}",
      "test -f {{file}}",
      "{{file}} 存在",
    ])).toEqual(["topic", "file"]);
  });

  it("expands required parameters without accepting silent missing values", () => {
    expect(expandBlockTemplate("制作 {{topic}} 到 {{file}}", {
      topic: "季度复盘",
    })).toEqual({
      missing: ["file"],
      value: "制作 季度复盘 到 ",
    });
  });

  it("matches an expanded template back to parameter values", () => {
    expect(matchBlockTemplate("制作 {{topic}} 到 {{file}}", "制作 季度复盘 到 out.pptx"))
      .toEqual({ topic: "季度复盘", file: "out.pptx" });
  });
});
