import { marketplaceRequest } from "./client";

export const DEFAULT_MARKET_SLUG = "do-worker-market";

export type MarketplaceResourceType =
  | "application"
  | "skill"
  | "mcp_connector"
  | "resource";

export interface MarketplaceSpace {
  slug: string;
  name: string;
}

export interface MarketplaceListingSummary {
  listing_id: string;
  listing_version_id: string;
  slug: string;
  resource_type: MarketplaceResourceType;
  display_name: string;
  tagline: string;
  publisher: {
    display_name: string;
    verified: boolean;
  };
  spaces: MarketplaceSpace[];
  quota?: {
    mode: string;
    estimated_credits_micro: string;
  };
}

export interface MarketplaceListingDetail extends MarketplaceListingSummary {
  description: string;
  outcomes: string[];
  use_cases: string[];
  target_audience: string[];
  requirements: string[];
  permissions: string[];
  version: string;
  release_notes: string;
  documentation_url?: string;
  support_url?: string;
}

export interface MarketplaceSummary {
  name: string;
  summary: string;
}

export function fetchMarketplaceSummary(): Promise<MarketplaceSummary> {
  return marketplaceRequest(`/markets/${DEFAULT_MARKET_SLUG}`);
}

export async function fetchMarketplaceListings(): Promise<MarketplaceListingSummary[]> {
  const response = await marketplaceRequest<{ items: MarketplaceListingSummary[] }>(
    `/markets/${DEFAULT_MARKET_SLUG}/listings`,
  );
  return response.items;
}

export function fetchMarketplaceListingDetail(
  listingSlug: string,
): Promise<MarketplaceListingDetail> {
  return marketplaceRequest(
    `/markets/${DEFAULT_MARKET_SLUG}/listings/${encodeURIComponent(listingSlug)}`,
  );
}
