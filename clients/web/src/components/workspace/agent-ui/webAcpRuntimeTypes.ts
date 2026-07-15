import type { AgentArtifactItem } from "@do-worker/agent-ui";

import type { AcpSessionState } from "@/stores/acpSessionTypes";
import type { RelayStatusInfo } from "@/stores/relayConnection";

export interface RelayAdapter {
  subscribe(
    podKey: string,
    subscriptionId: string,
    onMessage: (data: Uint8Array | string) => void,
  ): Promise<unknown>;
  unsubscribe(podKey: string, subscriptionId: string): void;
  onAcpMessage(
    podKey: string,
    listener: (messageType: number, payload: unknown) => void,
  ): () => void;
  onStatusChange(
    podKey: string,
    listener: (status: RelayStatusInfo) => void,
  ): () => void;
  sendAcpCommand(
    podKey: string,
    command: Record<string, unknown>,
  ): Promise<void>;
}

export interface WebAcpRuntimeDeps {
  relay: RelayAdapter;
  readSession: (podKey: string) => AcpSessionState | null;
  subscribeSession: (listener: () => void) => () => void;
  dispatchRelayEvent: (
    podKey: string,
    messageType: number,
    payload: unknown,
  ) => void;
  listWorkspaceArtifacts: (podKey: string) => Promise<AgentArtifactItem[]>;
  loadWorkspaceArtifact: (podKey: string, path: string) => Promise<Blob>;
  removePermission: (podKey: string, permissionId: string) => void;
}

export interface WebAcpSessionRuntimeInput {
  agentLabel: string;
  deps?: WebAcpRuntimeDeps;
  paneId: string;
  podKey: string;
  title: string;
}
