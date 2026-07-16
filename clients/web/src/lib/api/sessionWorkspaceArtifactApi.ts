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
  const token = getAuthManager().get_token();
  const org = readCurrentOrg()?.slug;
  if (!token || !org) throw new Error("Not authenticated");
  const encodedPath = path.split("/").map(encodeURIComponent).join("/");
  const base = getApiBaseUrl().replace(/\/$/, "");
  const response = await fetch(
    `${base}/v1/sessions/${encodeURIComponent(session.id)}` +
      `/resources/environments/workspace/filesystem/${encodedPath}`,
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
  const body = (await response.json()) as {
    content?: unknown;
    content_type?: unknown;
    encoding?: unknown;
    truncated?: unknown;
  };
  if (body.truncated === true) {
    throw new Error("Workspace artifact exceeds the preview size limit");
  }
  if (typeof body.content !== "string") {
    throw new Error("Workspace artifact response is invalid");
  }
  const mimeType =
    typeof body.content_type === "string" ? body.content_type : "";
  if (body.encoding === "base64") {
    const decoded = atob(body.content);
    const bytes = Uint8Array.from(decoded, (char) => char.charCodeAt(0));
    return new Blob([bytes], { type: mimeType });
  }
  if (body.encoding !== undefined && body.encoding !== "utf-8") {
    throw new Error("Workspace artifact encoding is unsupported");
  }
  return new Blob([body.content], { type: mimeType });
}
