import type {
  ListingCollection,
  ListingDetail,
  ListingQuery,
  Market,
  MarketplaceErrorEnvelope,
} from "./marketplace-types";

const MARKET_SLUG = "do-worker-market";
const DEFAULT_API_URL = "http://marketplace:8080";
const DEFAULT_REQUEST_HOST = "market.l8ai.cn";

export class MarketplaceApiError extends Error {
  constructor(
    readonly code: string,
    message: string,
    readonly detail?: string,
  ) {
    super(message);
  }
}

function marketplaceUrl(path: string): string {
  const baseUrl = (process.env.MARKETPLACE_API_INTERNAL_URL || DEFAULT_API_URL)
    .replace(/\/+$/, "");
  return `${baseUrl}/api/marketplace/v1/markets/${MARKET_SLUG}${path}`;
}

async function request<T>(path: string): Promise<T> {
  const response = await fetch(marketplaceUrl(path), {
    headers: {
      "X-Forwarded-Host":
        process.env.MARKETPLACE_REQUEST_HOST || DEFAULT_REQUEST_HOST,
    },
    cache: "no-store",
  });
  if (response.ok) {
    return response.json() as Promise<T>;
  }

  const body = (await response.json().catch(() => ({}))) as MarketplaceErrorEnvelope;
  throw new MarketplaceApiError(
    body.error?.code || "INTERNAL_ERROR",
    body.error?.message || "市场服务暂时不可用",
    body.error?.detail,
  );
}

export function getMarket(): Promise<Market> {
  return request<Market>("");
}

function listingQueryPath(query: ListingQuery): string {
  const params = new URLSearchParams();
  const values = Object.entries(query) as Array<[keyof ListingQuery, string]>;
  values.forEach(([key, value]) => {
    if (value) params.set(key, value);
  });
  return `/listings?${params.toString()}`;
}

export function listListings(query: ListingQuery): Promise<ListingCollection> {
  return request<ListingCollection>(listingQueryPath(query));
}

export function getListing(slug: string): Promise<ListingDetail> {
  return request<ListingDetail>(`/listings/${encodeURIComponent(slug)}`);
}

export { MARKET_SLUG };
