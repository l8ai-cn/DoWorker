import type { ArtifactDescriptor } from "@do-worker/proto/agent_workbench/v2/artifact_pb";
import { createAgentArtifactLoader, type AgentArtifactLoadRequest } from "@do-worker/agent-ui";

import type { EmbedSessionClient } from "@/embed-session-api";

interface ArtifactConnection {
  getStore(): {
    getState(): {
      snapshot: {
        artifacts: readonly ArtifactDescriptor[];
        sessionId: string;
      };
    } | null;
  };
}

type ArtifactResources = Pick<EmbedSessionClient, "loadDownload" | "loadResource">;

export function createEmbeddedArtifactLoader(
  connection: ArtifactConnection,
  resources: ArtifactResources,
): (request: AgentArtifactLoadRequest) => Promise<Blob> {
  return createAgentArtifactLoader({
    getArtifacts(sessionId) {
      const state = connection.getStore().getState();
      if (!state || state.snapshot.sessionId !== sessionId) {
        throw new Error("artifact_session_mismatch");
      }
      return state.snapshot.artifacts;
    },
    loadDownload: (url) => resources.loadDownload(url),
    loadResource: (resourceId, context) =>
      resources.loadResource(resourceId, context),
  });
}
