import { beforeEach, describe, expect, it } from "vitest";
import { sanitizeReturnTo } from "@/lib/auth-return-to";

const ORIGIN = "https://app.example.com";

describe("sanitizeReturnTo", () => {
  beforeEach(() => {
    Object.defineProperty(window, "location", {
      configurable: true,
      value: { origin: ORIGIN },
    });
  });

  it("replaces a backslash protocol-relative payload with the safe default", () => {
    expect(sanitizeReturnTo("/\\evil.com")).toBe("/");
  });

  it.each([
    ["//evil.com", "protocol-relative"],
    ["/\\evil.com", "backslash variant"],
    ["https://evil.com", "absolute off-origin URL"],
    ["\\\\evil.com", "double backslash"],
  ])("rejects %s (%s) → '/'", (payload) => {
    expect(sanitizeReturnTo(payload)).toBe("/");
  });

  it("preserves a legitimate same-origin path with query and fragment", () => {
    expect(sanitizeReturnTo("/sessions/abc?tab=logs#top")).toBe("/sessions/abc?tab=logs#top");
  });

  it("repairs a malformed ?file= deep link embedded in return_to", () => {
    expect(sanitizeReturnTo("/c/conv_abc?file%3Dgomoku%2Findex.html=")).toBe(
      "/c/conv_abc?file=gomoku%2Findex.html",
    );
  });
});
