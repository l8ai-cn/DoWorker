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
    expect(hideIdeChrome("/acme/mobile/pods/pod-1")).toBe(true);
    expect(hideIdeChrome("/acme/mobile/workers")).toBe(false);
    expect(hideIdeChrome("/acme/mobile/workers/pod-1")).toBe(true);
    expect(hideIdeChrome("/acme/workflows/my-workflow")).toBe(false);
  });

  it("hides mobile tab bar when sidebar or chrome is hidden", () => {
    expect(hideMobileTabBar("/settings/git")).toBe(true);
    expect(hideMobileTabBar("/acme/do-agent/x")).toBe(true);
    expect(hideMobileTabBar("/acme/channels")).toBe(false);
    expect(hideMobileTabBar("/acme/mobile/workers")).toBe(false);
    expect(hideMobileTabBar("/acme/mobile/workers/")).toBe(false);
    expect(hideMobileTabBar("/acme/mobile/workers/pod-1")).toBe(true);
  });
});
