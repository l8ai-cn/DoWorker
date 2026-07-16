import { describe, expect, it } from "vitest";

import {
  chineseToolActivityGroupSummary,
  englishToolActivityGroupSummary,
} from "./toolActivityGroupText";

const counts = [
  { count: 2, label: "Command" },
  { count: 1, label: "Read file" },
  { count: 1, label: "Browser" },
];

describe("tool activity group text", () => {
  it("preserves category order in English", () => {
    expect(englishToolActivityGroupSummary(counts)).toBe(
      "Ran 2 commands · Read 1 file · Used browser 1 time",
    );
  });

  it("preserves category order in Chinese", () => {
    expect(chineseToolActivityGroupSummary(counts)).toBe(
      "运行了 2 个命令 · 读取了 1 个文件 · 使用浏览器 1 次",
    );
  });
});
