import type { Metadata } from "next";
import { notFound } from "next/navigation";

import { DetailContent } from "@/components/detail-content";
import { DetailHero } from "@/components/detail-hero";
import { getListing, MarketplaceApiError } from "@/lib/marketplace-api";

type ListingParams = Promise<{ listingSlug: string }>;

export const dynamic = "force-dynamic";

async function loadListing(params: ListingParams) {
  try {
    return await getListing((await params).listingSlug);
  } catch (error) {
    if (
      error instanceof MarketplaceApiError &&
      error.code === "LISTING_NOT_AVAILABLE"
    ) {
      notFound();
    }
    throw error;
  }
}

export async function generateMetadata({
  params,
}: {
  params: ListingParams;
}): Promise<Metadata> {
  try {
    const listing = await getListing((await params).listingSlug);
    return {
      title: listing.display_name,
      description: listing.tagline,
    };
  } catch {
    return { title: "应用详情" };
  }
}

export default async function ApplicationPage({
  params,
}: {
  params: ListingParams;
}) {
  const listing = await loadListing(params);
  return (
    <main className="shell page-main">
      <DetailHero listing={listing} />
      <DetailContent listing={listing} />
    </main>
  );
}
