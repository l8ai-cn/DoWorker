import { useEffect, useState } from "react";
import type { AgentArtifactItem } from "@do-worker/agent-ui";

import { listPodWorkspaceArtifacts } from "@/lib/api/podWorkspaceArtifactApi";
import { prepareWorkspaceArtifacts } from "./webAgentWorkbenchWorkspaceArtifacts";

interface WorkspaceArtifactState {
  artifacts: readonly AgentArtifactItem[];
  error: string | null;
  podKey: string | null;
}

const emptyState: WorkspaceArtifactState = {
  artifacts: [],
  error: null,
  podKey: null,
};

export function usePodWorkspaceArtifacts(
  podKey: string,
  enabled: boolean,
): WorkspaceArtifactState {
  const [state, setState] = useState<WorkspaceArtifactState>(emptyState);

  useEffect(() => {
    if (!enabled) return;
    let active = true;
    void listPodWorkspaceArtifacts(podKey)
      .then((artifacts) => {
        if (active) {
          setState({
            artifacts: prepareWorkspaceArtifacts(artifacts),
            error: null,
            podKey,
          });
        }
      })
      .catch((cause: unknown) => {
        if (!active) return;
        setState({
          artifacts: [],
          error: cause instanceof Error ? cause.message : String(cause),
          podKey,
        });
      });
    return () => {
      active = false;
    };
  }, [enabled, podKey]);

  return enabled && state.podKey === podKey ? state : emptyState;
}
