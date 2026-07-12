import { readLightAuthToken, resolveLightBaseUrl } from "@/lib/light-session";

export interface MarketplaceListingDetail {
  listing_id: string;
  listing_version_id: string;
  slug: string;
  resource_type: string;
  display_name: string;
  tagline: string;
  description: string;
  outcomes: string[];
  requirements: string[];
  permissions: string[];
  version: string;
  publisher: {
    display_name: string;
    verified: boolean;
  };
}

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
): Promise<InstallationPlan> {
  return marketplaceRequest(
    `/markets/${encodeURIComponent(marketSlug)}/listings/${encodeURIComponent(listingSlug)}/plans`,
    {
      method: "POST",
      body: JSON.stringify({
        listing_version_id: listingVersionID,
        target_platform_organization_id: String(organizationID),
        requested_configuration: {},
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

async function marketplaceRequest<T>(
  path: string,
  init: RequestInit = {},
): Promise<T> {
  const baseURL = resolveLightBaseUrl();
  const token = readLightAuthToken(baseURL);
  const response = await fetch(`${baseURL}/api/marketplace/v1${path}`, {
    ...init,
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...init.headers,
    },
  });
  const payload = await response.json().catch(() => null);
  if (!response.ok) {
    const error = payload?.error;
    throw new MarketplaceAcquireError(
      error?.code ?? "MARKETPLACE_REQUEST_FAILED",
      error?.message ?? "市场服务暂时不可用",
    );
  }
  return payload as T;
}

export class MarketplaceAcquireError extends Error {
  constructor(
    public readonly code: string,
    message: string,
  ) {
    super(message);
    this.name = "MarketplaceAcquireError";
  }
}
