import {
  type AgentArtifactItem,
  workspaceFileArtifacts,
} from "@agent-cloud/agent-ui";

import { getApiBaseUrl } from "@/lib/env";
import { getAuthManager } from "@/lib/wasm-core";
import { readCurrentOrg } from "@/stores/auth";

export async function listPodWorkspaceArtifacts(
  podKey: string,
): Promise<AgentArtifactItem[]> {
  const response = await podWorkspaceFetch(
    podKey,
    "/resources/workspace/changes",
  );
  const body = (await response.json()) as { data?: unknown };
  return workspaceFileArtifacts("workspace-discovery", body.data);
}

export async function loadPodWorkspaceArtifact(
  podKey: string,
  path: string,
): Promise<Blob> {
  const encodedPath = path.split("/").map(encodeURIComponent).join("/");
  const response = await podWorkspaceFetch(
    podKey,
    `/resources/workspace/filesystem/${encodedPath}`,
  );
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

async function podWorkspaceFetch(
  podKey: string,
  suffix: string,
): Promise<Response> {
  const token = getAuthManager().get_token();
  const org = readCurrentOrg()?.slug;
  if (!token || !org) {
    throw new Error("Not authenticated");
  }
  const base = getApiBaseUrl().replace(/\/$/, "");
  const apiRoot = base.endsWith("/api") ? base : `${base}/api`;
  const response = await fetch(
    `${apiRoot}/v1/orgs/${encodeURIComponent(org)}/pods/${encodeURIComponent(podKey)}${suffix}`,
    {
      cache: "no-store",
      headers: { Authorization: `Bearer ${token}` },
    },
  );
  if (!response.ok) {
    throw new Error(`Workspace artifact request failed (${response.status})`);
  }
  return response;
}
