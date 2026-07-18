type EmbeddedRequest = (path: string, init?: RequestInit) => Promise<Response>;

export interface EmbeddedArtifactIdentity {
  artifactId: string;
  representationId: string;
  revision: bigint;
}

export async function loadEmbeddedWorkspaceArtifact(
  request: EmbeddedRequest,
  sessionPath: string,
  path: string,
  identity: EmbeddedArtifactIdentity,
): Promise<Blob> {
  if (!path) throw new Error("Workspace artifact path is empty");
  const query = new URLSearchParams({
    artifact_id: identity.artifactId,
    representation_id: identity.representationId,
    revision: identity.revision.toString(),
  });
  const response = await request(
    `${sessionPath}/resources/environments/workspace/artifacts/content/${encodePath(path)}?${query}`,
  );
  if (!response.ok) {
    throw new Error(`Embedded session request failed (${response.status})`);
  }
  return response.blob();
}

function encodePath(path: string): string {
  return path.split("/").map(encodeURIComponent).join("/");
}
