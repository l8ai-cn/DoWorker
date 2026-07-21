import type {
  AgentAttachmentReference,
  AgentArtifactLoadRequest,
  AgentSessionRuntime,
} from "@agent-cloud/agent-ui";

export interface WebAgentWorkbenchStream {
  close(): void;
  status(): string;
  terminalError(): string | undefined;
}

export interface WebAgentWorkbenchService {
  executeCommandConnect(
    orgSlug: string,
    bearerToken: string,
    command: Uint8Array,
  ): Promise<Uint8Array>;
  getSessionSnapshotConnect(
    orgSlug: string,
    bearerToken: string,
    sessionId: string,
  ): Promise<Uint8Array>;
  streamSessionDeltasConnect(
    orgSlug: string,
    bearerToken: string,
    sessionId: string,
    replayLimit: number,
    onCommit: () => void,
    onError: (error: string) => void,
    onClose: (detail: unknown) => void,
  ): Promise<WebAgentWorkbenchStream>;
}

export interface WebAgentWorkbenchState {
  projectionStatus(sessionId: string): string | undefined;
  resyncReason(sessionId: string): string | undefined;
  revision(sessionId: string): bigint | undefined;
  snapshotBytes(sessionId: string): Uint8Array | undefined;
}

export interface WebAgentWorkbenchAccess {
  bearerToken: string;
  orgSlug: string;
}

export interface WebAgentWorkbenchAttachmentUploadInput {
  access: WebAgentWorkbenchAccess;
  file: File;
  sessionId: string;
}

export interface WebAgentWorkbenchRuntimeDeps {
  getAccess(): WebAgentWorkbenchAccess;
  loadArtifact?: (
    request: AgentArtifactLoadRequest,
  ) => Promise<Blob>;
  maxReconnectAttempts?: number;
  service: WebAgentWorkbenchService;
  sleep(milliseconds: number): Promise<void>;
  state: WebAgentWorkbenchState;
  uploadAttachment?: (
    input: WebAgentWorkbenchAttachmentUploadInput,
  ) => Promise<AgentAttachmentReference>;
}

export interface WebAgentWorkbenchRuntimeInput {
  agentLabel: string;
  deps?: WebAgentWorkbenchRuntimeDeps;
  interactionMode: "acp" | "pty";
  loadArtifact?: (
    request: AgentArtifactLoadRequest,
  ) => Promise<Blob>;
  live?: boolean;
  sessionId: string;
  title: string;
  workspaceArtifactError?: string | null;
}

export type WebAgentWorkbenchLoadArtifact =
  AgentSessionRuntime["loadArtifact"];
