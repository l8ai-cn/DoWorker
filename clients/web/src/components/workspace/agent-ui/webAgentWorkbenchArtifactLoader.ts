import {
  createAgentArtifactLoader,
  type AgentArtifactLoadRequest,
} from "@do-worker/agent-ui";

import { loadSessionWorkspaceArtifactById } from "@/lib/api/sessionWorkspaceArtifactApi";
import {
  decodeWebAgentWorkbenchSnapshot,
} from "./webAgentWorkbenchProjection";
import type {
  WebAgentWorkbenchState,
} from "./webAgentWorkbenchRuntimeTypes";

export function createWebAgentWorkbenchArtifactLoader(
  state: WebAgentWorkbenchState,
  fetcher: typeof globalThis.fetch = globalThis.fetch,
): (request: AgentArtifactLoadRequest) => Promise<Blob> {
  return createAgentArtifactLoader({
    getArtifacts: (sessionId) =>
      decodeWebAgentWorkbenchSnapshot(
        state.snapshotBytes(sessionId),
        sessionId,
      )?.artifacts ?? [],
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
    loadResource: (resourceId, context) => {
      if (!resourceId.startsWith("workspace:")) {
        throw new Error(`artifact_resource_unsupported:${resourceId}`);
      }
      return loadSessionWorkspaceArtifactById(
        context.sessionId,
        resourceId.slice("workspace:".length),
        {
          artifactId: context.artifactId,
          representationId: context.representationId!,
          revision: context.representation.revision,
        },
      );
    },
  });
}
