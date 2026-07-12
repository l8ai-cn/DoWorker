import type { ListingSummary } from "@/lib/marketplace-types";

import { ListingCard } from "./listing-card";

export function ListingGrid({ listings }: { listings: ListingSummary[] }) {
  return (
    <div className="listing-grid">
      {listings.map((listing) => (
        <ListingCard key={listing.listing_id} listing={listing} />
      ))}
    </div>
  );
}
