import type {
  AgentArtifactActionCommand,
} from "./agentArtifactContracts";
import type {
  AgentPermissionResolution,
} from "./agentPermissionContracts";
import type { AgentSessionSnapshot } from "./contracts";

export interface CreateAgentSessionInput {
  agentId: string;
  title?: string;
  initialMessage?: string;
}

export interface AgentSessionReference {
  sessionId: string;
}

export interface AgentAttachmentReference {
  id: string;
  name: string;
  mediaType: string;
  bytes: number;
}

export interface AgentSessionRuntime {
  create?(input: CreateAgentSessionInput): Promise<AgentSessionReference>;
  open(sessionId: string): Promise<void>;
  close(sessionId: string): void;
  getSnapshot(sessionId: string): AgentSessionSnapshot;
  subscribe(sessionId: string, listener: () => void): () => void;
  sendMessage(
    sessionId: string,
    commandId: string,
    input: { text: string; attachments?: AgentAttachmentReference[] },
  ): Promise<void>;
  uploadAttachment?(
    sessionId: string,
    file: File,
  ): Promise<AgentAttachmentReference>;
  sendSlashCommand?(
    sessionId: string,
    commandId: string,
    input: { name: string; arguments: string },
  ): Promise<void>;
  interrupt(sessionId: string, commandId: string): Promise<void>;
  resolvePermission(
    sessionId: string,
    commandId: string,
    permissionId: string,
    result: AgentPermissionResolution,
  ): Promise<void>;
  updateConfiguration(
    sessionId: string,
    commandId: string,
    patch: Record<string, unknown>,
  ): Promise<void>;
  loadArtifact?(
    sessionId: string,
    artifactId: string,
    representationId?: string,
  ): Promise<Blob>;
  executeArtifactAction?(
    sessionId: string,
    command: AgentArtifactActionCommand,
  ): Promise<void>;
  loadOlder(sessionId: string, beforeItemId?: string): Promise<void>;
}
