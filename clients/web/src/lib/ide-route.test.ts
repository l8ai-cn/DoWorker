import { describe, expect, it } from "vitest";
import {
  pathnameHidesIdeSidebar,
  resolveActivityFromPathname,
} from "./ide-route";

describe("resolveActivityFromPathname", () => {
  it("resolves api-access for org slugs containing workspace", () => {
    expect(resolveActivityFromPathname("/admin-workspace/api-access")).toBe("apiAccess");
    expect(resolveActivityFromPathname("/dev-org/workspace")).toBe("workspace");
  });
});

describe("pathnameHidesIdeSidebar", () => {
  it("hides IDE sidebar on standalone dashboard pages", () => {
    expect(pathnameHidesIdeSidebar("/admin-workspace/api-access")).toBe(true);
    expect(pathnameHidesIdeSidebar("/dev-org/automation")).toBe(true);
    expect(pathnameHidesIdeSidebar("/dev-org/knowledge-base")).toBe(true);
    expect(pathnameHidesIdeSidebar("/dev-org/channels")).toBe(false);
  });
});
