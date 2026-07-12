import { apiFetch } from "./api-fetch";
import { readOrgSlug } from "./auth-store";

export type MobilePodPreviewSession = {
  sessionUrl: string;
};

function orgSlug(): string {
  const slug = readOrgSlug();
  if (!slug) throw new Error("当前登录未选择组织");
  return slug;
}

export async function getMobilePodPreviewSession(
  podKey: string,
): Promise<MobilePodPreviewSession> {
  const response = await apiFetch(
    `/api/v1/orgs/${encodeURIComponent(orgSlug())}/pods/${encodeURIComponent(podKey)}/preview`,
  );
  if (!response.ok) {
    throw new Error((await response.text()) || `Preview 请求失败 (${response.status})`);
  }
  const body = (await response.json()) as { session_url?: unknown };
  if (typeof body.session_url !== "string" || body.session_url.length === 0) {
    throw new Error("Preview session URL 无效");
  }
  return { sessionUrl: body.session_url };
}

export async function replaceWithMobilePodPreview(
  podKey: string,
  replace: (url: string) => void,
): Promise<void> {
  const { sessionUrl } = await getMobilePodPreviewSession(podKey);
  replace(sessionUrl);
}
