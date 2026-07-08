import { getApiBaseUrl } from "@/lib/env";
import { getAuthManager } from "@/lib/wasm-core";
import { readCurrentOrg } from "@/stores/auth";

export type ModelConfig = {
  id: number;
  name: string;
  provider_type: string;
  model: string;
  scope: string;
  token_budget?: number;
};

export type VirtualKey = {
  id: number;
  name: string;
  key_prefix: string;
  ai_model_id: number;
  token_budget?: number;
  status: string;
  last_used_at?: string;
  created_at: string;
};

export type TokenQuota = {
  id: number;
  user_id?: number;
  model?: string;
  limit_tokens: number;
  period: string;
};

export type ScopeUsage = {
  user_id?: number;
  model?: string;
  virtual_api_key_id?: number;
  tokens: number;
  cost_usd: number;
  limit_tokens?: number;
  over_limit: boolean;
};

export type QuotaReport = {
  total_tokens: number;
  total_cost_usd: number;
  by_user: ScopeUsage[];
  by_model: ScopeUsage[];
  by_virtual_key: ScopeUsage[];
  quotas: ScopeUsage[];
};

function headers(): HeadersInit | null {
  const token = getAuthManager().get_token();
  const org = readCurrentOrg()?.slug;
  if (!token || !org) return null;
  return {
    Authorization: `Bearer ${token}`,
    "X-Organization-Slug": org,
    "Content-Type": "application/json",
  };
}

async function req<T>(path: string, init?: RequestInit): Promise<T> {
  const h = headers();
  if (!h) throw new Error("not authenticated");
  const base = getApiBaseUrl().replace(/\/$/, "");
  const res = await fetch(`${base}/v1${path}`, { ...init, headers: h });
  if (!res.ok) {
    const body = await res.text();
    throw new Error(body || `request failed: ${res.status}`);
  }
  if (res.status === 204) return undefined as T;
  return res.json() as Promise<T>;
}

export async function listModelConfigs(): Promise<ModelConfig[]> {
  const data = await req<{ data: ModelConfig[] }>("/model-configs");
  return data.data ?? [];
}

export async function listVirtualKeys(): Promise<VirtualKey[]> {
  const data = await req<{ data: VirtualKey[] }>("/virtual-keys");
  return data.data ?? [];
}

export async function createVirtualKey(input: {
  name: string;
  ai_model_id: number;
  token_budget?: number;
}): Promise<{ token: string; key: VirtualKey }> {
  return req("/virtual-keys", { method: "POST", body: JSON.stringify(input) });
}

export async function revokeVirtualKey(id: number): Promise<void> {
  await req(`/virtual-keys/${id}`, { method: "DELETE" });
}

export async function listTokenQuotas(): Promise<TokenQuota[]> {
  const data = await req<{ data: TokenQuota[] }>("/token-quotas");
  return data.data ?? [];
}

export async function upsertTokenQuota(input: {
  user_id?: number | null;
  model?: string | null;
  limit_tokens: number;
  period?: string;
}): Promise<void> {
  await req("/token-quotas", { method: "PUT", body: JSON.stringify(input) });
}

export async function deleteTokenQuota(id: number): Promise<void> {
  await req(`/token-quotas/${id}`, { method: "DELETE" });
}

export async function getQuotaReport(): Promise<QuotaReport> {
  return req("/usage/quota-report");
}
