import { Suspense } from "react";

import { MarketplaceAcquireFlow } from "@/components/marketplace/acquire/MarketplaceAcquireFlow";

export default function MarketplaceAcquirePage() {
  return (
    <Suspense fallback={<AcquirePageLoading />}>
      <MarketplaceAcquireFlow />
    </Suspense>
  );
}

function AcquirePageLoading() {
  return (
    <main className="min-h-screen bg-surface px-4 py-16">
      <div className="mx-auto max-w-3xl animate-pulse rounded-xl border border-border bg-card p-8">
        <div className="h-5 w-28 rounded bg-muted" />
        <div className="mt-5 h-9 w-2/3 rounded bg-muted" />
        <div className="mt-8 h-40 rounded bg-muted" />
      </div>
    </main>
  );
}
