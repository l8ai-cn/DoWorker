import { MarketplaceCatalogPage } from "@/components/marketplace/MarketplaceCatalogPage";

export default async function MarketplacePage({
  params,
}: {
  params: Promise<{ org: string }>;
}) {
  const { org } = await params;
  return <MarketplaceCatalogPage orgSlug={org} />;
}
