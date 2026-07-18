import type {
  AgentAttachmentReference,
  AgentArtifactActionCommand,
  AgentPermissionResolution,
  AgentSessionRuntime,
  AgentSessionSnapshot,
} from "../contracts";
import type { AgentSessionConnection } from "./AgentSessionConnection";
import {
  artifactActionPayload,
  configurationPayload,
  interruptPayload,
  permissionPayload,
  sendPromptPayload,
} from "./agentSessionCommandPayloads";
import { createAgentCommandEnvelope } from "./createAgentCommandEnvelope";
import {
  projectConnectedSession,
  type AgentSessionProjectionContext,
} from "./agentSessionRuntimeProjection";

export interface AgentArtifactLoadRequest {
  artifactId: string;
  representationId?: string;
  sessionId: string;
}

export interface AgentSessionRuntimeV2Options
  extends AgentSessionProjectionContext {
  connection: AgentSessionConnection;
  loadArtifact?: (request: AgentArtifactLoadRequest) => Promise<Blob>;
  now?: () => string;
}

export class AgentSessionRuntimeV2 implements AgentSessionRuntime {
  readonly loadArtifact?: AgentSessionRuntime["loadArtifact"];
  private readonly connection: AgentSessionConnection;
  private readonly context: AgentSessionProjectionContext;
  private readonly listeners = new Set<() => void>();
  private readonly now: () => string;
  private snapshot: AgentSessionSnapshot;

  constructor(options: AgentSessionRuntimeV2Options) {
    this.connection = options.connection;
    this.context = {
      agentLabel: options.agentLabel,
      interactionMode: options.interactionMode,
      sessionId: options.sessionId,
      title: options.title,
    };
    this.now = options.now ?? (() => new Date().toISOString());
    if (options.loadArtifact) {
      this.loadArtifact = (sessionId, artifactId, representationId) => {
        this.requireSession(sessionId);
        return options.loadArtifact!({
          artifactId,
          representationId,
          sessionId,
        });
      };
    }
    this.snapshot = projectConnectedSession(this.connection, this.context);
    this.connection.subscribe(() => this.refresh());
    this.connection.getStore().subscribe(() => this.refresh());
  }

  async open(sessionId: string): Promise<void> {
    this.requireSession(sessionId);
    await this.connection.open();
    this.refresh();
  }

  close(sessionId: string): void {
    this.requireSession(sessionId);
    this.connection.close();
  }

  getSnapshot(sessionId: string): AgentSessionSnapshot {
    this.requireSession(sessionId);
    return this.snapshot;
  }

  subscribe(sessionId: string, listener: () => void): () => void {
    this.requireSession(sessionId);
    this.listeners.add(listener);
    return () => this.listeners.delete(listener);
  }

  async sendMessage(
    sessionId: string,
    commandId: string,
    input: { text: string; attachments?: AgentAttachmentReference[] },
  ): Promise<void> {
    await this.execute(
      sessionId,
      commandId,
      sendPromptPayload(input.text, input.attachments),
    );
  }

  async interrupt(sessionId: string, commandId: string): Promise<void> {
    const turnId = this.connection.getStore().getState()?.snapshot.activeTurnId;
    await this.execute(sessionId, commandId, interruptPayload(turnId));
  }

  async resolvePermission(
    sessionId: string,
    commandId: string,
    permissionId: string,
    result: AgentPermissionResolution,
  ): Promise<void> {
    await this.execute(
      sessionId,
      commandId,
      permissionPayload(permissionId, result),
    );
  }

  async updateConfiguration(
    sessionId: string,
    commandId: string,
    patch: Record<string, unknown>,
  ): Promise<void> {
    await this.execute(sessionId, commandId, configurationPayload(patch));
  }

  async executeArtifactAction(
    sessionId: string,
    command: AgentArtifactActionCommand,
  ): Promise<void> {
    await this.execute(
      sessionId,
      command.commandId,
      artifactActionPayload(command),
    );
  }

  async loadOlder(sessionId: string): Promise<void> {
    this.requireSession(sessionId);
  }

  private async execute(
    sessionId: string,
    commandId: string,
    command: Parameters<typeof createAgentCommandEnvelope>[0]["command"],
  ): Promise<void> {
    this.requireSession(sessionId);
    const state = this.connection.getStore().getState();
    if (!state || state.status !== "ready") {
      throw new Error("agent_workbench_session_not_ready");
    }
    const envelope = await createAgentCommandEnvelope({
      sessionId,
      streamEpoch: state.snapshot.streamEpoch,
      commandId,
      expectedRevision: state.snapshot.revision,
      issuedAt: this.now(),
      command,
    });
    await this.connection.getStore().execute(envelope);
  }

  private refresh(): void {
    this.snapshot = projectConnectedSession(this.connection, this.context);
    for (const listener of this.listeners) listener();
  }

  private requireSession(sessionId: string): void {
    if (sessionId !== this.context.sessionId) {
      throw new Error("agent_workbench_runtime_session_mismatch");
    }
  }
}
