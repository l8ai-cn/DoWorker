import { getApiBaseUrl } from "@/lib/env";
import { getAuthManager } from "@/lib/wasm-core";
import { readCurrentOrg } from "@/stores/auth";

export async function loadSessionArtifactRepresentation(
  input: {
    artifactId: string;
    digest: string;
    representationId: string;
    resourceId: string;
    revision: bigint;
    sessionId: string;
  },
): Promise<Blob> {
  requireSessionFileResource(input.resourceId);
  if (!input.artifactId || !input.representationId || !input.digest) {
    throw new Error("artifact_identity_missing");
  }
  const token = getAuthManager().get_token();
  const org = readCurrentOrg()?.slug;
  if (!token || !org) throw new Error("Not authenticated");
  const base = getApiBaseUrl().replace(/\/$/, "");
  const query = new URLSearchParams({
    artifact_id: input.artifactId,
    digest: input.digest,
    representation_id: input.representationId,
    revision: input.revision.toString(),
  });
  const response = await fetch(
    `${base}/v1/sessions/${encodeURIComponent(input.sessionId)}` +
      `/artifacts/content?${query.toString()}`,
    {
      headers: {
        Authorization: `Bearer ${token}`,
        "X-Organization-Slug": org,
      },
    },
  );
  if (!response.ok) {
    throw new Error(`Artifact request failed (${response.status})`);
  }
  return response.blob();
}

function requireSessionFileResource(resourceId: string): void {
  const fileID = resourceId.startsWith("session-file:")
    ? resourceId.slice("session-file:".length)
    : "";
  if (!fileID) {
    throw new Error(`artifact_resource_unsupported:${resourceId}`);
  }
}
