import type { EffectiveResource, ProviderDefinition } from "@/lib/api/facade/aiResource";

const AGENT_PROTOCOLS: Record<string, string[]> = {
  "do-agent": ["openai-compatible", "anthropic", "minimax"],
  "codex-cli": ["openai-compatible"],
  "video-studio": ["openai-compatible"],
  "claude-code": ["anthropic"],
  "gemini-cli": ["gemini"],
  openclaw: ["openai-compatible", "anthropic", "gemini"],
  hermes: ["openai-compatible", "anthropic", "gemini"],
};

const MODEL_RESOURCE_AGENTS = new Set(Object.keys(AGENT_PROTOCOLS));

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
): EffectiveResource[] {
  const allowed = agentSlug ? AGENT_PROTOCOLS[agentSlug] : undefined;
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
        agentSupportsProtocol(agentSlug ?? "", protocol),
    );
  });
}

export function modelResourceLabel(resource: EffectiveResource): string {
  const model = resource.resource;
  const connection = resource.connection;
  const name = model?.displayName || model?.modelId || model?.identifier || "";
  if (!connection?.name) return name;
  return `${connection.name} · ${name}`;
}
