import {
  createAgentArtifactLoader,
  type AgentArtifactLoadRequest,
} from "@agent-cloud/agent-ui";

import { loadPodWorkspaceArtifact } from "@/lib/api/podWorkspaceArtifactApi";
import { loadSessionArtifactRepresentation } from "@/lib/api/sessionWorkspaceArtifactApi";
import { decodeWebAgentWorkbenchSnapshot } from "./webAgentWorkbenchProjection";
import type { WebAgentWorkbenchState } from "./webAgentWorkbenchRuntimeTypes";
import { workspaceArtifactPath } from "./webAgentWorkbenchWorkspaceArtifacts";

interface ArtifactLoaderOptions {
  fetcher?: typeof globalThis.fetch;
  podKey?: string;
}

export function createWebAgentWorkbenchArtifactLoader(
  state: WebAgentWorkbenchState,
  options: ArtifactLoaderOptions = {},
): (request: AgentArtifactLoadRequest) => Promise<Blob> {
  const fetcher = options.fetcher ?? globalThis.fetch;
  const loadSessionArtifact = createAgentArtifactLoader({
    getArtifacts: (sessionId) =>
      decodeWebAgentWorkbenchSnapshot(state.snapshotBytes(sessionId), sessionId)
        ?.artifacts ?? [],
    loadDownload: async (url) => {
      const response = await fetcher(url, {
        cache: "no-store",
        credentials: "same-origin",
      });
      if (!response.ok) {
        throw new Error(`artifact_download_failed:${response.status}`);
      }
      return response.blob();
    },
    loadResource: (resourceId, context) =>
      loadSessionArtifactRepresentation({
        artifactId: context.artifactId,
        digest: context.representation.digest ?? "",
        representationId: context.representationId ?? "",
        resourceId,
        revision: context.descriptor.revision,
        sessionId: context.sessionId,
      }),
  });
  return async (request) => {
    const path = workspaceArtifactPath(request.artifactId);
    if (path) {
      if (request.representationId !== "workspace-file") {
        throw new Error("workspace_artifact_representation_invalid");
      }
      if (!options.podKey) throw new Error("workspace_artifact_pod_missing");
      return loadPodWorkspaceArtifact(options.podKey, path);
    }
    return loadSessionArtifact(request);
  };
}
