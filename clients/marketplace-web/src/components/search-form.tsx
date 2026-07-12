import { Search } from "lucide-react";

import type { CatalogFilters } from "@/lib/listing-filters";

const hiddenFilterKeys: Array<keyof CatalogFilters> = [
  "scene",
  "industry",
  "audience",
  "capability",
  "type",
  "integration",
  "readiness",
  "space",
  "sort",
];

export function SearchForm({ filters }: { filters: CatalogFilters }) {
  return (
    <form className="search-form" action="/catalog">
      <Search aria-hidden="true" size={19} />
      <label className="sr-only" htmlFor="marketplace-search">
        搜索市场内容
      </label>
      <input
        id="marketplace-search"
        name="q"
        defaultValue={filters.q}
        placeholder="搜索应用、交付结果或系统连接"
      />
      {hiddenFilterKeys.map(
        (key) =>
          filters[key] && (
            <input key={key} type="hidden" name={key} value={filters[key]} />
          ),
      )}
      <button className="button button-primary" type="submit">
        搜索
      </button>
    </form>
  );
}
