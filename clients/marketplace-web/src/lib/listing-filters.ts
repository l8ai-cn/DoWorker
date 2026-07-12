import type { ListingSummary, ResourceType } from "./marketplace-types";

export interface CatalogFilters {
  q: string;
  type: ResourceType | "";
  space: string;
}

type SearchParams = Record<string, string | string[] | undefined>;

const resourceTypes = new Set<ResourceType>([
  "application",
  "skill",
  "mcp_connector",
  "resource",
]);

function first(value: string | string[] | undefined): string {
  return Array.isArray(value) ? value[0] || "" : value || "";
}

export function parseCatalogFilters(params: SearchParams): CatalogFilters {
  const type = first(params.type);
  return {
    q: first(params.q).trim(),
    type: resourceTypes.has(type as ResourceType) ? (type as ResourceType) : "",
    space: first(params.space).trim(),
  };
}

export function filterListings(
  listings: ListingSummary[],
  filters: CatalogFilters,
): ListingSummary[] {
  const query = filters.q.toLocaleLowerCase("zh-CN");
  return listings.filter((listing) => {
    const searchable = [
      listing.display_name,
      listing.tagline,
      listing.publisher.display_name,
      ...listing.spaces.map((space) => space.name),
    ]
      .join(" ")
      .toLocaleLowerCase("zh-CN");
    return (
      (!query || searchable.includes(query)) &&
      (!filters.type || listing.resource_type === filters.type) &&
      (!filters.space ||
        listing.spaces.some((space) => space.slug === filters.space))
    );
  });
}
