import { BadgeCheck, ChevronLeft } from "lucide-react";
import Link from "next/link";

import { MARKET_SLUG } from "@/lib/marketplace-api";
import {
  formatQuotaSummary,
  resourceTypeLabels,
} from "@/lib/listing-presentation";
import type { ListingDetail } from "@/lib/marketplace-types";

import { AcquireButton } from "./acquire-button";

export function DetailHero({ listing }: { listing: ListingDetail }) {
  const quota = formatQuotaSummary(listing.quota);

  return (
    <section className="detail-hero">
      <Link className="back-link" href="/catalog">
        <ChevronLeft aria-hidden="true" size={17} />
        返回全部内容
      </Link>
      <div className="detail-hero-grid">
        <div>
          <div className="detail-badges">
            <span>{resourceTypeLabels[listing.resource_type]}</span>
            {listing.spaces.map((space) => (
              <span key={space.slug}>{space.name}</span>
            ))}
          </div>
          <h1>{listing.display_name}</h1>
          <p className="detail-tagline">{listing.tagline}</p>
          <div className="publisher-line">
            <span>{listing.publisher.display_name}</span>
            {listing.publisher.verified && (
              <span className="verified">
                <BadgeCheck aria-hidden="true" size={15} />
                已认证发布方
              </span>
            )}
          </div>
        </div>
        <aside className="acquire-panel">
          <span className="eyebrow">当前版本</span>
          <strong>{listing.version}</strong>
          <dl>
            <div>
              <dt>额度</dt>
              <dd>{quota || "请在启用确认页核对"}</dd>
            </div>
          </dl>
          <AcquireButton
            coreWebUrl={process.env.CORE_WEB_URL}
            resourceType={listing.resource_type}
            target={{
              market: MARKET_SLUG,
              listing: listing.slug,
              version: listing.listing_version_id,
            }}
          />
        </aside>
      </div>
    </section>
  );
}
