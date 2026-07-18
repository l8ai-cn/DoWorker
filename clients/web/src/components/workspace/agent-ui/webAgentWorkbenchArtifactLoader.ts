import {
  createAgentArtifactLoader,
  type AgentArtifactLoadRequest,
} from "@do-worker/agent-ui";

import { loadSessionArtifactRepresentation } from "@/lib/api/sessionWorkspaceArtifactApi";
import { decodeWebAgentWorkbenchSnapshot } from "./webAgentWorkbenchProjection";
import type { WebAgentWorkbenchState } from "./webAgentWorkbenchRuntimeTypes";

export function createWebAgentWorkbenchArtifactLoader(
  state: WebAgentWorkbenchState,
  fetcher: typeof globalThis.fetch = globalThis.fetch,
): (request: AgentArtifactLoadRequest) => Promise<Blob> {
  return createAgentArtifactLoader({
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
}
