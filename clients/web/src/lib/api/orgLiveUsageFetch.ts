import { getApiBaseUrl } from "@/lib/env";
import { getAuthManager } from "@/lib/wasm-core";
import { readCurrentOrg } from "@/stores/auth";

export type OrgLiveModelUsage = {
  model: string;
  input_tokens: number;
  output_tokens: number;
  cache_read_tokens?: number;
  cache_creation_tokens?: number;
  total_cost_usd?: number;
};

export type OrgLiveUsageSummary = {
  object?: string;
  total_cost_usd?: number;
  usage_by_model?: Record<string, OrgLiveModelUsage>;
};

export async function fetchOrgLiveUsageSummary(): Promise<OrgLiveUsageSummary | null> {
  const token = getAuthManager().get_token();
  const org = readCurrentOrg()?.slug;
  if (!token || !org) return null;

  const base = getApiBaseUrl().replace(/\/$/, "");
  const res = await fetch(`${base}/v1/org/usage/summary`, {
    headers: {
      Authorization: `Bearer ${token}`,
      "X-Organization-Slug": org,
    },
  });
  if (!res.ok) return null;
  return res.json() as Promise<OrgLiveUsageSummary>;
}
