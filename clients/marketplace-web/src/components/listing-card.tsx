import {
  AppWindow,
  BadgeCheck,
  Blocks,
  Database,
  PlugZap,
} from "lucide-react";
import Link from "next/link";

import {
  formatQuotaSummary,
  resourceTypeLabels,
} from "@/lib/listing-presentation";
import type { ListingSummary, ResourceType } from "@/lib/marketplace-types";

const icons = {
  application: AppWindow,
  skill: Blocks,
  mcp_connector: PlugZap,
  resource: Database,
} satisfies Record<ResourceType, typeof AppWindow>;

export function ListingCard({ listing }: { listing: ListingSummary }) {
  const Icon = icons[listing.resource_type];
  const quota = formatQuotaSummary(listing.quota);

  return (
    <article className="listing-card">
      <div className="listing-card-topline">
        <span className={`listing-icon type-${listing.resource_type}`}>
          <Icon aria-hidden="true" size={21} />
        </span>
        <span className="type-label">
          {resourceTypeLabels[listing.resource_type]}
        </span>
      </div>
      <div className="listing-card-body">
        <div>
          <h3>{listing.display_name}</h3>
          <p>{listing.tagline}</p>
        </div>
        <div className="publisher-line">
          <span>{listing.publisher.display_name}</span>
          {listing.publisher.verified && (
            <span className="verified">
              <BadgeCheck aria-hidden="true" size={15} />
              已认证
            </span>
          )}
        </div>
      </div>
      <div className="listing-card-meta">
        <span>{listing.spaces[0]?.name || "未分配专区"}</span>
        {quota && <span>{quota}</span>}
      </div>
      <Link
        className="card-link"
        href={`/listings/${listing.slug}`}
        aria-label={`查看${listing.display_name}`}
      >
        查看详情
        <span aria-hidden="true">→</span>
      </Link>
    </article>
  );
}
