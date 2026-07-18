import workerRuntimeCatalog from "@/generated/worker-runtime-catalog.json";
import { lightConnect } from "@/lib/light-auth/api-fetch";

const AI_RESOURCE_SERVICE = "proto.ai_resource.v1.AIResourceService";

interface ToolModelRequirement {
  id: string;
  provider_keys: string[];
  modality: string;
  capability: string;
}

interface EffectiveResourcesResponse {
  resources?: Array<{
    selectable?: boolean;
    connection?: {
      providerKey?: string;
      name?: string;
      isEnabled?: boolean;
    };
    resource?: {
      id?: string | number;
      displayName?: string;
      modelId?: string;
      isEnabled?: boolean;
      modalities?: string[];
      capabilities?: string[];
    };
  }>;
}

export interface MarketplaceToolModelResource {
  id: number;
  label: string;
}

export interface MarketplaceToolModelGroup {
  role: string;
  resources: MarketplaceToolModelResource[];
}

export async function listMarketplaceToolModelResources(
  orgSlug: string,
  agentSlug: string,
): Promise<MarketplaceToolModelGroup[]> {
  const requirements = toolModelRequirements(agentSlug);
  if (requirements.length === 0) return [];
  const effective = await lightConnect<
    { orgSlug: string; modalities: string[] },
    EffectiveResourcesResponse
  >(
    AI_RESOURCE_SERVICE,
    "ListOrganizationEffectiveResources",
    {
      orgSlug,
      modalities: [...new Set(requirements.map((item) => item.modality))],
    },
    { authenticated: true },
  );
  return requirements.map((requirement) => ({
    role: requirement.id,
    resources: (effective.resources ?? []).flatMap((item) =>
      matchesRequirement(item, requirement)
        ? [marketplaceToolResource(item)]
        : [],
    ),
  }));
}

function toolModelRequirements(agentSlug: string): ToolModelRequirement[] {
  const worker = workerRuntimeCatalog.workers.find(
    (item) => item.slug === agentSlug,
  );
  return "toolModelRequirements" in (worker ?? {})
    ? ((worker as { toolModelRequirements?: ToolModelRequirement[] })
        .toolModelRequirements ?? [])
    : [];
}

function matchesRequirement(
  item: NonNullable<EffectiveResourcesResponse["resources"]>[number],
  requirement: ToolModelRequirement,
): boolean {
  const connection = item.connection;
  const resource = item.resource;
  return Boolean(
    item.selectable &&
      connection?.isEnabled !== false &&
      resource?.isEnabled &&
      connection?.providerKey &&
      requirement.provider_keys.includes(connection.providerKey) &&
      resource.modalities?.includes(requirement.modality) &&
      resource.capabilities?.includes(requirement.capability) &&
      matchesModelFamily(connection.providerKey, requirement, resource.modelId),
  );
}

function matchesModelFamily(
  providerKey: string,
  requirement: ToolModelRequirement,
  modelID: string | undefined,
): boolean {
  if (requirement.capability !== "video-generation") return true;
  if (providerKey === "doubao") {
    return modelID?.trim().startsWith("doubao-seedance-") ?? false;
  }
  if (providerKey === "sub2api-seedance") {
    return modelID?.trim() === "doubao-seedance-2-0-260128";
  }
  return true;
}

function marketplaceToolResource(
  item: NonNullable<EffectiveResourcesResponse["resources"]>[number],
): MarketplaceToolModelResource {
  const id = Number(item.resource?.id);
  if (!Number.isSafeInteger(id) || id <= 0) {
    throw new Error("unsafe tool model resource id");
  }
  const name =
    item.resource?.displayName || item.resource?.modelId || String(id);
  return {
    id,
    label: item.connection?.name ? `${item.connection.name} · ${name}` : name,
  };
}
