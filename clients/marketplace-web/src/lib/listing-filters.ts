import type { ListingQuery, ListingSort, ResourceType } from "./marketplace-types";

export type CatalogFilters = ListingQuery;

type SearchParams = Record<string, string | string[] | undefined>;

const resourceTypes = new Set<ResourceType>([
  "application",
  "skill",
  "mcp_connector",
  "resource",
]);
const listingSorts = new Set<ListingSort>(["featured", "latest", "relevance"]);

function first(value: string | string[] | undefined): string {
  return Array.isArray(value) ? value[0] || "" : value || "";
}

export function parseCatalogFilters(params: SearchParams): CatalogFilters {
  const type = first(params.type);
  const sort = first(params.sort);
  return {
    q: first(params.q).trim(),
    scene: first(params.scene).trim(),
    industry: first(params.industry).trim(),
    audience: first(params.audience).trim(),
    capability: first(params.capability).trim(),
    type: resourceTypes.has(type as ResourceType) ? (type as ResourceType) : "",
    integration: first(params.integration).trim(),
    readiness: first(params.readiness).trim(),
    space: first(params.space).trim(),
    sort: listingSorts.has(sort as ListingSort)
      ? (sort as ListingSort)
      : "featured",
  };
}
