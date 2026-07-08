// Model pool API — org/user model configs for Worker launch.
// Wire: GET/POST/DELETE /v1/model-configs

import { authenticatedFetch } from "./identity";

export interface ModelConfig {
  id: number;
  name: string;
  provider_type: string;
  model: string;
  base_url: string;
  is_default: boolean;
  scope: "org" | "user";
  token_budget?: number | null;
}

interface ListResponse {
  object: "list";
  data: ModelConfig[];
}

export async function listModelConfigs(): Promise<ModelConfig[]> {
  const res = await authenticatedFetch("/v1/model-configs");
  if (!res.ok) return [];
  const json = (await res.json()) as ListResponse;
  return json.data ?? [];
}

export interface CreateModelConfigInput {
  name: string;
  provider_type: string;
  model: string;
  base_url?: string;
  credentials: Record<string, string>;
  is_default?: boolean;
  token_budget?: number | null;
  scope?: "org" | "user";
}

export async function createModelConfig(input: CreateModelConfigInput): Promise<ModelConfig> {
  const res = await authenticatedFetch("/v1/model-configs", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
  if (!res.ok) {
    const body = (await res.json().catch(() => ({}))) as { error?: string };
    throw new Error(body.error ?? `Failed to create model (${res.status})`);
  }
  return (await res.json()) as ModelConfig;
}

export async function deleteModelConfig(id: number): Promise<void> {
  const res = await authenticatedFetch(`/v1/model-configs/${id}`, { method: "DELETE" });
  if (!res.ok && res.status !== 404) {
    throw new Error(`Failed to delete model (${res.status})`);
  }
}
