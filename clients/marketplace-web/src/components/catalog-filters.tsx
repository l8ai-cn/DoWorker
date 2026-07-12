import Link from "next/link";

import {
  collectTaxonomyTags,
  taxonomyFilterGroups,
} from "@/lib/catalog-filter-options";
import type { CatalogFilters } from "@/lib/listing-filters";
import { resourceTypeLabels } from "@/lib/listing-presentation";
import type {
  ListingSort,
  ListingSummary,
  ResourceType,
  Space,
} from "@/lib/marketplace-types";

const types = Object.entries(resourceTypeLabels) as Array<[ResourceType, string]>;
const sorts: Array<[ListingSort, string]> = [
  ["featured", "优先推荐"],
  ["latest", "最新发布"],
  ["relevance", "与搜索最相关"],
];

function selectedTag(
  kind: Parameters<typeof collectTaxonomyTags>[1],
  value: string,
) {
  return value ? { slug: value, display_name: value, kind } : undefined;
}

function filterHref(
  filters: CatalogFilters,
  change: Partial<CatalogFilters>,
): string {
  const next = { ...filters, ...change };
  const params = new URLSearchParams();
  (Object.entries(next) as Array<[keyof CatalogFilters, string]>).forEach(
    ([key, value]) => {
      if (value && !(key === "sort" && value === "featured")) {
        params.set(key, value);
      }
    },
  );
  const query = params.toString();
  return query ? `/catalog?${query}` : "/catalog";
}

interface CatalogFiltersProps {
  filters: CatalogFilters;
  listings: ListingSummary[];
  spaces: Space[];
}

function FilterLinks({
  label,
  options,
  value,
  filters,
  filterKey,
}: {
  label: string;
  options: Array<[string, string]>;
  value: string;
  filters: CatalogFilters;
  filterKey: keyof CatalogFilters;
}) {
  return (
    <div className="filter-group">
      <span>{label}</span>
      <div className="filter-options">
        <Link className={!value ? "selected" : ""} href={filterHref(filters, { [filterKey]: "" })}>
          全部
        </Link>
        {options.map(([optionValue, optionLabel]) => (
          <Link
            className={value === optionValue ? "selected" : ""}
            href={filterHref(filters, { [filterKey]: optionValue })}
            key={optionValue}
          >
            {optionLabel}
          </Link>
        ))}
      </div>
    </div>
  );
}

export function CatalogFilters({ filters, listings, spaces }: CatalogFiltersProps) {
  const controls = (
    <div className="filter-groups">
      <FilterLinks label="资源类型" options={types} value={filters.type} filters={filters} filterKey="type" />
      <FilterLinks
        label="工作专区"
        options={spaces.map((space) => [space.slug, space.name])}
        value={filters.space}
        filters={filters}
        filterKey="space"
      />
      {taxonomyFilterGroups.map((group) => (
        <FilterLinks
          filters={filters}
          filterKey={group.key}
          key={group.key}
          label={group.label}
          options={collectTaxonomyTags(
            listings,
            group.key,
            selectedTag(group.key, filters[group.key]),
          ).map((tag) => [tag.slug, tag.display_name])}
          value={filters[group.key]}
        />
      ))}
      <FilterLinks label="排序方式" options={sorts} value={filters.sort} filters={filters} filterKey="sort" />
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
