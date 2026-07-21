import type { SessionSnapshot } from "@agent-cloud/proto/agent_workbench/v2/session_pb";

export type AgentSessionResyncReason =
  | "base_revision_mismatch"
  | "buffer_overflow"
  | "digest_conflict"
  | "manual_reconnect"
  | "sequence_gap"
  | "stream_epoch_changed"
  | "transport_reconnect";

export interface AgentSessionState {
  snapshot: SessionSnapshot;
  status: "ready" | "resync_required";
  resyncReason?: AgentSessionResyncReason;
  appliedBatches: ReadonlyMap<string, string>;
}

export class AgentSessionReductionError extends Error {
  readonly code: string;

  constructor(code: string) {
    super(code);
    this.name = "AgentSessionReductionError";
    this.code = code;
  }
}

export function batchIdentity(input: {
  streamEpoch: string;
  revision: bigint;
  firstSequence: bigint;
  lastSequence: bigint;
}): string {
  return [
    input.streamEpoch,
    input.revision.toString(),
    input.firstSequence.toString(),
    input.lastSequence.toString(),
  ].join(":");
}
