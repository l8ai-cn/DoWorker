import type { EffectiveResource, ProviderDefinition } from "@/lib/api/facade/aiResource";
import type { WorkerToolModelRequirement } from "@/lib/api/facade/podConnect";

const AGENT_PROTOCOLS: Record<string, string[]> = {
  "do-agent": ["openai-compatible", "anthropic", "minimax"],
  "seedance-expert": ["openai-compatible", "anthropic"],
  "codex-cli": ["openai-compatible"],
  "video-studio": ["openai-compatible"],
  "claude-code": ["anthropic"],
  "gemini-cli": ["gemini"],
  "minimax-cli": ["minimax"],
  openclaw: ["openai-compatible"],
  hermes: ["openai-compatible"],
};

const MODEL_RESOURCE_AGENTS = new Set(Object.keys(AGENT_PROTOCOLS));

export interface WorkerModelResourceRequirement {
  required: boolean;
  protocolAdapters: string[];
}

export function agentRequiresModelResource(agentSlug: string | null): boolean {
  return Boolean(agentSlug && MODEL_RESOURCE_AGENTS.has(agentSlug));
}

export function agentSupportsProtocol(agentSlug: string, protocol: string): boolean {
  return AGENT_PROTOCOLS[agentSlug]?.includes(protocol) ?? false;
}

export function compatibleModelResources(
  agentSlug: string | null,
  resources: EffectiveResource[],
  providers: ProviderDefinition[],
  requirement?: WorkerModelResourceRequirement,
): EffectiveResource[] {
  const allowed = requirement
    ? (requirement.required ? requirement.protocolAdapters : [])
    : (agentSlug ? AGENT_PROTOCOLS[agentSlug] : undefined);
  if (!allowed?.length) return [];
  const protocolByProvider = new Map(providers.map((p) => [p.key, p.protocolAdapter]));
  return resources.filter((item) => {
    const providerKey = item.connection?.providerKey;
    const protocol = providerKey ? protocolByProvider.get(providerKey) : undefined;
    return Boolean(
      item.selectable &&
        item.connection?.isEnabled &&
        item.resource?.isEnabled &&
        item.resource.modalities.includes("chat") &&
        item.resource.capabilities.includes("text-generation") &&
        protocol &&
        allowed.includes(protocol),
    );
  });
}

export function compatibleToolModelResources(
  requirement: WorkerToolModelRequirement,
  resources: EffectiveResource[],
): EffectiveResource[] {
  return resources.filter((item) => {
    const providerKey = item.connection?.providerKey;
    return Boolean(
      item.selectable &&
        item.connection?.isEnabled &&
        item.resource?.isEnabled &&
        providerKey &&
        requirement.provider_keys.includes(providerKey) &&
        item.resource.modalities.includes(requirement.modality) &&
        item.resource.capabilities.includes(requirement.capability) &&
        matchesToolModelFamily(requirement, providerKey, item.resource.modelId),
    );
  });
}

function matchesToolModelFamily(
  requirement: WorkerToolModelRequirement,
  providerKey: string,
  modelId: string,
): boolean {
  if (requirement.capability !== "video-generation") return true;
  if (providerKey === "doubao") {
    return modelId.trim().startsWith("doubao-seedance-");
  }
  if (providerKey === "sub2api-seedance") {
    return modelId.trim() === "doubao-seedance-2-0-260128";
  }
  return true;
}

export function modelResourceLabel(resource: EffectiveResource): string {
  const model = resource.resource;
  const connection = resource.connection;
  const name = model?.displayName || model?.modelId || model?.identifier || "";
  if (!connection?.name) return name;
  return `${connection.name} · ${name}`;
}

export function toolModelRoleLabel(role: string): string {
  const words = role.split("-").filter(Boolean);
  if (words.length === 0) return role;
  return [capitalize(words[0]), ...words.slice(1)].join(" ");
}

function capitalize(value: string): string {
  return value.charAt(0).toUpperCase() + value.slice(1);
}
