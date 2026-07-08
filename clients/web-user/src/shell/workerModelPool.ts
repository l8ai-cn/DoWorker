import type { AvailableAgent } from "@/hooks/useAvailableAgents";

const ENV_MOUNT_AGENT_SLUGS = new Set(["codex-cli", "claude-code", "gemini-cli"]);

/** Agents whose create path mounts an org model pool row (mirrors backend HarnessMountKindFor). */
export function agentUsesWorkerModelPool(agent: AvailableAgent | null | undefined): boolean {
  if (!agent) return false;
  if (agent.harness === "do-agent") return true;
  return ENV_MOUNT_AGENT_SLUGS.has(agent.id);
}
