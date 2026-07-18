import type { AgentPermissionResolution, AgentSessionRuntime, AgentSessionSnapshot } from "@do-worker/agent-ui";
import type { EmbedSessionClient } from "@/embed-session-api";
import {
  applyEmbeddedAcpConnectionState,
  applyStreamBlock,
  createRuntimeState,
  type EmbeddedRuntimeState,
} from "./embeddedRuntimeState";
import { EmbeddedAcpConfigurationConnection } from "./EmbeddedAcpConfigurationConnection";
import { EmbeddedHydrationCoordinator } from "./embeddedHydrationCoordinator";
import { applyEmbeddedRuntimeHydration } from "./embeddedRuntimeHydration";
import { EmbeddedSessionCommands } from "./EmbeddedSessionCommands";
import { consumeEmbeddedSessionStream } from "./embeddedSessionStream";
import { projectEmbeddedWorkspaceSnapshot } from "./embeddedWorkspaceProjection";

export class EmbeddedAgentSessionRuntime implements AgentSessionRuntime {
  private readonly client: EmbedSessionClient;
  private readonly commands: EmbeddedSessionCommands;
  private readonly configuration: EmbeddedAcpConfigurationConnection;
  private controller: AbortController | null = null;
  private readonly hydration = new EmbeddedHydrationCoordinator();
  private listeners = new Set<() => void>();
  private snapshot: AgentSessionSnapshot;
  private state: EmbeddedRuntimeState;

  constructor(client: EmbedSessionClient) {
    this.client = client;
    this.state = createRuntimeState(client);
    this.snapshot = projectEmbeddedWorkspaceSnapshot(this.state);
    this.configuration = new EmbeddedAcpConfigurationConnection(
      client,
      (update) => {
        this.state = applyEmbeddedAcpConnectionState(this.state, update);
        this.emit();
      },
      (cause) => this.setError(cause),
    );
    this.commands = new EmbeddedSessionCommands(
      client,
      this.configuration,
      () => this.state,
      (state) => {
        this.state = state;
        this.emit();
      },
      (cause) => this.setError(cause),
    );
  }

  async open(sessionId: string): Promise<void> {
    this.close(this.state.sessionId);
    const controller = new AbortController();
    this.controller = controller;
    this.hydration.reset();
    this.state = {
      ...createRuntimeState(this.client, sessionId),
      connection: "connecting",
    };
    this.emit();
    void consumeEmbeddedSessionStream(this.client, controller, {
      onBlock: (block) => {
        this.state = applyStreamBlock(this.state, block);
        if (block.type === "response_start" || block.type === "response_end") {
          this.hydration.markStatusChanged();
        }
        this.emit();
        if (block.type === "response_end") void this.hydrate(controller);
      },
      onConnection: (connection) => {
        this.state = {
          ...this.state,
          connection,
          error: connection === "connected" ? null : this.state.error,
        };
        this.emit();
      },
      onError: (cause) => this.setError(cause),
      onReconnect: () => this.hydrate(controller),
      onStatus: (status) => {
        this.hydration.markStatusChanged();
        this.state = { ...this.state, status };
        this.emit();
        if (status === "idle") void this.hydrate(controller);
      },
    });
    await this.hydrate(controller);
  }

  close(sessionId: string): void {
    if (sessionId && this.state.sessionId && sessionId !== this.state.sessionId) return;
    this.controller?.abort();
    this.controller = null;
    this.configuration.close();
    if (this.state.connection !== "disconnected") {
      this.state = { ...this.state, connection: "disconnected" };
      this.emit();
    }
  }

  getSnapshot(sessionId: string): AgentSessionSnapshot {
    if (!this.snapshot.sessionId && sessionId) {
      this.snapshot = { ...this.snapshot, sessionId };
    }
    return this.snapshot;
  }

  subscribe(_sessionId: string, listener: () => void): () => void {
    this.listeners.add(listener);
    return () => this.listeners.delete(listener);
  }

  async sendMessage(
    sessionId: string,
    commandId: string,
    input: { text: string },
  ): Promise<void> {
    return this.commands.sendMessage(sessionId, commandId, input.text);
  }

  sendSlashCommand(
    sessionId: string,
    commandId: string,
    input: { name: string; arguments: string },
  ): Promise<void> {
    return this.commands.sendSlashCommand(
      sessionId,
      commandId,
      input.name,
      input.arguments,
    );
  }

  async interrupt(sessionId: string, _commandId: string): Promise<void> {
    return this.commands.interrupt(sessionId);
  }

  async resolvePermission(
    sessionId: string,
    _commandId: string,
    permissionId: string,
    result: AgentPermissionResolution,
  ): Promise<void> {
    return this.commands.resolvePermission(sessionId, permissionId, result);
  }

  updateConfiguration(
    sessionId: string,
    _commandId: string,
    patch: Record<string, unknown>,
  ): Promise<void> {
    return this.commands.updateConfiguration(sessionId, patch);
  }

  async loadArtifact(sessionId: string, artifactId: string): Promise<Blob> {
    return this.commands.loadArtifact(sessionId, artifactId);
  }

  async loadOlder(sessionId: string, beforeItemId?: string): Promise<void> {
    return this.commands.loadOlder(sessionId, beforeItemId);
  }

  private setError(cause: unknown): void {
    this.state = {
      ...this.state,
      error: cause instanceof Error ? cause.message : String(cause),
    };
    this.emit();
  }

  private async hydrate(controller: AbortController): Promise<void> {
    try {
      const state = await this.hydration.hydrate(this.client, controller.signal);
      if (!state) return;
      this.state = applyEmbeddedRuntimeHydration(
        this.state,
        state.hydration,
        state.preserveStatus,
      );
      this.emit();
      if (state.hydration.session.interactionMode === "acp") {
        this.configuration.connect();
      }
    } catch (cause) {
      if (!controller.signal.aborted) this.setError(cause);
    }
  }

  private emit(): void {
    const snapshot = projectEmbeddedWorkspaceSnapshot(this.state);
    this.snapshot = snapshot.sessionId
      ? snapshot
      : { ...snapshot, sessionId: this.state.sessionId };
    this.listeners.forEach((listener) => listener());
  }
}
