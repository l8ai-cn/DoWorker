import { describe, expect, it } from "vitest";
import { hideIdeChrome, hideIdeSidebar, hideMobileTabBar } from "./ide-chrome";

describe("ide-chrome", () => {
  it("hides sidebar on personal settings, support, and standalone pages", () => {
    expect(hideIdeSidebar("/settings/general")).toBe(true);
    expect(hideIdeSidebar("/support/abc")).toBe(true);
    expect(hideIdeSidebar("/admin-workspace/api-access")).toBe(true);
    expect(hideIdeSidebar("/acme/workspace")).toBe(false);
  });

  it("hides full chrome on agent consoles", () => {
    expect(hideIdeChrome("/acme/do-agent/pod-1")).toBe(true);
    expect(hideIdeChrome("/acme/loopal/pod-1")).toBe(true);
    expect(hideIdeChrome("/acme/loops/my-loop")).toBe(false);
  });

  it("hides mobile tab bar when sidebar or chrome is hidden", () => {
    expect(hideMobileTabBar("/settings/git")).toBe(true);
    expect(hideMobileTabBar("/acme/do-agent/x")).toBe(true);
    expect(hideMobileTabBar("/acme/channels")).toBe(false);
  });
});
