import { describe, expect, it } from "vitest";
import {
  isHtmlWorkspacePath,
  normalizeWorkspaceFileSearch,
  parseWorkspaceFileParam,
} from "./workspace-file-param";

describe("parseWorkspaceFileParam", () => {
  it("reads a normal ?file= deep link", () => {
    const params = new URLSearchParams("file=gomoku%2Findex.html");
    expect(parseWorkspaceFileParam(params)).toBe("gomoku/index.html");
  });

  it("repairs a double-encoded ?file%3D…= link", () => {
    const params = new URLSearchParams("file%3Dgomoku%2Findex.html=");
    expect(parseWorkspaceFileParam(params)).toBe("gomoku/index.html");
  });

  it("strips a stray trailing equals from the value", () => {
    const params = new URLSearchParams("file=gomoku%2Findex.html=");
    expect(parseWorkspaceFileParam(params)).toBe("gomoku/index.html");
  });
});

describe("normalizeWorkspaceFileSearch", () => {
  it("rewrites malformed file query strings", () => {
    expect(normalizeWorkspaceFileSearch("?file%3Dgomoku%2Findex.html=")).toBe(
      "?file=gomoku%2Findex.html",
    );
  });

  it("preserves unrelated params", () => {
    expect(normalizeWorkspaceFileSearch("?debug=1&file%3Dfoo.txt=")).toBe(
      "?debug=1&file=foo.txt",
    );
  });
});

describe("isHtmlWorkspacePath", () => {
  it("detects html workspace paths", () => {
    expect(isHtmlWorkspacePath("gomoku/index.html")).toBe(true);
    expect(isHtmlWorkspacePath("README.md")).toBe(false);
  });
});
