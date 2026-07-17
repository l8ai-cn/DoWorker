import { getApiBaseUrl } from "@/lib/env";
import { getAuthManager } from "@/lib/wasm-core";
import { readCurrentOrg } from "@/stores/auth";
import { fetchSessionByPodKey } from "./sessionImportApi";

export async function loadSessionWorkspaceArtifact(
  podKey: string,
  path: string,
): Promise<Blob> {
  const session = await fetchSessionByPodKey(podKey);
  if (!session) throw new Error("No session is linked to this Worker");
  return loadSessionWorkspaceArtifactById(session.id, path);
}

export async function loadSessionWorkspaceArtifactById(
  sessionId: string,
  path: string,
): Promise<Blob> {
  const token = getAuthManager().get_token();
  const org = readCurrentOrg()?.slug;
  if (!token || !org) throw new Error("Not authenticated");
  const encodedPath = path.split("/").map(encodeURIComponent).join("/");
  const base = getApiBaseUrl().replace(/\/$/, "");
  const response = await fetch(
    `${base}/v1/sessions/${encodeURIComponent(sessionId)}` +
      `/resources/environments/workspace/artifacts/content/${encodedPath}`,
    {
      headers: {
        Authorization: `Bearer ${token}`,
        "X-Organization-Slug": org,
      },
    },
  );
  if (!response.ok) {
    throw new Error(`Workspace artifact request failed (${response.status})`);
  }
  return response.blob();
}
