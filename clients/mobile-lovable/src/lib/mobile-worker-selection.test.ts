import { describe, expect, it } from "vitest";
import type { AgentPickerOption } from "./agent-display";
import { resolveMobileWorkerSelection } from "./mobile-worker-selection";

const codex: AgentPickerOption = {
  id: "codex-cli",
  name: "Codex",
  vendor: "OpenAI",
  avatar: "O",
  desc: "codex-cli",
  supportedModes: ["acp", "pty"],
  requiresModelResource: true,
};

const claude: AgentPickerOption = {
  id: "claude-code",
  name: "Claude Code",
  vendor: "Anthropic",
  avatar: "A",
  desc: "claude-code",
  supportedModes: ["pty"],
  requiresModelResource: true,
};

describe("resolveMobileWorkerSelection", () => {
  it("does not present a missing Worker error before authentication", () => {
    expect(resolveMobileWorkerSelection([], "codex-cli", false, false, null)).toEqual({
      kind: "unauthenticated",
      current: null,
      message: "登录后加载可用 Worker。",
    });
  });

  it("waits for the real Worker catalog instead of exposing a fallback", () => {
    expect(resolveMobileWorkerSelection([], "codex-cli", true, true, null)).toEqual({
      kind: "loading",
      current: null,
      message: "正在加载可用 Worker…",
    });
  });

  it("reports an unavailable catalog instead of fabricating an Agent", () => {
    expect(resolveMobileWorkerSelection([], "codex-cli", true, false, "HTTP 503")).toEqual({
      kind: "error",
      current: null,
      message: "无法加载可用 Worker：HTTP 503",
    });
    expect(resolveMobileWorkerSelection([], "codex-cli", true, false, null).kind).toBe("empty");
  });

  it("only selects an Agent returned by the backend catalog", () => {
    expect(
      resolveMobileWorkerSelection([codex, claude], "claude-code", true, false, null),
    ).toMatchObject({
      kind: "ready",
      current: claude,
    });
    expect(resolveMobileWorkerSelection([claude], "codex-cli", true, false, null)).toMatchObject({
      kind: "ready",
      current: claude,
    });
  });
});
