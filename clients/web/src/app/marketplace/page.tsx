import { Suspense } from "react";

import { MarketplaceEntryRedirect } from "@/components/marketplace/MarketplaceEntryRedirect";

export default function MarketplacePage() {
  return <Suspense fallback={null}><MarketplaceEntryRedirect /></Suspense>;
}
