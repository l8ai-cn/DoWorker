import { Search } from "lucide-react";

import type { CatalogFilters } from "@/lib/listing-filters";

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
        placeholder="搜索应用、Skill、系统连接或资源"
      />
      {filters.type && <input type="hidden" name="type" value={filters.type} />}
      {filters.space && <input type="hidden" name="space" value={filters.space} />}
      <button className="button button-primary" type="submit">
        搜索
      </button>
    </form>
  );
}
