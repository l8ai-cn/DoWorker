import { lightFetch } from "@/lib/light-auth/api-fetch";

export interface PodPreviewSession {
  preview_base_url: string;
  session_url: string;
  expires_at: string;
}

interface PodPreviewSessionWire extends PodPreviewSession {
  token?: string;
}

export async function getPodPreviewSession(
  orgSlug: string,
  podKey: string,
): Promise<PodPreviewSession> {
  const data = await lightFetch<PodPreviewSessionWire>(
    `/api/v1/orgs/${encodeURIComponent(orgSlug)}/pods/${encodeURIComponent(podKey)}/preview`,
    { authenticated: true },
  );
  return {
    preview_base_url: data.preview_base_url,
    session_url: data.session_url,
    expires_at: data.expires_at,
  };
}
