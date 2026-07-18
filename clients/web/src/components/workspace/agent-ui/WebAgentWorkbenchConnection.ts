import type { AgentConnectionStatus } from "@do-worker/agent-ui";

import type {
  WebAgentWorkbenchRuntimeDeps,
  WebAgentWorkbenchStream,
} from "./webAgentWorkbenchRuntimeTypes";

const RECONNECT_DELAYS = [0, 250, 1_000, 2_000, 4_000];

export interface WebAgentWorkbenchConnectionState {
  connection: AgentConnectionStatus;
  error: string | null;
}

export class WebAgentWorkbenchConnection {
  private stream: WebAgentWorkbenchStream | null = null;
  private opened = false;
  private generation = 0;
  private streamCycle = 0;
  private reconnectAttempts = 0;
  private reconnectTask: {
    generation: number;
    promise: Promise<void>;
  } | null = null;
  private lastProgressRevision = BigInt(0);

  constructor(
    private readonly deps: WebAgentWorkbenchRuntimeDeps,
    private readonly sessionId: string,
    private readonly onState: (
      state: WebAgentWorkbenchConnectionState,
    ) => void,
  ) {}

  async open(): Promise<void> {
    if (this.opened) return;
    this.opened = true;
    const generation = ++this.generation;
    this.update("connecting", null);
    try {
      await this.connect(generation);
    } catch (cause) {
      this.handleFailure(cause, generation);
    }
  }

  close(): void {
    this.opened = false;
    this.generation++;
    this.streamCycle++;
    this.stream?.close();
    this.stream = null;
    this.update("disconnected", null);
  }

  private async connect(generation: number): Promise<void> {
    const access = this.deps.getAccess();
    await this.deps.service.getSessionSnapshotConnect(
      access.orgSlug,
      access.bearerToken,
      this.sessionId,
    );
    if (!this.isCurrent(generation)) return;
    this.recordProgress();
    this.update("connected", null);
    const cycle = ++this.streamCycle;
    const stream = await this.deps.service.streamSessionDeltasConnect(
      access.orgSlug,
      access.bearerToken,
      this.sessionId,
      500,
      () => this.handleCommit(generation, cycle),
      (error) => this.handleStreamEnd(error, generation, cycle),
      (detail) =>
        this.handleStreamEnd(closeError(detail), generation, cycle),
    );
    if (!this.isStreamCurrent(generation, cycle)) {
      stream.close();
      return;
    }
    this.stream = stream;
  }

  private handleCommit(generation: number, cycle: number): void {
    if (!this.isStreamCurrent(generation, cycle)) return;
    if (this.deps.state.projectionStatus(this.sessionId) === "resync_required") {
      this.handleStreamEnd(
        this.deps.state.resyncReason(this.sessionId) ?? "resync_required",
        generation,
        cycle,
      );
      return;
    }
    this.recordProgress();
    this.update("connected", null);
  }

  private handleStreamEnd(
    reason: string,
    generation: number,
    cycle: number,
  ): void {
    if (!this.isStreamCurrent(generation, cycle)) return;
    this.streamCycle++;
    this.stream?.close();
    this.stream = null;
    this.handleFailure(reason, generation);
  }

  private handleFailure(cause: unknown, generation: number): void {
    if (!this.isCurrent(generation)) return;
    const error = cause instanceof Error ? cause.message : String(cause);
    this.update("reconnecting", error);
    this.startReconnectLoop(generation, error);
  }

  private startReconnectLoop(generation: number, initialError: string): void {
    if (this.reconnectTask?.generation === generation) return;
    const promise = this.reconnect(generation, initialError).finally(() => {
      if (this.reconnectTask?.promise === promise) this.reconnectTask = null;
    });
    this.reconnectTask = { generation, promise };
  }

  private async reconnect(
    generation: number,
    initialError: string,
  ): Promise<void> {
    let error = initialError;
    const maximum =
      this.deps.maxReconnectAttempts ?? RECONNECT_DELAYS.length;
    while (this.isCurrent(generation) && this.reconnectAttempts < maximum) {
      const attempt = this.reconnectAttempts++;
      await this.deps.sleep(RECONNECT_DELAYS[attempt] ?? 4_000);
      if (!this.isCurrent(generation)) return;
      try {
        await this.connect(generation);
        return;
      } catch (cause) {
        error = cause instanceof Error ? cause.message : String(cause);
        this.update("reconnecting", error);
      }
    }
    if (this.isCurrent(generation)) this.update("disconnected", error);
  }

  private recordProgress(): void {
    const revision = this.deps.state.revision(this.sessionId) ?? BigInt(0);
    if (revision <= this.lastProgressRevision) return;
    this.lastProgressRevision = revision;
    this.reconnectAttempts = 0;
  }

  private update(
    connection: AgentConnectionStatus,
    error: string | null,
  ): void {
    this.onState({ connection, error });
  }

  private isCurrent(generation: number): boolean {
    return this.opened && generation === this.generation;
  }

  private isStreamCurrent(generation: number, cycle: number): boolean {
    return this.isCurrent(generation) && cycle === this.streamCycle;
  }
}

function closeError(detail: unknown): string {
  if (!detail || typeof detail !== "object") {
    return "agent_workbench_stream_closed";
  }
  const value = detail as { error?: unknown; status?: unknown };
  return typeof value.error === "string" && value.error
    ? value.error
    : String(value.status || "agent_workbench_stream_closed");
}
