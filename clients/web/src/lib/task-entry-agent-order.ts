import type { AgentData } from "@/lib/api";

const TASK_ENTRY_PREFERRED_SLUG = "do-agent";

export function sortAgentsForTaskEntry(agents: AgentData[]): AgentData[] {
  const preferred: AgentData[] = [];
  const rest: AgentData[] = [];
  for (const agent of agents) {
    if (agent.slug === TASK_ENTRY_PREFERRED_SLUG) {
      preferred.push(agent);
    } else {
      rest.push(agent);
    }
  }
  return [...preferred, ...rest];
}
