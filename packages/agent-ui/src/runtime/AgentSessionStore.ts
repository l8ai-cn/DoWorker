import { clone } from "@bufbuild/protobuf";

import {
  CommandReceiptSchema,
  type CommandEnvelope,
  type CommandReceipt,
} from "@agent-cloud/proto/agent_workbench/v2/command_pb";
import type {
  SessionDeltaBatch,
  SessionSnapshot,
} from "@agent-cloud/proto/agent_workbench/v2/session_pb";
import { SessionSnapshotSchema } from "@agent-cloud/proto/agent_workbench/v2/session_pb";
import {
  applyDeltaBatch,
  applySessionSnapshot,
  AgentSessionReductionError,
  type AgentSessionResyncReason,
  type AgentSessionState,
} from "./agentSessionReducer";
import { mergeTransportCommandReceipt } from "./commandReceiptTransitions";
import { validateSnapshotAdvance } from "./sessionSnapshotAdvance";

export interface AgentSessionTransport {
  execute(command: CommandEnvelope): Promise<CommandReceipt>;
  requestResync(sessionId: string, reason: AgentSessionResyncReason): Promise<void>;
}

export class AgentSessionStore {
  private state: AgentSessionState | null = null;
  private readonly listeners = new Set<() => void>();
  private readonly seenEpochs = new Set<string>();

  constructor(private readonly transport: AgentSessionTransport) {}

  getState(): AgentSessionState | null {
    return this.state ? cloneSessionState(this.state) : null;
  }

  subscribe(listener: () => void): () => void {
    this.listeners.add(listener);
    return () => this.listeners.delete(listener);
  }

  applySnapshot(snapshot: SessionSnapshot): void {
    const next = applySessionSnapshot(snapshot);
    if (this.state) {
      if (
        next.snapshot.streamEpoch !== this.state.snapshot.streamEpoch &&
        this.seenEpochs.has(next.snapshot.streamEpoch)
      ) {
        throw new AgentSessionReductionError("snapshot_epoch_stale");
      }
      const advance = validateSnapshotAdvance(
        this.state.snapshot,
        next.snapshot,
      );
      if (advance === "identical" && this.state.status === "ready") return;
      if (advance === "metadata" && this.state.status === "ready") {
        next.appliedBatches = this.state.appliedBatches;
      }
    }
    this.rememberEpoch(next.snapshot.streamEpoch);
    this.commit(next);
  }

  applyDeltaBatch(batch: SessionDeltaBatch): void {
    const state = this.requireState();
    const next = applyDeltaBatch(state, batch);
    if (next === state) return;
    this.commit(next);
  }

  async execute(command: CommandEnvelope): Promise<CommandReceipt> {
    const state = this.requireReadyState();
    if (
      command.sessionId !== state.snapshot.sessionId ||
      command.streamEpoch !== state.snapshot.streamEpoch
    ) {
      throw new Error("command_session_mismatch");
    }

    const receipt = await this.transport.execute(command);
    if (
      receipt.sessionId !== command.sessionId ||
      receipt.commandId !== command.commandId ||
      receipt.payloadDigest !== command.payloadDigest
    ) {
      throw new AgentSessionReductionError("command_receipt_mismatch");
    }
    const current = this.requireState();
    const snapshot = clone(SessionSnapshotSchema, current.snapshot);
    const commandReceipts = mergeTransportCommandReceipt(
      snapshot.commandReceipts,
      receipt,
      snapshot.sessionId,
    );
    if (commandReceipts === undefined) return receipt;
    snapshot.commandReceipts = commandReceipts;
    this.commit({ ...current, snapshot });
    return receipt;
  }

  getCommandReceipt(commandId: string): CommandReceipt | undefined {
    const receipt = this.state?.snapshot.commandReceipts.find(
      (receipt) => receipt.commandId === commandId,
    );
    return receipt ? clone(CommandReceiptSchema, receipt) : undefined;
  }

  async requestResync(reason: AgentSessionResyncReason): Promise<void> {
    const state = this.requireState();
    if (state.status !== "resync_required" || state.resyncReason !== reason) {
      this.commit({ ...state, status: "resync_required", resyncReason: reason });
    }
    await this.transport.requestResync(state.snapshot.sessionId, reason);
  }

  private requireState(): AgentSessionState {
    if (!this.state) throw new Error("session_snapshot_missing");
    return this.state;
  }

  private requireReadyState(): AgentSessionState {
    const state = this.requireState();
    if (state.status !== "ready") throw new Error("session_resync_required");
    return state;
  }

  private commit(state: AgentSessionState): void {
    this.state = state;
    this.publish();
  }

  private rememberEpoch(streamEpoch: string): void {
    this.seenEpochs.add(streamEpoch);
  }

  private publish(): void {
    for (const listener of this.listeners) {
      try {
        listener();
      } catch (error) {
        console.error("agent_session_listener_failed", error);
      }
    }
  }
}

function cloneSessionState(state: AgentSessionState): AgentSessionState {
  return {
    ...state,
    snapshot: clone(SessionSnapshotSchema, state.snapshot),
    appliedBatches: new Map(state.appliedBatches),
  };
}
