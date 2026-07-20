import { describe, expect, it } from "vitest";

import { shouldRenderAgentPanel } from "../SplitTreeRenderer";

describe("workspace pane routing", () => {
  it("uses the Agent Workbench for a video studio PTY Worker", () => {
    expect(shouldRenderAgentPanel("pty", "video-studio")).toBe(true);
  });

  it("keeps ordinary PTY Workers in the terminal pane", () => {
    expect(shouldRenderAgentPanel("pty", "codex-cli")).toBe(false);
  });
});
