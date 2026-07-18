import { toBinary } from "@bufbuild/protobuf";
import {
  artifactActionPayload,
  configurationPayload,
  createAgentCommandEnvelope,
  interruptPayload,
  permissionPayload,
  sendPromptPayload,
  type AgentArtifactActionCommand,
  type AgentConnectionStatus,
  type AgentPermissionResolution,
  type AgentSessionRuntime,
  type AgentSessionSnapshot,
} from "@do-worker/agent-ui";
import { CommandEnvelopeSchema } from "@proto/agent_workbench/v2/command_pb";
import { defaultWebAgentWorkbenchRuntimeDeps } from "./webAgentWorkbenchRuntimeDefaults";
import { createWebAgentWorkbenchAttachmentUploader } from "./webAgentWorkbenchAttachmentRuntime";
import { WebAgentWorkbenchConnection } from "./WebAgentWorkbenchConnection";
import {
  decodeWebAgentWorkbenchSnapshot,
  projectWebAgentWorkbenchSnapshot,
} from "./webAgentWorkbenchProjection";
import type {
  WebAgentWorkbenchRuntimeDeps,
  WebAgentWorkbenchRuntimeInput,
} from "./webAgentWorkbenchRuntimeTypes";

export class WebAgentWorkbenchRuntime implements AgentSessionRuntime {
  readonly sessionId: string;
  readonly loadArtifact?: AgentSessionRuntime["loadArtifact"];
  readonly uploadAttachment?: AgentSessionRuntime["uploadAttachment"];
  private readonly listeners = new Set<() => void>();
  private readonly deps: WebAgentWorkbenchRuntimeDeps;
  private connection: AgentConnectionStatus = "connecting";
  private error: string | null = null;
  private snapshot: AgentSessionSnapshot;
  private readonly connectionRuntime: WebAgentWorkbenchConnection;
  constructor(private readonly input: WebAgentWorkbenchRuntimeInput) {
    if (!input.sessionId) throw new Error("agent_workbench_session_missing");
    this.sessionId = input.sessionId;
    this.deps = input.deps ?? defaultWebAgentWorkbenchRuntimeDeps;
    this.snapshot = this.project();
    this.connectionRuntime = new WebAgentWorkbenchConnection(
      this.deps,
      this.sessionId,
      input.live !== false,
      ({ connection, error }) => {
        this.connection = connection;
        this.error = error;
        this.refresh();
      },
    );
    const loadArtifact = input.loadArtifact ?? this.deps.loadArtifact;
    if (loadArtifact) {
      this.loadArtifact = (sessionId, artifactId, representationId) => {
        this.assertSession(sessionId);
        return loadArtifact({
          artifactId,
          representationId,
          sessionId,
        });
      };
    }
    this.uploadAttachment = createWebAgentWorkbenchAttachmentUploader(
      this.deps,
      (sessionId) => this.assertSession(sessionId),
    );
  }

  async open(sessionId: string): Promise<void> {
    this.assertSession(sessionId);
    await this.connectionRuntime.open();
  }

  close(sessionId: string): void {
    this.assertSession(sessionId);
    this.connectionRuntime.close();
  }

  getSnapshot(sessionId: string): AgentSessionSnapshot {
    this.assertSession(sessionId);
    return this.snapshot;
  }

  subscribe(sessionId: string, listener: () => void): () => void {
    this.assertSession(sessionId);
    this.listeners.add(listener);
    return () => this.listeners.delete(listener);
  }
  sendMessage(
    sessionId: string,
    commandId: string,
    input: Parameters<AgentSessionRuntime["sendMessage"]>[2],
  ): Promise<void> {
    return this.execute(sessionId, commandId, sendPromptPayload(input.text, input.attachments));
  }
  sendSlashCommand(
    sessionId: string,
    commandId: string,
    input: { name: string; arguments: string },
  ): Promise<void> {
    const text = `/${input.name}${input.arguments ? ` ${input.arguments}` : ""}`;
    return this.execute(sessionId, commandId, sendPromptPayload(text));
  }
  interrupt(sessionId: string, commandId: string): Promise<void> {
    const turnId = this.rawSnapshot()?.activeTurnId;
    return this.execute(sessionId, commandId, interruptPayload(turnId));
  }

  resolvePermission(
    sessionId: string,
    commandId: string,
    permissionId: string,
    result: AgentPermissionResolution,
  ): Promise<void> {
    return this.execute(
      sessionId,
      commandId,
      permissionPayload(permissionId, result),
    );
  }

  updateConfiguration(sessionId: string, commandId: string, patch: Record<string, unknown>): Promise<void> {
    return this.execute(
      sessionId,
      commandId,
      configurationPayload(patch),
    );
  }

  executeArtifactAction(
    sessionId: string,
    command: AgentArtifactActionCommand,
  ): Promise<void> {
    return this.execute(
      sessionId,
      command.commandId,
      artifactActionPayload(command),
    );
  }

  loadOlder(sessionId: string): Promise<void> {
    this.assertSession(sessionId);
    return Promise.reject(new Error("agent_workbench_history_not_implemented"));
  }

  private async execute(
    sessionId: string,
    commandId: string,
    command: Parameters<typeof createAgentCommandEnvelope>[0]["command"],
  ): Promise<void> {
    this.assertSession(sessionId);
    if (this.deps.state.projectionStatus(sessionId) !== "ready") {
      throw new Error("agent_workbench_session_not_ready");
    }
    const snapshot = this.rawSnapshot();
    if (!snapshot) throw new Error("agent_workbench_snapshot_missing");
    const envelope = await createAgentCommandEnvelope({
      command,
      commandId,
      expectedRevision: snapshot.revision,
      issuedAt: new Date().toISOString(),
      sessionId,
      streamEpoch: snapshot.streamEpoch,
    });
    const access = this.deps.getAccess();
    await this.deps.service.executeCommandConnect(
      access.orgSlug,
      access.bearerToken,
      toBinary(CommandEnvelopeSchema, envelope),
    );
  }

  private rawSnapshot() {
    return decodeWebAgentWorkbenchSnapshot(
      this.deps.state.snapshotBytes(this.sessionId),
      this.sessionId,
    );
  }

  private project(): AgentSessionSnapshot {
    return projectWebAgentWorkbenchSnapshot(
      this.rawSnapshot(),
      this.input,
      this.connection,
      this.error,
    );
  }

  private refresh(): void {
    this.snapshot = this.project();
    this.listeners.forEach((listener) => listener());
  }

  private assertSession(sessionId: string): void {
    if (sessionId !== this.sessionId) {
      throw new Error("agent_workbench_runtime_session_mismatch");
    }
  }
}
