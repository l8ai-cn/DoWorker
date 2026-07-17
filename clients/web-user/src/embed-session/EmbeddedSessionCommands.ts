import type { AgentPermissionResolution } from "@do-worker/agent-ui";

import type { EmbedSessionClient } from "@/embed-session-api";
import { EmbeddedAcpConfigurationConnection } from "./EmbeddedAcpConfigurationConnection";
import { loadEmbeddedOlderItems } from "./embeddedHistoryLoader";
import { sendEmbeddedMessage } from "./embeddedMessageState";
import { resolveEmbeddedPermission } from "./embeddedPermissionResolution";
import type { EmbeddedRuntimeState } from "./embeddedRuntimeState";

export class EmbeddedSessionCommands {
  private readonly client: EmbedSessionClient;
  private readonly commit: (state: EmbeddedRuntimeState) => void;
  private readonly configuration: EmbeddedAcpConfigurationConnection;
  private readonly getState: () => EmbeddedRuntimeState;
  private readonly onError: (cause: unknown) => void;

  constructor(
    client: EmbedSessionClient,
    configuration: EmbeddedAcpConfigurationConnection,
    getState: () => EmbeddedRuntimeState,
    commit: (state: EmbeddedRuntimeState) => void,
    onError: (cause: unknown) => void,
  ) {
    this.client = client;
    this.configuration = configuration;
    this.getState = getState;
    this.commit = commit;
    this.onError = onError;
  }

  async sendMessage(
    sessionId: string,
    commandId: string,
    text: string,
  ): Promise<void> {
    const state = this.assertSession(sessionId);
    try {
      await sendEmbeddedMessage(
        this.client,
        state,
        commandId,
        text,
        (transform) => this.commit(transform(this.getState())),
      );
    } catch (cause) {
      this.onError(cause);
      throw cause;
    }
  }

  sendSlashCommand(
    sessionId: string,
    commandId: string,
    name: string,
    args: string,
  ): Promise<void> {
    return this.sendMessage(
      sessionId,
      commandId,
      `/${name}${args ? ` ${args}` : ""}`,
    );
  }

  async interrupt(sessionId: string): Promise<void> {
    this.assertSession(sessionId);
    if (!this.client.interrupt) throw new Error("Interrupt is not permitted");
    await this.client.interrupt();
  }

  async resolvePermission(
    sessionId: string,
    permissionId: string,
    result: AgentPermissionResolution,
  ): Promise<void> {
    const state = this.assertSession(sessionId);
    this.commit(
      await resolveEmbeddedPermission(
        this.client,
        state,
        permissionId,
        result,
      ),
    );
  }

  updateConfiguration(
    sessionId: string,
    patch: Record<string, unknown>,
  ): Promise<void> {
    this.assertSession(sessionId);
    return this.configuration.update(patch);
  }

  loadArtifact(sessionId: string, artifactId: string): Promise<Blob> {
    this.assertSession(sessionId);
    if (!this.client.loadArtifact) {
      return Promise.reject(new Error("Artifact loading is unavailable"));
    }
    return this.client.loadArtifact(artifactId);
  }

  async loadOlder(sessionId: string, beforeItemId?: string): Promise<void> {
    const state = this.assertSession(sessionId);
    this.commit(
      await loadEmbeddedOlderItems(
        this.client,
        state,
        beforeItemId,
      ),
    );
  }

  private assertSession(sessionId: string): EmbeddedRuntimeState {
    const state = this.getState();
    if (sessionId !== state.sessionId) {
      throw new Error("Embedded session is not open");
    }
    return state;
  }
}
