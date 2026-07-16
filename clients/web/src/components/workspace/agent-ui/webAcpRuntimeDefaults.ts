import type { AgentConnectionStatus } from "@do-worker/agent-ui";

import { dispatchAcpRelayEvent } from "@/stores/acpEventDispatcher";
import {
  listPodWorkspaceArtifacts,
  loadPodWorkspaceArtifact,
} from "@/lib/api/podWorkspaceArtifactApi";
import { readAcpSession, useAcpSessionStore } from "@/stores/acpSession";
import { relayPool, type RelayStatusInfo } from "@/stores/relayConnection";
import type { WebAcpRuntimeDeps } from "./webAcpRuntimeTypes";

export const defaultWebAcpRuntimeDeps: WebAcpRuntimeDeps = {
  relay: relayPool,
  readSession: readAcpSession,
  subscribeSession: (listener) => useAcpSessionStore.subscribe(listener),
  dispatchRelayEvent: dispatchAcpRelayEvent,
  listWorkspaceArtifacts: listPodWorkspaceArtifacts,
  loadWorkspaceArtifact: loadPodWorkspaceArtifact,
  removePermission: (podKey, permissionId) =>
    useAcpSessionStore.getState().removePermissionRequest(podKey, permissionId),
};

export function relayConnection(
  status: RelayStatusInfo,
): AgentConnectionStatus {
  if (status.status === "connected") return "connected";
  if (status.status === "connecting") return "connecting";
  if (status.status === "disconnected" || status.status === "error") {
    return "disconnected";
  }
  return "reconnecting";
}
