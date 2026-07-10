// AI resource API — org/user model resources for Worker launch.

import { authenticatedFetch } from "./identity";

export interface ModelConfig {
  id: number;
  name: string;
  provider_key: string;
  model: string;
  is_default: boolean;
  token_budget?: number | null;
}

interface ListResponse {
  object: "list";
  data: ModelConfig[];
}

export async function listModelResources(): Promise<ModelConfig[]> {
  const res = await authenticatedFetch("/v1/model-resources");
  if (!res.ok) {
    const message = (await res.text()).trim();
    throw new Error(message || `Failed to load AI resources (${res.status})`);
  }
  const json = (await res.json()) as ListResponse;
  return json.data ?? [];
}
