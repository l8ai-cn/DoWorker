type EmbeddedRequest = (path: string, init?: RequestInit) => Promise<Response>;

export async function loadEmbeddedWorkspaceArtifact(
  request: EmbeddedRequest,
  sessionPath: string,
  path: string,
): Promise<Blob> {
  if (!path) throw new Error("Workspace artifact path is empty");
  const response = await request(
    `${sessionPath}/resources/environments/workspace/artifacts/content/${encodePath(path)}`,
  );
  if (!response.ok) {
    throw new Error(`Embedded session request failed (${response.status})`);
  }
  return response.blob();
}

function encodePath(path: string): string {
  return path.split("/").map(encodeURIComponent).join("/");
}
