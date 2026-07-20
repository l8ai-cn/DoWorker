import type {
  AgentArtifactItem,
  AgentSessionSnapshot,
} from "@do-worker/agent-ui";

export function mergeWorkspaceArtifactDiscoveryError(
  snapshot: AgentSessionSnapshot,
  error: string | null,
): AgentSessionSnapshot {
  if (!error) return snapshot;
  return {
    ...snapshot,
    items: [
      ...snapshot.items,
      {
        id: "workspace-artifact-discovery",
        kind: "system" as const,
        title: "Workspace artifact discovery failed",
        detail: error,
        status: "failed" as const,
      },
    ],
  };
}

export function prepareWorkspaceArtifacts(
  artifacts: readonly AgentArtifactItem[],
): AgentArtifactItem[] {
  return artifacts.map((artifact) => {
    const path = workspaceArtifactPath(artifact.artifactId);
    if (!path || !artifact.mimeType) return artifact;
    return {
      ...artifact,
      grants: [{
        actions: ["artifact.download"],
        grantId: `workspace-file:${path}`,
        representationIds: ["workspace-file"],
      }],
      representations: [{
        filename: path,
        mediaType: artifact.mimeType,
        representationId: "workspace-file",
        revision: artifact.revision,
        status: "ready",
      }],
      selectedRepresentationId: "workspace-file",
    };
  });
}

export function workspaceArtifactPath(artifactId: string): string | null {
  const path = artifactId.startsWith("workspace:")
    ? artifactId.slice("workspace:".length)
    : "";
  if (!path || path.split("/").some((segment) => segment === "..")) {
    return null;
  }
  return path;
}
