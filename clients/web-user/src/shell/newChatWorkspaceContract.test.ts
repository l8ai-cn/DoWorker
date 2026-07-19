import { describe, expect, it } from "vitest";
import { describeCreateError } from "./newChatCreateError";
import {
  composeSandboxWorkspace,
  deriveHomeDir,
  deriveRepoName,
  isValidSandboxRepoUrl,
  isValidWorkspace,
  matchSkillInvocation,
  normalizeWorkspacePath,
  sessionsSharingDirectory,
} from "./newChatWorkspace";

describe("new chat workspace contract", () => {
  it.each([
    ["/Users/me/repo", true],
    [" /Users/me/repo ", true],
    ["/", true],
    ["~/repo", false],
    ["repo", false],
    ["", false],
  ])("validates workspace %s", (value, expected) => {
    expect(isValidWorkspace(value)).toBe(expected);
  });

  it.each([
    ["/Users/me/repo/", "/Users/me/repo"],
    [" /a/b  ", "/a/b"],
    ["/", "/"],
    ["", null],
  ])("normalizes workspace %s", (value, expected) => {
    expect(normalizeWorkspacePath(value)).toBe(expected);
  });

  it("validates and composes sandbox repository input", () => {
    expect(isValidSandboxRepoUrl("https://github.com/org/repo")).toBe(true);
    expect(isValidSandboxRepoUrl("org/repo")).toBe(false);
    expect(deriveRepoName("https://github.com/org/repo.git")).toBe("repo");
    expect(composeSandboxWorkspace("https://github.com/org/repo", "release-1.2")).toBe(
      "https://github.com/org/repo#release-1.2",
    );
  });

  it("derives a home directory and detects occupied workspaces", () => {
    expect(deriveHomeDir([{ path: "/Users/alice/projects", type: "directory" }])).toBe("/Users/alice");
    expect(sessionsSharingDirectory(
      [{ id: "one", host_id: "h1", workspace: "/Users/alice/repo", runner_id: "r1", status: "idle" }],
      "h1",
      "/Users/alice/repo/",
      () => true,
    ).map((session) => session.id)).toEqual(["one"]);
  });

  it("extracts a bundled skill invocation without treating file paths as commands", () => {
    const skills = [{ name: "review-pr", description: "Review a pull request" }];
    expect(matchSkillInvocation("/review-pr 123", skills)).toEqual({ name: "review-pr", args: "123" });
    expect(matchSkillInvocation("/etc/hosts", skills)).toBeNull();
  });

  it("uses the server reason when create fails", async () => {
    expect(await describeCreateError({
      status: 409,
      json: async () => ({ detail: "host is offline" }),
    } as Response)).toBe("host is offline");
  });
});
