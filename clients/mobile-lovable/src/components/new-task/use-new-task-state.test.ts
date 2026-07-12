import { describe, expect, it } from "vitest";
import { resolveTaskInteractionMode } from "./use-new-task-state";

describe("resolveTaskInteractionMode", () => {
  it("keeps ACP for a Worker that supports visual conversations", () => {
    expect(
      resolveTaskInteractionMode(
        [
          {
            id: "codex-cli",
            name: "Codex",
            vendor: "OpenAI",
            avatar: "O",
            desc: "",
            supportedModes: ["acp", "pty"],
          },
        ],
        "codex-cli",
        "acp",
      ),
    ).toBe("acp");
  });

  it("projects an unsupported ACP choice to the Worker's PTY mode", () => {
    expect(
      resolveTaskInteractionMode(
        [
          {
            id: "aider",
            name: "Aider",
            vendor: "Aider",
            avatar: "A",
            desc: "",
            supportedModes: ["pty"],
          },
        ],
        "aider",
        "acp",
      ),
    ).toBe("pty");
  });
});
