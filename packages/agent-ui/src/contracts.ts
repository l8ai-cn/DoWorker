import type { AgentArtifactItem } from "./agentArtifactContracts";
import type { AgentToolActivityItem } from "./agentToolContracts";
import type {
  AgentPermissionRequest,
} from "./agentPermissionContracts";

export type {
  AgentArtifactActionCommand,
  AgentArtifactDimensions,
  AgentArtifactGrant,
  AgentArtifactItem,
  AgentArtifactManifest,
  AgentArtifactRepresentation,
  AgentImageEditManifest,
  AgentNormalizedRegion,
  AgentPresentationManifest,
  AgentPresentationSlide,
  AgentPresentationVersion,
  AgentVideoManifest,
} from "./agentArtifactContracts";
export type {
  AgentToolActivityItem,
  AgentToolResult,
  AgentToolStatus,
} from "./agentToolContracts";
export type {
  AgentApprovalPermissionRequest,
  AgentPermissionAnswerContent,
  AgentPermissionQuestion,
  AgentPermissionQuestionOption,
  AgentPermissionRequest,
  AgentPermissionResolution,
  AgentQuestionPermissionRequest,
} from "./agentPermissionContracts";
export type {
  AgentAttachmentReference,
  AgentSessionReference,
  AgentSessionRuntime,
  CreateAgentSessionInput,
} from "./agentSessionRuntime";

export type AgentSessionStatus =
  | "idle"
  | "launching"
  | "running"
  | "waiting"
  | "failed"
  | "completed";

export type AgentConnectionStatus =
  | "connecting"
  | "connected"
  | "reconnecting"
  | "disconnected";

export interface AgentWorkspaceCapabilities {
  sendMessage: boolean;
  interrupt: boolean;
  resolvePermission: boolean;
  updateConfiguration: boolean;
  terminal: boolean;
}

export interface AgentMessageItem {
  id: string;
  kind: "message";
  role: "user" | "assistant" | "system";
  text: string;
  status: "streaming" | "completed" | "failed";
}

export interface AgentAttachmentItem {
  attachmentId: string;
  filename: string;
  id: string;
  kind: "attachment";
  mimeType: string | null;
}

export interface AgentActivityItem {
  id: string;
  kind: "reasoning" | "error" | "system";
  title: string;
  detail?: string;
  status: "pending" | "running" | "completed" | "failed";
}

export type AgentTimelineItem =
  | AgentMessageItem
  | AgentAttachmentItem
  | AgentActivityItem
  | AgentToolActivityItem
  | AgentArtifactItem;

export interface AgentPlanStep {
  id: string;
  title: string;
  status: "pending" | "running" | "completed" | "failed";
}

export interface TerminalResource {
  id: string;
  label: string;
  status: AgentConnectionStatus;
  writable: boolean;
  controlMode?: "surface" | "host";
}

export interface AgentCommand {
  name: string;
  label: string;
  description?: string;
  requiresArgument?: boolean;
}

export interface AgentConfigurationOption {
  value: string;
  label: string;
  description?: string;
}

export interface AgentConfigurationControl {
  id: string;
  label: string;
  value: string;
  options: AgentConfigurationOption[];
}

export interface AgentSessionMetadata {
  id: string;
  label: string;
  value: string;
}

export interface AgentSessionSnapshot {
  sessionId: string;
  title: string;
  agentLabel: string;
  status: AgentSessionStatus;
  connection: AgentConnectionStatus;
  interactionMode: "acp" | "pty";
  capabilities: AgentWorkspaceCapabilities;
  commands?: AgentCommand[];
  configuration?: AgentConfigurationControl[];
  metadata?: AgentSessionMetadata[];
  items: AgentTimelineItem[];
  plan: AgentPlanStep[];
  permissions: AgentPermissionRequest[];
  terminals: TerminalResource[];
  hasOlderItems: boolean;
  error: string | null;
}

export interface TerminalControlLease {
  leaseId: string;
  expiresAt: number;
}

export interface TerminalRuntime {
  connect(resource: TerminalResource): Promise<void>;
  disconnect(resourceId: string): void;
  subscribeOutput(
    resourceId: string,
    listener: (bytes: Uint8Array) => void,
  ): () => void;
  subscribeStatus(
    resourceId: string,
    listener: (status: AgentConnectionStatus) => void,
  ): () => void;
  write(resourceId: string, bytes: Uint8Array): Promise<void>;
  resize(resourceId: string, columns: number, rows: number): Promise<void>;
  acquireControl(
    resourceId: string,
    clientLabel: string,
  ): Promise<TerminalControlLease>;
  renewControl(resourceId: string, leaseId: string): Promise<void>;
  releaseControl(resourceId: string, leaseId: string): Promise<void>;
}
