import { capitalizeAgentName } from "@/lib/agentLabels";
import {
  nativeCodingAgentForAgentName,
  nativeCodingAgentForAvailableAgent,
  nativeCodingAgentForHarness,
} from "@/lib/nativeCodingAgents";
import { authenticatedFetch } from "@/lib/identity";
import type { AvailableAgent, SessionInteractionMode } from "./availableAgentTypes";

const DISPLAY_NAMES: Record<string, string> = {
  nessie: "Nessie",
  polly: "Polly",
  debby: "Debby",
};

interface BuiltinAgentWire {
  id: string;
  worker_type_slug?: string;
  supported_modes?: SessionInteractionMode[];
  requires_model_resource?: boolean;
  name: string;
  description?: string | null;
  harness?: string | null;
  skills?: { name: string; description: string }[];
  builtin?: boolean;
  created_at?: number | null;
}

interface BuiltinAgentsListWire {
  data: BuiltinAgentWire[];
  has_more?: boolean;
  last_id?: string | null;
}

export function displayNameForAgent(name: string, harness?: string | null): string {
  return (
    nativeCodingAgentForHarness(harness)?.displayName ??
    nativeCodingAgentForAgentName(name)?.displayName ??
    DISPLAY_NAMES[name] ??
    capitalizeAgentName(name)
  );
}

export function dedupeNativeAgents(agents: AvailableAgent[]): AvailableAgent[] {
  const result: AvailableAgent[] = [];
  const nativeIndex = new Map<string, number>();
  for (const agent of agents) {
    const nativeAgent = nativeCodingAgentForAvailableAgent(agent);
    if (nativeAgent === undefined) {
      result.push(agent);
      continue;
    }
    const existingIndex = nativeIndex.get(nativeAgent.key);
    if (existingIndex === undefined) {
      nativeIndex.set(nativeAgent.key, result.length);
      result.push(agent);
      continue;
    }
    const existing = result[existingIndex];
    if (agent.name === nativeAgent.agentName && existing.name !== nativeAgent.agentName) {
      result[existingIndex] = agent;
    }
  }
  return result;
}

export async function fetchCatalogAgents(): Promise<AvailableAgent[]> {
  const rows: BuiltinAgentWire[] = [];
  let after: string | null = null;
  do {
    const url = after === null ? "/v1/agents" : `/v1/agents?after=${encodeURIComponent(after)}`;
    const response = await authenticatedFetch(url);
    if (!response.ok) throw new Error(`${response.status} ${response.statusText}`);
    const body = (await response.json()) as BuiltinAgentsListWire;
    rows.push(...body.data);
    after = body.has_more === true && body.last_id ? body.last_id : null;
  } while (after !== null);
  return rows.map((agent) => ({
    id: agent.id,
    name: agent.name,
    display_name: displayNameForAgent(agent.name, agent.harness),
    description: agent.description ?? null,
    harness: agent.harness ?? null,
    skills: agent.skills ?? [],
    builtin: agent.builtin,
    created_at: agent.created_at,
    ...workerCreationMetadata(agent),
  }));
}

function workerCreationMetadata(agent: BuiltinAgentWire): Partial<AvailableAgent> {
  const workerTypeSlug = agent.worker_type_slug ?? (agent.builtin === true ? agent.id : undefined);
  if (
    workerTypeSlug === undefined ||
    agent.supported_modes === undefined ||
    agent.requires_model_resource === undefined
  ) {
    return {};
  }
  return {
    workerTypeSlug,
    supportedModes: agent.supported_modes,
    requiresModelResource: agent.requires_model_resource,
  };
}
