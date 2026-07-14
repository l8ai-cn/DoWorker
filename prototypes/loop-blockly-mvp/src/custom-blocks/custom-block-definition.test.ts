import { describe, expect, it } from "vitest";

import {
  createCustomBlockDefinition,
  expandCustomBlockTemplate,
} from "./custom-block-definition";

describe("createCustomBlockDefinition", () => {
  it("extracts unique parameters in first-use order", () => {
    const result = createCustomBlockDefinition({
      id: "fix-file",
      name: "修复文件",
      template: "修复 {{file-path}}，运行 {{test-command}}，复查 {{file-path}}。",
    });

    expect(result).toEqual({
      definition: {
        id: "fix-file",
        name: "修复文件",
        template: "修复 {{file-path}}，运行 {{test-command}}，复查 {{file-path}}。",
        parameters: ["file-path", "test-command"],
      },
      errors: [],
    });
  });

  it.each([
    ["bad id", { id: "Bad_ID" }, "积木 ID"],
    ["blank name", { name: " " }, "名称"],
    ["blank template", { template: "" }, "模板"],
    [
      "invalid parameter",
      { template: "处理 {{File_Path}}" },
      "参数",
    ],
  ])("rejects %s", (_label, override, message) => {
    const result = createCustomBlockDefinition({
      id: "valid-id",
      name: "有效积木",
      template: "处理 {{file-path}}",
      ...override,
    });

    expect(result.definition).toBeUndefined();
    expect(result.errors.join(" ")).toContain(message);
  });
});

describe("expandCustomBlockTemplate", () => {
  it("expands repeated parameters deterministically", () => {
    const result = createCustomBlockDefinition({
      id: "fix-file",
      name: "修复文件",
      template: "修复 {{file-path}}，运行 {{test-command}}，复查 {{file-path}}。",
    });

    expect(expandCustomBlockTemplate(result.definition!, {
      "file-path": "src/cart.ts",
      "test-command": "pnpm test",
    })).toEqual({
      value: "修复 src/cart.ts，运行 pnpm test，复查 src/cart.ts。",
      missingParameters: [],
    });
  });

  it("rejects missing values instead of substituting empty text", () => {
    const result = createCustomBlockDefinition({
      id: "run-command",
      name: "运行命令",
      template: "执行 {{command-name}}",
    });

    expect(expandCustomBlockTemplate(result.definition!, {
      "command-name": " ",
    })).toEqual({
      missingParameters: ["command-name"],
    });
  });
});
