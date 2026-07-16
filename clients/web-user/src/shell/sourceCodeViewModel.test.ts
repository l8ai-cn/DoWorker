import { describe, expect, it } from "vitest";
import { buildSourceLineModels } from "./sourceCodeViewModel";

describe("buildSourceLineModels", () => {
  it("keeps a line highlighted when a selection only covers its newline", () => {
    const [line] = buildSourceLineModels(
      ["ab", "cd"],
      null,
      [],
      { start_index: 2, end_index: 3, anchor_content: "\n" },
      "",
      undefined,
    );

    expect(line.selectionOverlaps).toBe(true);
  });
});
