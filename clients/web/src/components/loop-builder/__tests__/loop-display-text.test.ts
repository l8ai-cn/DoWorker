import { describe, expect, it } from "vitest";
import { loopParseStatusLabel, loopRunStatusLabel } from "../loop-display-text";

describe("Loop display text", () => {
  it("maps protocol states to Chinese labels", () => {
    expect(loopParseStatusLabel("valid")).toBe("有效");
    expect(loopParseStatusLabel("parsing")).toBe("校验中");
    expect(loopParseStatusLabel("syntax-error")).toBe("存在错误");
    expect(loopRunStatusLabel("active")).toBe("运行中");
    expect(loopRunStatusLabel("completed")).toBe("已完成");
  });
});
