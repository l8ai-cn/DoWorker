import type {
  EffectiveResource,
  ProviderDefinition,
} from "@/lib/api/facade/aiResource";
import type {
  WorkerToolModelRequirement,
  WorkerTypeOption,
} from "@/lib/api/facade/podConnect";

export interface WorkerModelRequirement {
  provider_keys?: string[];
  protocol_adapters: string[];
  modality: string;
  capability: string;
}

export const generationModelRequirement: WorkerModelRequirement = {
  protocol_adapters: ["openai-compatible", "anthropic", "gemini"],
  modality: "chat",
  capability: "text-generation",
};

export function primaryModelRequirement(
  workerType: WorkerTypeOption,
): WorkerModelRequirement | null {
  if (!workerType.requires_model_resource) return null;
  return {
    protocol_adapters: workerType.model_protocol_adapters,
    modality: "chat",
    capability: "text-generation",
  };
}

export function toolModelRequirement(
  requirement: WorkerToolModelRequirement,
): WorkerModelRequirement {
  return requirement;
}

export function compatibleWorkerModelResources(
  resources: EffectiveResource[],
  providers: ProviderDefinition[],
  requirement: WorkerModelRequirement,
): EffectiveResource[] {
  const adapterByProvider = new Map(
    providers.map((provider) => [provider.key, provider.protocolAdapter]),
  );
  return resources.filter((item) => {
    const connection = item.connection;
    const resource = item.resource;
    if (!item.selectable || !connection?.isEnabled || !resource?.isEnabled) {
      return false;
    }
    if (
      requirement.provider_keys?.length &&
      !requirement.provider_keys.includes(connection.providerKey)
    ) {
      return false;
    }
    const adapter = adapterByProvider.get(connection.providerKey);
    return Boolean(
      adapter &&
        requirement.protocol_adapters.includes(adapter) &&
        resource.modalities.includes(requirement.modality) &&
        resource.capabilities.includes(requirement.capability),
    );
  });
}
