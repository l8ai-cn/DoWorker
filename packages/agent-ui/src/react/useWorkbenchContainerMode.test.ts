import { describe, expect, it } from "vitest";

import { workbenchContainerMode } from "./useWorkbenchContainerMode";

describe("workbenchContainerMode", () => {
  it("uses container width instead of viewport width", () => {
    expect(workbenchContainerMode(639)).toBe("narrow");
    expect(workbenchContainerMode(640)).toBe("medium");
    expect(workbenchContainerMode(959)).toBe("medium");
    expect(workbenchContainerMode(960)).toBe("wide");
  });
});
