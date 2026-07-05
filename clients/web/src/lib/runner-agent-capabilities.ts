import type { AgentData } from "@/lib/api";

type RunnerCapability = Pick<AgentData, "available_agents"> & {
  capabilities?: Record<string, string>;
};

export type CapabilityAxis =
  | "resume"
  | "permission"
  | "usage"
  | "control"
  | "interrupt"
  | "streaming"
  | "subagents"
  | "model_family";

export function agentSupports(
  agent: Pick<AgentData, "capabilities"> | null | undefined,
  axis: CapabilityAxis,
  value?: string,
): boolean {
  const caps = agent?.capabilities;
  if (!caps) return false;
  const v = caps[axis];
  if (v === undefined) return false;
  if (value === undefined) return true;
  if (axis === "control") {
    return v.split(",").map((t) => t.trim()).includes(value);
  }
  return v.toLowerCase() === value.toLowerCase();
}

export function runnerSupportsAgent(
  runner: RunnerCapability | null | undefined,
  agentSlug: string | null | undefined,
): boolean {
  if (!agentSlug) return false;
  return runner?.available_agents?.includes(agentSlug) ?? false;
}

export function runnersSupportingAgent<T extends RunnerCapability>(
  runners: T[],
  agentSlug: string | null | undefined,
): T[] {
  if (!agentSlug) return runners;
  return runners.filter((runner) => runnerSupportsAgent(runner, agentSlug));
}

export function hasRunnerForAgent(
  runners: RunnerCapability[],
  agentSlug: string | null | undefined,
): boolean {
  return runnersSupportingAgent(runners, agentSlug).length > 0;
}

export function agentsSupportedByRunners<T extends Pick<AgentData, "slug">>(
  agents: T[],
  runners: RunnerCapability[],
): T[] {
  const supportedSlugs = new Set(runners.flatMap((runner) => runner.available_agents ?? []));
  if (supportedSlugs.size === 0) return [];
  return agents.filter((agent) => supportedSlugs.has(agent.slug));
}
