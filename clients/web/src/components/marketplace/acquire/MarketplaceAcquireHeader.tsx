import Link from "next/link";
import { ArrowLeft } from "lucide-react";

import type { MarketplaceListingDetail } from "@/lib/marketplace/acquire-api";

export function MarketplaceAcquireHeader({
  listing,
  organizationSlug,
}: {
  listing: MarketplaceListingDetail;
  organizationSlug?: string;
}) {
  return (
    <header className="border-b border-border pb-6">
      <Link
        href={
          organizationSlug
            ? `/${organizationSlug}/marketplace/${listing.slug}`
            : "/marketplace"
        }
        className="inline-flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground"
      >
        <ArrowLeft className="h-4 w-4" />
        返回应用市场
      </Link>
      <p className="mt-6 text-sm font-medium text-primary">专家应用启用向导</p>
      <h1 className="mt-2 text-3xl font-semibold text-foreground">
        {listing.display_name}
      </h1>
      <p className="mt-3 text-base leading-7 text-muted-foreground">
        {listing.tagline}
      </p>
    </header>
  );
}
