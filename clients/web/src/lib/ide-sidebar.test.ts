import { describe, expect, it } from "vitest";
import { activityHasSidebar } from "./ide-sidebar";

describe("activityHasSidebar", () => {
  it("returns false for standalone pages without IDE sidebar content", () => {
    expect(activityHasSidebar("apiAccess")).toBe(false);
    expect(activityHasSidebar("knowledge")).toBe(false);
    expect(activityHasSidebar("automation")).toBe(false);
  });

  it("returns true for activities with sidebar panels", () => {
    expect(activityHasSidebar("workspace")).toBe(true);
    expect(activityHasSidebar("channels")).toBe(true);
    expect(activityHasSidebar("infra")).toBe(true);
  });
});
