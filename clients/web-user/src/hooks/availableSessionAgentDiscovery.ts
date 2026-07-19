import { authenticatedFetch } from "@/lib/identity";
import { displayNameForAgent } from "./availableAgentCatalog";
import type { AvailableAgent } from "./availableAgentTypes";

interface SessionListItemWire {
  id: string;
  agent_id?: string | null;
  agent_name?: string | null;
  created_at?: number | null;
}

interface AgentObjectWire {
  id: string;
  name: string;
  description?: string | null;
  harness?: string | null;
  skills?: { name: string; description: string }[];
}

export interface ScannedSessionAgent {
  agentId: string;
  agentName: string;
  sessionId: string;
  createdAt: number | null;
}

export async function scanSessionAgents(): Promise<ScannedSessionAgent[]> {
  const response = await authenticatedFetch("/v1/sessions?limit=100&kind=any");
  if (!response.ok) throw new Error(`${response.status} ${response.statusText}`);
  const body = (await response.json()) as { data: SessionListItemWire[] };
  const seen = new Map<string, ScannedSessionAgent>();
  for (const session of body.data) {
    if (!session.agent_id || !session.agent_name || seen.has(session.agent_id)) continue;
    seen.set(session.agent_id, {
      agentId: session.agent_id,
      agentName: session.agent_name,
      sessionId: session.id,
      createdAt: session.created_at ?? null,
    });
  }
  return Array.from(seen.values());
}

export async function enrichSessionAgent(scanned: ScannedSessionAgent): Promise<AvailableAgent> {
  const agent: AvailableAgent = {
    id: scanned.agentId,
    name: scanned.agentName,
    display_name: displayNameForAgent(scanned.agentName),
    description: null,
    harness: null,
    skills: [],
  };
  try {
    const response = await authenticatedFetch(
      `/v1/sessions/${encodeURIComponent(scanned.sessionId)}/agent`,
    );
    if (!response.ok) return agent;
    const json = (await response.json()) as AgentObjectWire;
    return {
      ...agent,
      display_name: displayNameForAgent(json.name, json.harness),
      description: json.description ?? null,
      harness: json.harness ?? null,
      skills: json.skills ?? [],
    };
  } catch {
    return agent;
  }
}
