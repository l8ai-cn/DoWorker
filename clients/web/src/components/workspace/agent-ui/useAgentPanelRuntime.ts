import { useMemo } from "react";

import { getAgentWorkbenchState } from "@/lib/wasm-core";
import { WebAgentWorkbenchRuntime } from "./WebAgentWorkbenchRuntime";
import {
  createWebAgentWorkbenchArtifactLoader,
} from "./webAgentWorkbenchArtifactLoader";

interface AgentPanelRuntimeInput {
  agentLabel?: string;
  interactionMode?: "acp" | "pty";
  live: boolean;
  podKey: string;
  sessionId: string | null;
  title: string;
  workspaceArtifactError: string | null;
}

export function useAgentPanelRuntime(input: AgentPanelRuntimeInput) {
  return useMemo(() => {
    if (!input.sessionId) return null;
    const state = getAgentWorkbenchState();
    return new WebAgentWorkbenchRuntime({
      agentLabel: input.agentLabel ?? "Agent",
      interactionMode: input.interactionMode ?? "acp",
      loadArtifact: createWebAgentWorkbenchArtifactLoader(state, {
        podKey: input.podKey,
      }),
      live: input.live,
      sessionId: input.sessionId,
      title: input.title,
      workspaceArtifactError: input.workspaceArtifactError,
    });
  }, [
    input.agentLabel,
    input.interactionMode,
    input.live,
    input.podKey,
    input.sessionId,
    input.title,
    input.workspaceArtifactError,
  ]);
}
