import { lightFetch } from "@/lib/light-auth/api-fetch";
import { readCurrentOrg } from "@/stores/auth";

export type ExpertMarketReleaseStatus =
  | "draft"
  | "pending_review"
  | "published"
  | "rejected"
  | "withdrawn";

export interface ExpertMarketRelease {
  id: number;
  application_id: number;
  application_slug: string;
  source_expert_id: number;
  version: number;
  status: ExpertMarketReleaseStatus;
  name: string;
  summary: string;
  description: string;
  category: string;
  icon: string;
  tags: string[];
  outcomes: string[];
  rejection_reason?: string;
  created_at: string;
}

export interface SubmitExpertMarketReleaseInput {
  slug: string;
  summary: string;
  description: string;
  category: string;
  icon: string;
  tags: string[];
  outcomes: string[];
}

function marketBase(): string {
  return `/api/v1/orgs/${readCurrentOrg()?.slug ?? ""}`;
}

export async function listExpertMarketSubmissions(): Promise<{
  releases: ExpertMarketRelease[];
  total: number;
}> {
  const releases: ExpertMarketRelease[] = [];
  let total = 0;

  do {
    const response = await lightFetch<{
      releases?: ExpertMarketRelease[];
      total?: number;
    }>(`${marketBase()}/marketplace/submissions`, {
      authenticated: true,
      query: { limit: 100, offset: releases.length },
    });
    const page = response.releases ?? [];
    total = response.total ?? 0;
    releases.push(...page);
    if (page.length === 0 && releases.length < total) {
      throw new Error("Marketplace submissions pagination ended before total");
    }
  } while (releases.length < total);

  return {
    releases,
    total,
  };
}

export async function submitExpertMarketRelease(
  expertSlug: string,
  input: SubmitExpertMarketReleaseInput,
): Promise<void> {
  await lightFetch(`${marketBase()}/experts/${expertSlug}/market-submissions`, {
    method: "POST",
    authenticated: true,
    body: input,
  });
}

export async function withdrawExpertMarketRelease(releaseID: number): Promise<void> {
  await lightFetch(`${marketBase()}/marketplace/releases/${releaseID}/withdraw`, {
    method: "POST",
    authenticated: true,
  });
}

export async function getExpertMarketUpgrade(
  expertSlug: string,
): Promise<{ upgrade_available: boolean }> {
  return lightFetch(`${marketBase()}/experts/${expertSlug}/market-upgrade`, {
    authenticated: true,
  });
}

export async function upgradeExpertFromMarket(
  expertSlug: string,
): Promise<{ upgraded: boolean }> {
  return lightFetch(`${marketBase()}/experts/${expertSlug}/market-upgrade`, {
    method: "POST",
    authenticated: true,
  });
}
