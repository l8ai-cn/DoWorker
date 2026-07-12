import { MarketplaceDetailPage } from "@/components/marketplace/MarketplaceDetailPage";

export default async function MarketplaceDetailRoute({
  params,
}: {
  params: Promise<{ org: string; listingSlug: string }>;
}) {
  const { org, listingSlug } = await params;
  return <MarketplaceDetailPage orgSlug={org} listingSlug={listingSlug} />;
}
