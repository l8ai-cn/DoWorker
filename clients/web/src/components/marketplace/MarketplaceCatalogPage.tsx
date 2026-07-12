"use client";

import { useEffect, useState } from "react";
import { Search, SlidersHorizontal } from "lucide-react";

import {
  fetchMarketplaceListings,
  fetchMarketplaceSummary,
  type MarketplaceListingSummary,
  type MarketplaceResourceType,
} from "@/lib/marketplace/catalog-api";
import { Input } from "@/components/ui/input";
import { PillTabs } from "@/components/ui/pill-tabs";
import { MarketplaceListingCard } from "./MarketplaceListingCard";

const typeTabs = [
  { id: "", label: "全部" },
  { id: "application", label: "应用" },
  { id: "skill", label: "Skill" },
  { id: "mcp_connector", label: "系统连接" },
  { id: "resource", label: "资源" },
];

export function MarketplaceCatalogPage({ orgSlug }: { orgSlug: string }) {
  const [listings, setListings] = useState<MarketplaceListingSummary[]>([]);
  const [name, setName] = useState("应用市场");
  const [summary, setSummary] = useState("为当前组织选择可以直接开始工作的 AI 能力。");
  const [query, setQuery] = useState("");
  const [type, setType] = useState<MarketplaceResourceType | "">("");
  const [space, setSpace] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    Promise.all([fetchMarketplaceSummary(), fetchMarketplaceListings()])
      .then(([market, items]) => {
        setName(market.name);
        setSummary(market.summary);
        setListings(items);
      })
      .catch((cause: unknown) => {
        setError(cause instanceof Error ? cause.message : "市场内容加载失败，请稍后重试。");
      })
      .finally(() => setLoading(false));
  }, []);

  const spaces = [...new Map(
    listings.flatMap((item) => item.spaces).map((item) => [item.slug, item]),
  ).values()];
  const normalizedQuery = query.trim().toLocaleLowerCase("zh-CN");
  const visible = listings.filter((item) => {
    const text = [item.display_name, item.tagline, item.publisher.display_name, ...item.spaces.map((s) => s.name)]
      .join(" ")
      .toLocaleLowerCase("zh-CN");
    return (!normalizedQuery || text.includes(normalizedQuery))
      && (!type || item.resource_type === type)
      && (!space || item.spaces.some((itemSpace) => itemSpace.slug === space));
  });

  return (
    <div className="mx-auto w-full max-w-7xl space-y-6 p-5 lg:p-8">
      <section className="overflow-hidden rounded-2xl border border-border bg-surface-raised shadow-[var(--shadow-soft)]">
        <div className="grid gap-6 bg-[linear-gradient(118deg,color-mix(in_srgb,var(--primary)_12%,transparent),transparent_55%)] p-6 sm:p-8 lg:grid-cols-[1fr_270px]">
          <div>
            <p className="text-sm font-medium text-primary">当前组织的能力目录</p>
            <h1 className="mt-2 text-3xl font-semibold tracking-tight text-foreground">{name}</h1>
            <p className="mt-3 max-w-2xl text-sm leading-7 text-muted-foreground">{summary}</p>
          </div>
          <div className="rounded-xl border border-primary/20 bg-primary/5 p-5">
            <p className="text-sm font-medium text-foreground">启用前可见</p>
            <p className="mt-2 text-sm leading-6 text-muted-foreground">
              当前组织的权限、运行要求与市场额度，确认后才会创建实例。
            </p>
          </div>
        </div>
      </section>

      <section className="rounded-xl border border-border bg-surface-raised p-4 shadow-[var(--shadow-soft)]">
        <div className="flex items-center gap-2">
          <Search className="h-4 w-4 text-muted-foreground" />
          <Input
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder="搜索应用、Skill、系统连接或资源"
            aria-label="搜索市场内容"
            className="border-0 bg-transparent shadow-none ring-0"
          />
        </div>
        <div className="mt-4 flex flex-col gap-3 border-t border-border pt-4">
          <PillTabs active={type} onChange={(value) => setType(value as MarketplaceResourceType | "")} tabs={typeTabs} />
          {spaces.length ? (
            <div className="flex items-center gap-2 overflow-x-auto pb-1">
              <SlidersHorizontal className="h-4 w-4 shrink-0 text-muted-foreground" />
              <button type="button" onClick={() => setSpace("")} className={space ? filterClass : selectedFilterClass}>
                全部专区
              </button>
              {spaces.map((item) => (
                <button key={item.slug} type="button" onClick={() => setSpace(item.slug)} className={space === item.slug ? selectedFilterClass : filterClass}>
                  {item.name}
                </button>
              ))}
            </div>
          ) : null}
        </div>
      </section>

      {loading ? <CatalogLoading /> : null}
      {error ? <p role="alert" className="rounded-lg bg-danger-bg p-4 text-sm text-danger">{error}</p> : null}
      {!loading && !error && listings.length === 0 ? <EmptyMarket /> : null}
      {!loading && !error && listings.length > 0 && visible.length === 0 ? <EmptyFilter /> : null}
      {!loading && visible.length ? (
        <section>
          <div className="mb-4 flex items-center justify-between">
            <h2 className="text-base font-semibold text-foreground">可用内容</h2>
            <span className="text-sm text-muted-foreground">{visible.length} 个结果</span>
          </div>
          <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
            {visible.map((listing) => <MarketplaceListingCard key={listing.listing_id} listing={listing} orgSlug={orgSlug} />)}
          </div>
        </section>
      ) : null}
    </div>
  );
}

const filterClass = "shrink-0 rounded-full px-3 py-1.5 text-sm text-muted-foreground hover:bg-muted hover:text-foreground";
const selectedFilterClass = "shrink-0 rounded-full bg-primary/10 px-3 py-1.5 text-sm font-medium text-primary";

function EmptyMarket() {
  return <p className="rounded-xl border border-dashed border-border p-10 text-center text-sm text-muted-foreground">这个市场还没有可用内容。市场管理员完成上架后，内容会显示在这里。</p>;
}

function EmptyFilter() {
  return <p className="rounded-xl border border-dashed border-border p-10 text-center text-sm text-muted-foreground">没有符合当前筛选条件的内容，请调整搜索词或专区。</p>;
}

function CatalogLoading() {
  return (
    <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3" aria-label="正在加载市场内容">
      {Array.from({ length: 6 }, (_, index) => (
        <div key={index} className="h-64 animate-pulse rounded-xl border border-border bg-surface-muted" />
      ))}
    </div>
  );
}
