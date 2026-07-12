import { CatalogFilters } from "@/components/catalog-filters";
import { ListingGrid } from "@/components/listing-grid";
import { MarketIntro } from "@/components/market-intro";
import { SearchForm } from "@/components/search-form";
import { SpaceStrip } from "@/components/space-strip";
import { StatePanel } from "@/components/state-panel";
import {
  filterListings,
  parseCatalogFilters,
} from "@/lib/listing-filters";
import {
  getMarket,
  listListings,
  MarketplaceApiError,
} from "@/lib/marketplace-api";
import type { Space } from "@/lib/marketplace-types";

type SearchParams = Promise<Record<string, string | string[] | undefined>>;

function uniqueSpaces(listings: Awaited<ReturnType<typeof listListings>>): Space[] {
  const spaces = new Map<string, Space>();
  listings.flatMap((listing) => listing.spaces).forEach((space) => {
    spaces.set(space.slug, space);
  });
  return [...spaces.values()];
}

async function loadCatalog(searchParams: SearchParams) {
  try {
    const [market, listings, rawParams] = await Promise.all([
      getMarket(),
      listListings(),
      searchParams,
    ]);
    return { kind: "ready" as const, market, listings, rawParams };
  } catch (error) {
    if (error instanceof MarketplaceApiError && error.code === "MARKET_SUSPENDED") {
      return { kind: "suspended" as const };
    }
    throw error;
  }
}

export async function CatalogPageContent({
  searchParams,
  catalogOnly = false,
}: {
  searchParams: SearchParams;
  catalogOnly?: boolean;
}) {
  const data = await loadCatalog(searchParams);
  if (data.kind === "suspended") {
    return (
      <main className="shell page-main">
        <StatePanel
          kind="suspended"
          title="市场暂时停止服务"
          description="你仍可查看已启用内容，但不能获取或安装新内容。"
        />
      </main>
    );
  }

  const filters = parseCatalogFilters(data.rawParams);
  const filtered = filterListings(data.listings, filters);
  const spaces = uniqueSpaces(data.listings);
  return (
    <main className="shell page-main">
      {!catalogOnly && <MarketIntro market={data.market} />}
      <SearchForm filters={filters} />
      {!catalogOnly && spaces.length > 0 && (
        <SpaceStrip spaces={spaces} listings={data.listings} />
      )}
      <section className="section-block">
        <div className="section-heading">
          <div>
            <span className="eyebrow">
              {catalogOnly ? "MARKET CATALOG" : "最新上架"}
            </span>
            <h2>{catalogOnly ? "全部市场内容" : "开始处理真实工作"}</h2>
          </div>
          <span className="result-count">{filtered.length} 个结果</span>
        </div>
        <CatalogFilters filters={filters} spaces={spaces} />
        {filtered.length ? (
          <ListingGrid listings={filtered} />
        ) : (
          <StatePanel
            kind="empty"
            title={data.listings.length ? "没有符合筛选条件的内容" : "这个市场还没有可用内容"}
            description={
              data.listings.length
                ? "尝试清除部分筛选条件或使用更短的搜索词。"
                : "市场管理员完成上架后，内容会显示在这里。"
            }
            action={data.listings.length ? { href: "/catalog", label: "清除筛选" } : undefined}
          />
        )}
      </section>
    </main>
  );
}
