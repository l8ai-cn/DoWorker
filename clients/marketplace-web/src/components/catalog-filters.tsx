import Link from "next/link";

import type { CatalogFilters } from "@/lib/listing-filters";
import { resourceTypeLabels } from "@/lib/listing-presentation";
import type { ResourceType, Space } from "@/lib/marketplace-types";

const types = Object.entries(resourceTypeLabels) as Array<[ResourceType, string]>;

function filterHref(
  filters: CatalogFilters,
  change: Partial<CatalogFilters>,
): string {
  const next = { ...filters, ...change };
  const params = new URLSearchParams();
  if (next.q) params.set("q", next.q);
  if (next.type) params.set("type", next.type);
  if (next.space) params.set("space", next.space);
  const query = params.toString();
  return query ? `/catalog?${query}` : "/catalog";
}

interface CatalogFiltersProps {
  filters: CatalogFilters;
  spaces: Space[];
}

export function CatalogFilters({ filters, spaces }: CatalogFiltersProps) {
  const controls = (
    <div className="filter-groups">
      <div className="filter-group">
        <span>内容类型</span>
        <div className="filter-options">
          <Link className={!filters.type ? "selected" : ""} href={filterHref(filters, { type: "" })}>
            全部
          </Link>
          {types.map(([value, label]) => (
            <Link
              className={filters.type === value ? "selected" : ""}
              href={filterHref(filters, { type: value })}
              key={value}
            >
              {label}
            </Link>
          ))}
        </div>
      </div>
      <div className="filter-group">
        <span>所属专区</span>
        <div className="filter-options">
          <Link className={!filters.space ? "selected" : ""} href={filterHref(filters, { space: "" })}>
            全部
          </Link>
          {spaces.map((space) => (
            <Link
              className={filters.space === space.slug ? "selected" : ""}
              href={filterHref(filters, { space: space.slug })}
              key={space.slug}
            >
              {space.name}
            </Link>
          ))}
        </div>
      </div>
    </div>
  );

  return (
    <>
      <div className="desktop-filters">{controls}</div>
      <details className="mobile-filters">
        <summary>筛选内容</summary>
        {controls}
      </details>
    </>
  );
}
