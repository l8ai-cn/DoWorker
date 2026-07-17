import {
  isAgentWorkbenchCursorRejected,
  type AgentWorkbenchSessionTransport,
} from "./AgentWorkbenchConnectTransport";
import { AgentSessionStore } from "./AgentSessionStore";
import type { AgentSessionResyncReason } from "./agentSessionState";
import {
  asConnectionError,
  createInitialConnection,
  cursorIdentity,
  defaultAgentSessionRetryDelay,
  rejectInitialConnection,
  resolveInitialConnection,
  sessionCursor,
  type InitialConnection,
} from "./agentSessionConnectionLifecycle";

export type AgentSessionConnectionStatus =
  | "idle"
  | "connecting"
  | "connected"
  | "reconnecting"
  | "disconnected"
  | "failed";

export interface AgentSessionConnectionOptions {
  maxConsecutiveFailures?: number;
  retryDelay?: (attempt: number, signal: AbortSignal) => Promise<void>;
}

export class AgentSessionConnection {
  private readonly listeners = new Set<() => void>();
  private readonly maxConsecutiveFailures: number;
  private readonly retryDelay: NonNullable<
    AgentSessionConnectionOptions["retryDelay"]
  >;
  private readonly store: AgentSessionStore;
  private error: Error | null = null;
  private status: AgentSessionConnectionStatus = "idle";
  private lifecycle?: AbortController;
  private iteration?: AbortController;
  private initial?: InitialConnection;
  private resyncRequested = false;

  constructor(
    private readonly transport: AgentWorkbenchSessionTransport,
    options: AgentSessionConnectionOptions = {},
  ) {
    this.maxConsecutiveFailures = options.maxConsecutiveFailures ?? 5;
    if (this.maxConsecutiveFailures < 0) {
      throw new Error("agent_workbench_retry_limit_invalid");
    }
    this.retryDelay = options.retryDelay ?? defaultAgentSessionRetryDelay;
    this.store = new AgentSessionStore({
      execute: (command) => this.transport.execute(command),
      requestResync: (sessionId, reason) =>
        this.requestResync(sessionId, reason),
    });
  }

  open(): Promise<void> {
    if (this.lifecycle && !this.lifecycle.signal.aborted) {
      return this.initial?.promise ?? Promise.resolve();
    }
    this.lifecycle = new AbortController();
    this.initial = createInitialConnection();
    this.error = null;
    this.setStatus("connecting");
    void this.run(this.lifecycle.signal);
    return this.initial.promise;
  }

  close(): void {
    const wasActive = Boolean(this.lifecycle && !this.lifecycle.signal.aborted);
    this.lifecycle?.abort();
    this.iteration?.abort();
    this.lifecycle = undefined;
    this.iteration = undefined;
    if (wasActive) {
      rejectInitialConnection(
        this.initial,
        new Error("agent_workbench_connection_closed"),
      );
    }
    this.setStatus("disconnected");
  }

  getStore(): AgentSessionStore {
    return this.store;
  }

  getStatus(): AgentSessionConnectionStatus {
    return this.status;
  }

  getError(): Error | null {
    return this.error;
  }

  subscribe(listener: () => void): () => void {
    this.listeners.add(listener);
    return () => this.listeners.delete(listener);
  }

  private async requestResync(
    sessionId: string,
    _reason: AgentSessionResyncReason,
  ): Promise<void> {
    const current = this.store.getState()?.snapshot.sessionId;
    if (current !== sessionId) {
      throw new Error("agent_workbench_resync_session_mismatch");
    }
    this.resyncRequested = true;
    this.setStatus("reconnecting");
    this.iteration?.abort();
  }

  private async run(lifecycle: AbortSignal): Promise<void> {
    let failures = 0;
    let lastCursor = "";
    while (!lifecycle.aborted) {
      this.iteration = new AbortController();
      const signal = this.iteration.signal;
      try {
        const snapshot = await this.transport.getSnapshot(signal);
        if (lifecycle.aborted) return;
        this.store.applySnapshot(snapshot);
        const snapshotCursor = cursorIdentity(snapshot);
        if (snapshotCursor !== lastCursor) failures = 0;
        lastCursor = snapshotCursor;
        this.setStatus("connected");
        resolveInitialConnection(this.initial);

        for await (const batch of this.transport.streamDeltas(
          sessionCursor(snapshot),
          signal,
        )) {
          if (lifecycle.aborted) return;
          this.store.applyDeltaBatch(batch);
          const state = this.store.getState();
          if (state?.status === "resync_required") {
            failures += 1;
            this.setStatus("reconnecting");
            break;
          }
          if (state) {
            lastCursor = cursorIdentity(state.snapshot);
            failures = 0;
          }
        }
        if (lifecycle.aborted) return;
        if (this.store.getState()?.status === "resync_required") {
          if (failures > this.maxConsecutiveFailures) {
            return this.fail(new Error("agent_workbench_resync_stalled"));
          }
          continue;
        }
        if (this.resyncRequested) {
          this.resyncRequested = false;
          failures += 1;
          continue;
        }
        throw new Error("agent_workbench_stream_ended");
      } catch (cause) {
        if (lifecycle.aborted) return;
        if (this.resyncRequested || isAgentWorkbenchCursorRejected(cause)) {
          this.resyncRequested = false;
          failures += 1;
        } else {
          failures += 1;
          this.error = asConnectionError(cause);
        }
        if (failures > this.maxConsecutiveFailures) {
          return this.fail(asConnectionError(cause));
        }
        this.setStatus("reconnecting");
        await this.retryDelay(failures, lifecycle);
      }
    }
  }

  private fail(error: Error): void {
    this.error = error;
    this.lifecycle?.abort();
    this.iteration?.abort();
    rejectInitialConnection(this.initial, error);
    this.setStatus("failed");
  }

  private setStatus(status: AgentSessionConnectionStatus): void {
    if (this.status === status) return;
    this.status = status;
    for (const listener of this.listeners) listener();
  }
}
