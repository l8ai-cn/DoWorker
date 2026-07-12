import { ArrowRight, FolderKanban } from "lucide-react";
import Link from "next/link";

import type { ListingSummary, Space } from "@/lib/marketplace-types";

export function SpaceStrip({
  spaces,
  listings,
}: {
  spaces: Space[];
  listings: ListingSummary[];
}) {
  return (
    <section className="section-block" id="spaces">
      <div className="section-heading">
        <div>
          <span className="eyebrow">按工作场景发现</span>
          <h2>专区</h2>
        </div>
      </div>
      <div className="space-grid">
        {spaces.map((space) => {
          const count = listings.filter((listing) =>
            listing.spaces.some((item) => item.slug === space.slug),
          ).length;
          return (
            <Link
              className="space-card"
              href={`/catalog?space=${space.slug}`}
              key={space.slug}
            >
              <FolderKanban aria-hidden="true" size={22} />
              <span>
                <strong>{space.name}</strong>
                <small>{count} 个已上架内容</small>
              </span>
              <ArrowRight aria-hidden="true" size={18} />
            </Link>
          );
        })}
      </div>
    </section>
  );
}
