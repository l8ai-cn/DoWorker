import { describe, expect, it } from "vitest";
import type { AvailableAgent } from "@/hooks/useAvailableAgents";
import { agentUsesWorkerModelPool } from "./workerModelPool";

function agent(partial: Partial<AvailableAgent> & Pick<AvailableAgent, "id">): AvailableAgent {
  return {
    name: partial.id,
    display_name: partial.id,
    description: null,
    harness: null,
    skills: [],
    ...partial,
  };
}

describe("agentUsesWorkerModelPool", () => {
  it("includes do-agent and env-mount CLI harness slugs", () => {
    expect(agentUsesWorkerModelPool(agent({ id: "do-agent", harness: "do-agent" }))).toBe(true);
    expect(agentUsesWorkerModelPool(agent({ id: "codex-cli", harness: "codex" }))).toBe(true);
    expect(agentUsesWorkerModelPool(agent({ id: "claude-code", harness: "claude-code" }))).toBe(true);
    expect(agentUsesWorkerModelPool(agent({ id: "gemini-cli", harness: "gemini-cli" }))).toBe(true);
  });

  it("excludes native-terminal wrappers and unrelated agents", () => {
    expect(
      agentUsesWorkerModelPool(agent({ id: "codex-native-ui", harness: "codex-native" })),
    ).toBe(false);
    expect(agentUsesWorkerModelPool(agent({ id: "claude-sdk", harness: "claude-sdk" }))).toBe(false);
    expect(agentUsesWorkerModelPool(null)).toBe(false);
  });
});
