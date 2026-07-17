import {
  type AgentArtifactItem,
  workspaceFileArtifacts,
} from "@do-worker/agent-ui";

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
    `/resources/workspace/artifacts/${encodedPath}`,
  );
  return response.blob();
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
