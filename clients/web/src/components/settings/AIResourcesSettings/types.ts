import type {
  EffectiveResource,
  ModelResource,
  ProviderConnection,
  ProviderDefinition,
} from "@/lib/api";

export type AIResourceScope = "personal" | "organization";

export interface AIResourcesData {
  catalog: ProviderDefinition[];
  connections: ProviderConnection[];
  effectiveResources: EffectiveResource[];
}

export interface AIResourceDeletionTarget {
  kind: "connection" | "resource";
  id: number;
  name: string;
}

export type { ModelResource, ProviderConnection, ProviderDefinition };
