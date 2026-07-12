import { marketplaceRequest } from "./client";
import type { MarketplaceResourceType } from "./catalog-api";

export interface MarketplaceOrganizationApplication {
  installation_id: string;
  market_slug: string;
  listing_slug: string;
  display_name: string;
  tagline: string;
  resource_type: MarketplaceResourceType;
  outcomes: string[];
  runtime_ref: string;
  status: "installing" | "verifying" | "active";
  installed_at: string;
}

export async function fetchOrganizationApplications(
  organizationID: number,
): Promise<MarketplaceOrganizationApplication[]> {
  const response = await marketplaceRequest<{
    applications: MarketplaceOrganizationApplication[];
  }>(`/organizations/${organizationID}/applications`);
  return response.applications;
}

export function expertIDFromRuntimeRef(runtimeRef: string): number | null {
  const match = /^expert:(\d+)$/.exec(runtimeRef);
  if (!match) return null;
  const expertID = Number(match[1]);
  return Number.isSafeInteger(expertID) && expertID > 0 ? expertID : null;
}
