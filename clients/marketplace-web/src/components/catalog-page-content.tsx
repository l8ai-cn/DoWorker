import { CatalogFilters } from "@/components/catalog-filters";
import { ListingGrid } from "@/components/listing-grid";
import { MarketIntro } from "@/components/market-intro";
import { SearchForm } from "@/components/search-form";
import { SpaceStrip } from "@/components/space-strip";
import { StatePanel } from "@/components/state-panel";
import { parseCatalogFilters } from "@/lib/listing-filters";
import {
  getMarket,
  listListings,
  MarketplaceApiError,
} from "@/lib/marketplace-api";
import type { ListingSummary, Space } from "@/lib/marketplace-types";

type SearchParams = Promise<Record<string, string | string[] | undefined>>;

function uniqueSpaces(listings: ListingSummary[]): Space[] {
  const spaces = new Map<string, Space>();
  listings.flatMap((listing) => listing.spaces).forEach((space) => {
    spaces.set(space.slug, space);
  });
  return [...spaces.values()];
}

async function loadCatalog(searchParams: SearchParams) {
  try {
    const rawParams = await searchParams;
    const filters = parseCatalogFilters(rawParams);
    const [market, collection] = await Promise.all([
      getMarket(),
      listListings(filters),
    ]);
    return { kind: "ready" as const, market, collection, filters };
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

  const listings = data.collection.items;
  const spaces = uniqueSpaces(listings);
  return (
    <main className="shell page-main">
      {!catalogOnly && <MarketIntro market={data.market} />}
      <SearchForm filters={data.filters} />
      {!catalogOnly && spaces.length > 0 && (
        <SpaceStrip spaces={spaces} listings={listings} />
      )}
      <section className="section-block">
        <div className="section-heading">
          <div>
            <span className="eyebrow">
              {catalogOnly ? "应用目录" : "已发布应用"}
            </span>
            <h2>{catalogOnly ? "全部应用" : "从业务结果开始选择"}</h2>
          </div>
          <span className="result-count">{listings.length} 个结果</span>
        </div>
        <CatalogFilters filters={data.filters} listings={listings} spaces={spaces} />
        {listings.length ? (
          <ListingGrid listings={listings} />
        ) : (
          <StatePanel
            kind="empty"
            title="没有符合筛选条件的应用"
            description="调整筛选条件，或回到全部应用查看当前已发布的内容。"
            action={{ href: "/catalog", label: "清除筛选" }}
          />
        )}
      </section>
    </main>
  );
}
