import { getApiBaseUrl } from "@/lib/env";
import { getAuthManager } from "@/lib/wasm-core";
import { readCurrentOrg } from "@/stores/auth";

export async function deleteTerminalPod(podKey: string): Promise<void> {
  const token = getAuthManager().get_token();
  const orgSlug = readCurrentOrg()?.slug;
  if (!token || !orgSlug) throw new Error("not authenticated");
  const base = getApiBaseUrl().replace(/\/$/, "");
  const response = await fetch(`${base}/v1/orgs/${encodeURIComponent(orgSlug)}/pods/${encodeURIComponent(podKey)}`, {
    method: "DELETE",
    headers: { Authorization: `Bearer ${token}`, "X-Organization-Slug": orgSlug },
  });
  if (!response.ok) throw new Error((await response.text()) || `request failed: ${response.status}`);
}
