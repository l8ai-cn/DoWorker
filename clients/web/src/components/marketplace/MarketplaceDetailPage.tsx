"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { ArrowLeft, BadgeCheck, Check, ExternalLink, ShieldCheck } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  fetchMarketplaceListingDetail,
  type MarketplaceListingDetail,
} from "@/lib/marketplace/catalog-api";
import {
  formatMarketplaceCredits,
  marketplaceTypeLabels,
} from "@/lib/marketplace/presentation";

export function MarketplaceDetailPage({
  orgSlug,
  listingSlug,
}: {
  orgSlug: string;
  listingSlug: string;
}) {
  const [listing, setListing] = useState<MarketplaceListingDetail | null>(null);
  const [error, setError] = useState("");

  useEffect(() => {
    fetchMarketplaceListingDetail(listingSlug)
      .then(setListing)
      .catch((cause: unknown) => setError(cause instanceof Error ? cause.message : "市场内容加载失败。"));
  }, [listingSlug]);

  if (error) return <State message={error} />;
  if (!listing) return <State message="正在加载应用详情" />;

  const credits = formatMarketplaceCredits(listing.quota);
  const canEnable = listing.resource_type === "application";
  return (
    <div className="mx-auto w-full max-w-6xl space-y-6 p-5 lg:p-8">
      <Link href={`/${orgSlug}/marketplace`} className="inline-flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground">
        <ArrowLeft className="h-4 w-4" />
        返回应用市场
      </Link>
      <section className="rounded-2xl border border-border bg-surface-raised p-6 shadow-[var(--shadow-soft)] sm:p-8">
        <div className="grid gap-8 lg:grid-cols-[1fr_300px]">
          <div>
            <div className="flex flex-wrap gap-2">
              <Badge variant="info">{marketplaceTypeLabels[listing.resource_type]}</Badge>
              {listing.spaces.map((space) => <Badge key={space.slug} variant="secondary">{space.name}</Badge>)}
            </div>
            <h1 className="mt-5 text-3xl font-semibold tracking-tight text-foreground">{listing.display_name}</h1>
            <p className="mt-3 text-base leading-7 text-muted-foreground">{listing.tagline}</p>
            <p className="mt-4 flex items-center gap-1.5 text-sm text-muted-foreground">
              {listing.publisher.display_name}
              {listing.publisher.verified ? <span className="inline-flex items-center gap-1 text-success"><BadgeCheck className="h-4 w-4" />已认证发布方</span> : null}
            </p>
          </div>
          <aside className="rounded-xl border border-border bg-surface-muted/60 p-5">
            <p className="text-xs font-medium text-muted-foreground">当前版本</p>
            <p className="mt-1 text-xl font-semibold text-foreground">v{listing.version}</p>
            <p className="mt-5 text-xs font-medium text-muted-foreground">预计额度</p>
            <p className="mt-1 text-sm text-foreground">{credits ?? "将在启用确认时核对"}</p>
            {canEnable ? (
              <Button asChild className="mt-6 w-full">
                <Link href={`/${orgSlug}/marketplace/acquire?market=do-worker-market&listing=${listing.slug}&version=${listing.listing_version_id}`}>
                  检查启用条件
                </Link>
              </Button>
            ) : (
              <div className="mt-6">
                <Button className="w-full" disabled>暂不支持启用</Button>
                <p className="mt-2 text-xs leading-5 text-muted-foreground">对应运行时接入完成后开放此资源类型。</p>
              </div>
            )}
          </aside>
        </div>
      </section>
      <section className="grid gap-6 lg:grid-cols-[1fr_320px]">
        <div className="space-y-6">
          <DetailSection title="能力说明"><p>{listing.description}</p></DetailSection>
          <DetailList title="可以完成什么" items={listing.outcomes} />
          <DetailList title="适用场景" items={listing.use_cases} />
          <DetailList title="适用对象" items={listing.target_audience} />
        </div>
        <div className="space-y-6">
          <DetailList title="启用要求" items={listing.requirements} />
          <DetailList title="所需权限" items={listing.permissions} icon={ShieldCheck} />
          <DetailSection title="版本说明"><p>{listing.release_notes || "暂无版本说明。"}</p></DetailSection>
          {(listing.documentation_url || listing.support_url) ? <div className="space-y-2 text-sm">
            {listing.documentation_url ? <a className="flex items-center gap-1 text-primary hover:text-primary/80" href={listing.documentation_url}>查看文档 <ExternalLink className="h-3.5 w-3.5" /></a> : null}
            {listing.support_url ? <a className="flex items-center gap-1 text-primary hover:text-primary/80" href={listing.support_url}>获取支持 <ExternalLink className="h-3.5 w-3.5" /></a> : null}
          </div> : null}
        </div>
      </section>
    </div>
  );
}

function State({ message }: { message: string }) {
  return <div className="mx-auto max-w-6xl p-8 text-sm text-muted-foreground">{message}</div>;
}

function DetailSection({ title, children }: { title: string; children: React.ReactNode }) {
  return <section className="rounded-xl border border-border bg-surface-raised p-5 shadow-[var(--shadow-soft)]"><h2 className="text-base font-semibold text-foreground">{title}</h2><div className="mt-3 text-sm leading-7 text-muted-foreground">{children}</div></section>;
}

function DetailList({ title, items, icon: Icon = Check }: { title: string; items: string[]; icon?: typeof Check }) {
  if (!items.length) return null;
  return <DetailSection title={title}><ul className="space-y-2">{items.map((item) => <li className="flex gap-2" key={item}><Icon className="mt-1 h-4 w-4 shrink-0 text-primary" />{item}</li>)}</ul></DetailSection>;
}
