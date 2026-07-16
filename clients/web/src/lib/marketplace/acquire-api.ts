import { marketplaceRequest, MarketplaceRequestError } from "./client";
import type { MarketplaceListingDetail } from "./catalog-api";

export type { MarketplaceListingDetail } from "./catalog-api";

export interface InstallationPlan {
  installation_id: string;
  operation_id: string;
  plan: {
    plan_id: string;
    plan_digest: string;
    expires_at: string;
    listing_version_id: string;
    estimated_credits_micro: string;
    required_permissions: string[];
  };
}

export interface InstallationResult {
  installation_id: string;
  operation_id: string;
  status: "planned" | "running" | "succeeded" | "failed";
  stage: string;
  runtime_ref?: string;
}

export async function fetchMarketplaceListing(
  marketSlug: string,
  listingSlug: string,
): Promise<MarketplaceListingDetail> {
  return marketplaceRequest(
    `/markets/${encodeURIComponent(marketSlug)}/listings/${encodeURIComponent(listingSlug)}`,
  );
}

export async function createInstallationPlan(
  marketSlug: string,
  listingSlug: string,
  listingVersionID: string,
  organizationID: number,
  modelResourceID: number,
  toolModelResourceIDs: Record<string, number>,
): Promise<InstallationPlan> {
  return marketplaceRequest(
    `/markets/${encodeURIComponent(marketSlug)}/listings/${encodeURIComponent(listingSlug)}/plans`,
    {
      method: "POST",
      body: JSON.stringify({
        listing_version_id: listingVersionID,
        target_platform_organization_id: String(organizationID),
        requested_configuration: {
          model_resource_id: modelResourceID,
          tool_model_resource_ids: toolModelResourceIDs,
        },
      }),
    },
  );
}

export async function applyInstallationPlan(
  plan: InstallationPlan,
): Promise<InstallationResult> {
  return marketplaceRequest(
    `/installation-operations/${encodeURIComponent(plan.operation_id)}/apply`,
    {
      method: "POST",
      headers: { "Idempotency-Key": plan.operation_id },
      body: JSON.stringify({
        plan_id: plan.plan.plan_id,
        plan_digest: plan.plan.plan_digest,
      }),
    },
  );
}

export { MarketplaceRequestError as MarketplaceAcquireError };
