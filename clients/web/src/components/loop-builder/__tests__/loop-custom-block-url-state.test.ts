import { describe, expect, it } from "vitest";
import {
  readLoopCustomBlocksFromHash,
  writeLoopCustomBlocksToHash,
} from "../loop-custom-block-url-state";

describe("loop custom block URL state", () => {
  it("round-trips draft custom block definitions through the URL hash", () => {
    const hash = writeLoopCustomBlocksToHash("", [{
      slug: "ppt-step",
      version: 1,
      label: "专业 PPT",
      parameters: ["topic", "file"],
      expansion: {
        agentLocalId: "ppt-step-task",
        verifierLocalId: "ppt-step-check",
        promptTemplate: "制作 {{topic}} 的专业 PPT",
        commandTemplate: "test -f {{file}}",
        acceptTemplate: "{{file}} 存在且可打开",
      },
    }]);

    expect(hash).toContain("loopCustomBlocks=");
    expect(readLoopCustomBlocksFromHash(hash)).toEqual([expect.objectContaining({
      label: "专业 PPT",
      parameters: ["topic", "file"],
      slug: "ppt-step",
      version: 1,
    })]);
  });

  it("rejects malformed or duplicate URL state", () => {
    const duplicate = encodeURIComponent(JSON.stringify([
      { slug: "bad_slug", label: "坏", promptTemplate: "x", commandTemplate: "x", acceptTemplate: "x" },
      { slug: "ppt-step", label: "A", promptTemplate: "{{x}}", commandTemplate: "c", acceptTemplate: "a" },
      { slug: "ppt-step", label: "B", promptTemplate: "x", commandTemplate: "x", acceptTemplate: "x" },
    ]));

    expect(readLoopCustomBlocksFromHash("not-json")).toEqual([]);
    expect(readLoopCustomBlocksFromHash(`#loopCustomBlocks=${duplicate}`))
      .toHaveLength(1);
  });
});
