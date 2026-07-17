import { act, renderHook } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import {
  availableAcpExperts,
  resolveTaskInteractionMode,
  useNewTaskState,
} from "./use-new-task-state";

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
            requiresModelResource: true,
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
            requiresModelResource: false,
          },
        ],
        "aider",
        "acp",
      ),
    ).toBe("pty");
  });

  it("only exposes ACP experts backed by an available ACP Worker", () => {
    expect(
      availableAcpExperts(
        [
          {
            slug: "codex-reviewer",
            name: "Codex Reviewer",
            description: null,
            agent_slug: "codex-cli",
            interaction_mode: "acp",
            run_count: 0,
            last_run_at: null,
            prompt: null,
          },
          {
            slug: "pty-operator",
            name: "PTY Operator",
            description: null,
            agent_slug: "codex-cli",
            interaction_mode: "pty",
            run_count: 0,
            last_run_at: null,
            prompt: null,
          },
          {
            slug: "offline-claude",
            name: "Offline Claude",
            description: null,
            agent_slug: "claude-code",
            interaction_mode: "acp",
            run_count: 0,
            last_run_at: null,
            prompt: null,
          },
        ],
        [
          {
            id: "codex-cli",
            name: "Codex",
            vendor: "OpenAI",
            avatar: "O",
            desc: "",
            supportedModes: ["acp", "pty"],
            requiresModelResource: true,
          },
        ],
      ).map((expert) => expert.slug),
    ).toEqual(["codex-reviewer"]);
  });

  it("clears an expert when the user manually selects a Worker", () => {
    const { result } = renderHook(() =>
      useNewTaskState({
        search: { expert: "codex-reviewer" },
        engines: [
          {
            id: "codex-cli",
            name: "Codex",
            vendor: "OpenAI",
            avatar: "O",
            desc: "",
            supportedModes: ["acp", "pty"],
            requiresModelResource: true,
          },
          {
            id: "gemini-cli",
            name: "Gemini",
            vendor: "Google",
            avatar: "G",
            desc: "",
            supportedModes: ["acp"],
            requiresModelResource: true,
          },
        ],
        projectNames: ["默认项目"],
        initialProjectName: "默认项目",
      }),
    );

    act(() => result.current.selectWorker("gemini-cli"));

    expect(result.current.engineID).toBe("gemini-cli");
    expect(result.current.expertSlug).toBeUndefined();
  });
});
