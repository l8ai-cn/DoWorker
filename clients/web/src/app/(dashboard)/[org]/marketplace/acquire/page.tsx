import { Suspense } from "react";

import { MarketplaceAcquireFlow } from "@/components/marketplace/acquire/MarketplaceAcquireFlow";

export default async function MarketplaceAcquireRoute({
  params,
}: {
  params: Promise<{ org: string }>;
}) {
  const { org } = await params;
  return <Suspense fallback={null}><MarketplaceAcquireFlow organizationSlug={org} /></Suspense>;
}
