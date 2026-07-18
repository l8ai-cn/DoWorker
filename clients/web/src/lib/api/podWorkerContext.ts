import { lightFetch } from "@/lib/light-auth/api-fetch";

export interface PodWorkerContext {
  snapshot_id: number;
  alias: string;
  expert?: {
    id: number;
    slug: string;
    name: string;
  };
  skill_slugs: string[];
}

export async function getPodWorkerContext(
  orgSlug: string,
  podKey: string,
): Promise<PodWorkerContext> {
  const response = await lightFetch<{ worker: PodWorkerContext }>(
    `/api/v1/orgs/${orgSlug}/pods/${podKey}/worker-context`,
    { authenticated: true },
  );
  return response.worker;
}
