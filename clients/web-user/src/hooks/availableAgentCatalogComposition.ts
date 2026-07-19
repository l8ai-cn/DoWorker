import { agentRootName } from "@/lib/forkHarness";
import { nativeCodingAgentForAvailableAgent } from "@/lib/nativeCodingAgents";
import { dedupeNativeAgents, fetchCatalogAgents } from "./availableAgentCatalog";
import { enrichSessionAgent, scanSessionAgents, type ScannedSessionAgent } from "./availableSessionAgentDiscovery";
import type { AvailableAgent } from "./availableAgentTypes";

interface AgentCandidate {
  recency: number;
  template: AvailableAgent | null;
  scanned: ScannedSessionAgent | null;
}

export async function fetchAvailableAgents(): Promise<AvailableAgent[]> {
  const [catalog, scanned] = await Promise.all([
    fetchCatalogAgents(),
    scanSessionAgents().catch(() => [] as ScannedSessionAgent[]),
  ]);
  const seeded = dedupeNativeAgents(catalog.filter((agent) => agent.builtin !== false));
  const templates = catalog.filter((agent) => agent.builtin === false);
  const catalogIds = new Set(catalog.map((agent) => agent.id));
  const seededNames = new Set(seeded.map((agent) => agentRootName(agent.name)));
  const hasKiroBuiltin = seeded.some(
    (agent) => nativeCodingAgentForAvailableAgent(agent)?.key === "kiro",
  );
  const candidates = new Map<string, AgentCandidate>();
  for (const template of templates) {
    const name = agentRootName(template.name);
    if (!seededNames.has(name)) {
      candidates.set(name, { recency: template.created_at ?? 0, template, scanned: null });
    }
  }
  for (const agent of scanned) {
    const name = agentRootName(agent.agentName);
    if (
      catalogIds.has(agent.agentId) ||
      seededNames.has(name) ||
      (hasKiroBuiltin && name.toLocaleLowerCase() === "kiro")
    ) {
      continue;
    }
    const recency = agent.createdAt ?? 0;
    const existing = candidates.get(name);
    if (!existing || recency > existing.recency) {
      candidates.set(name, { recency, template: null, scanned: agent });
    }
  }
  const resolved = await Promise.all(
    Array.from(candidates.values()).map((candidate) =>
      candidate.template === null
        ? enrichSessionAgent(candidate.scanned!)
        : Promise.resolve(candidate.template),
    ),
  );
  return [
    ...seeded,
    ...resolved
      .filter(
        (agent) =>
          nativeCodingAgentForAvailableAgent(agent)?.key !== "kiro" || !hasKiroBuiltin,
      )
      .sort((left, right) => (right.created_at ?? 0) - (left.created_at ?? 0)),
  ];
}
