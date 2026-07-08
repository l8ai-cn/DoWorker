import { lightFetch } from "@/lib/light-auth/api-fetch";
import { readCurrentOrg } from "@/stores/auth";

export interface QuickTaskInput {
  prompt: string;
  agent_slug?: string;
  runner_id?: number;
  repository_id?: number;
  alias?: string;
  queue_ttl_minutes?: number;
}

export interface QuickTaskResult {
  pod_key: string;
  status: string;
  queue_position?: number;
  expires_at?: string;
}

export const quickTaskApi = {
  create: async (input: QuickTaskInput): Promise<QuickTaskResult> => {
    const slug = readCurrentOrg()?.slug ?? "";
    return lightFetch<QuickTaskResult>(`/api/v1/orgs/${slug}/quick-tasks`, {
      method: "POST",
      body: input,
      authenticated: true,
    });
  },
};
