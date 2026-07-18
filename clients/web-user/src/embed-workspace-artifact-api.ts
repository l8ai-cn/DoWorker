type EmbeddedRequest = (path: string, init?: RequestInit) => Promise<Response>;

export async function loadEmbeddedArtifactRepresentation(
  request: EmbeddedRequest,
  sessionPath: string,
  input: {
    artifactId: string;
    digest: string;
    representationId: string;
    resourceId: string;
    revision: bigint;
  },
): Promise<Blob> {
  requireSessionFileResource(input.resourceId);
  if (!input.artifactId || !input.representationId || !input.digest) {
    throw new Error("artifact_identity_missing");
  }
  const query = new URLSearchParams({
    artifact_id: input.artifactId,
    digest: input.digest,
    representation_id: input.representationId,
    revision: input.revision.toString(),
  });
  const response = await request(
    `${sessionPath}/artifacts/content?${query.toString()}`,
  );
  if (!response.ok) {
    throw new Error(`Embedded session request failed (${response.status})`);
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
