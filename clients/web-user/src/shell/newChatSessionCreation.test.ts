import { beforeEach, describe, expect, it, vi } from "vitest";
import type { AvailableAgent } from "@/hooks/useAvailableAgents";

vi.mock("@/lib/workerSessionMutations", () => ({ createWorkerSession: vi.fn() }));

import { createWorkerSession } from "@/lib/workerSessionMutations";
import {
  createNewChatSession,
  newChatCreateDisabledReason,
  newChatInteractionMode,
} from "./newChatSessionCreation";

const createMock = vi.mocked(createWorkerSession);

function agent(overrides: Partial<AvailableAgent> = {}): AvailableAgent {
  return {
    id: overrides.id ?? "codex-cli",
    workerTypeSlug: overrides.workerTypeSlug ?? overrides.id ?? "codex-cli",
    supportedModes: overrides.supportedModes ?? ["acp", "pty"],
    requiresModelResource: overrides.requiresModelResource ?? false,
    name: overrides.name ?? "codex-cli",
    display_name: overrides.display_name ?? "Codex",
    description: overrides.description ?? null,
    harness: overrides.harness ?? null,
    skills: overrides.skills ?? [],
    builtin: overrides.builtin,
    created_at: overrides.created_at,
  };
}

function input(overrides: Partial<Parameters<typeof createNewChatSession>[0]> = {}) {
  return {
    agent: agent(),
    hostId: "host_1",
    workspace: "/repo",
    sandboxSelected: false,
    sandboxRepoUrl: "",
    sandboxRepoBranch: "",
    branchName: "",
    modelResourceId: null,
    tokenBudget: null,
    ...overrides,
  };
}

describe("newChatSessionCreation", () => {
  beforeEach(() => {
    createMock.mockReset();
    createMock.mockResolvedValue({ id: "conv_new" });
  });

  it("creates a normal Agent through the authoritative Worker mutation", async () => {
    await expect(
      createNewChatSession(
        input({ agent: agent({ requiresModelResource: true }), modelResourceId: 9, tokenBudget: 12000 }),
      ),
    ).resolves.toEqual({ id: "conv_new" });

    expect(createMock).toHaveBeenCalledWith({
      agentId: "codex-cli",
      workerTypeSlug: "codex-cli",
      supportedModes: ["acp", "pty"],
      requiresModelResource: true,
      initialItems: [],
      mode: "acp",
      hostId: "host_1",
      workspace: "/repo",
      modelResourceId: 9,
      tokenBudget: 12000,
    });
  });

  it("uses PTY mode for native terminal Workers", async () => {
    const native = agent({
      id: "codex-native-ui",
      name: "codex-native-ui",
      display_name: "Codex",
      harness: "codex-native",
    });
    expect(newChatInteractionMode(native)).toBe("pty");
    await createNewChatSession(input({ agent: native }));
    expect(createMock).toHaveBeenCalledWith(expect.objectContaining({ mode: "pty" }));
  });

  it.each([
    [
      "metadata",
      { agent: { ...agent(), workerTypeSlug: undefined } },
      "missing Worker creation metadata",
    ],
    ["mode", { agent: agent({ supportedModes: ["pty"] }) }, "does not support ACP sessions"],
    ["repo", { sandboxSelected: true, sandboxRepoUrl: "https://github.com/o/r" }, "workspace option"],
    ["branch", { branchName: "feature/x" }, "Git worktree starts"],
  ])("blocks unsupported %s configuration before POST", async (_name, patch, message) => {
    const value = input(patch as Partial<Parameters<typeof createNewChatSession>[0]>);
    expect(newChatCreateDisabledReason(value)).toContain(message);
    await expect(createNewChatSession(value)).rejects.toThrow(message);
    expect(createMock).not.toHaveBeenCalled();
  });
});
