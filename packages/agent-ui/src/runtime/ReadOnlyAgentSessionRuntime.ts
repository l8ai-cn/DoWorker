import type {
  AgentAttachmentReference,
  AgentPermissionResolution,
  AgentSessionRuntime,
  AgentSessionSnapshot,
} from "../contracts";

export class ReadOnlyAgentSessionRuntime implements AgentSessionRuntime {
  readonly loadArtifact?: AgentSessionRuntime["loadArtifact"];
  private sourceSnapshot: AgentSessionSnapshot | null = null;
  private projectedSnapshot: AgentSessionSnapshot | null = null;

  constructor(private readonly source: AgentSessionRuntime) {
    if (source.loadArtifact) {
      this.loadArtifact = source.loadArtifact.bind(source);
    }
  }

  open(sessionId: string): Promise<void> {
    return this.source.open(sessionId);
  }

  close(sessionId: string): void {
    this.source.close(sessionId);
  }

  getSnapshot(sessionId: string): AgentSessionSnapshot {
    const source = this.source.getSnapshot(sessionId);
    if (source !== this.sourceSnapshot || !this.projectedSnapshot) {
      this.sourceSnapshot = source;
      this.projectedSnapshot = {
        ...source,
        capabilities: {
          ...source.capabilities,
          interrupt: false,
          resolvePermission: false,
          sendMessage: false,
          terminal: false,
          updateConfiguration: false,
        },
        terminals: source.terminals.map((terminal) => ({
          ...terminal,
          writable: false,
        })),
      };
    }
    return this.projectedSnapshot;
  }

  subscribe(sessionId: string, listener: () => void): () => void {
    return this.source.subscribe(sessionId, listener);
  }

  sendMessage(
    _sessionId: string,
    _commandId: string,
    _input: { text: string; attachments?: AgentAttachmentReference[] },
  ): Promise<void> {
    return rejectedReadOnlyCommand();
  }

  interrupt(_sessionId: string, _commandId: string): Promise<void> {
    return rejectedReadOnlyCommand();
  }

  resolvePermission(
    _sessionId: string,
    _commandId: string,
    _permissionId: string,
    _result: AgentPermissionResolution,
  ): Promise<void> {
    return rejectedReadOnlyCommand();
  }

  updateConfiguration(
    _sessionId: string,
    _commandId: string,
    _patch: Record<string, unknown>,
  ): Promise<void> {
    return rejectedReadOnlyCommand();
  }

  loadOlder(sessionId: string, beforeItemId?: string): Promise<void> {
    return this.source.loadOlder(sessionId, beforeItemId);
  }
}

function rejectedReadOnlyCommand(): Promise<void> {
  return Promise.reject(new Error("agent_session_read_only"));
}
