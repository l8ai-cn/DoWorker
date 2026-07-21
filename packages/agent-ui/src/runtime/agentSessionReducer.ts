import { clone } from "@bufbuild/protobuf";

import {
  SessionSnapshotSchema,
  type SessionDeltaBatch,
  type SessionSnapshot,
} from "@agent-cloud/proto/agent_workbench/v2/session_pb";
import { applyAgentEvent } from "./agentSessionEventReducer";
import {
  AgentSessionReductionError,
  type AgentSessionState,
  batchIdentity,
} from "./agentSessionState";
import { validateSessionSnapshot } from "./sessionSnapshotValidation";

export {
  AgentSessionReductionError,
  type AgentSessionResyncReason,
  type AgentSessionState,
} from "./agentSessionState";

const MAX_APPLIED_BATCHES = 256;

export function applySessionSnapshot(snapshot: SessionSnapshot): AgentSessionState {
  validateSessionSnapshot(snapshot);
  return {
    snapshot: clone(SessionSnapshotSchema, snapshot),
    status: "ready",
    appliedBatches: new Map(),
  };
}

export function applyDeltaBatch(
  state: AgentSessionState,
  batch: SessionDeltaBatch,
): AgentSessionState {
  if (batch.sessionId !== state.snapshot.sessionId) {
    throw new AgentSessionReductionError("session_id_mismatch");
  }

  const identity = batchIdentity(batch);
  const appliedDigest = state.appliedBatches.get(identity);
  if (appliedDigest !== undefined) {
    if (appliedDigest !== batch.digest) {
      return requireResync(state, "digest_conflict");
    }
    return state;
  }
  if (state.status === "resync_required") return state;
  if (batch.streamEpoch !== state.snapshot.streamEpoch) {
    return requireResync(state, "stream_epoch_changed");
  }
  if (batch.baseRevision !== state.snapshot.revision) {
    return requireResync(state, "base_revision_mismatch");
  }
  if (batch.firstSequence !== state.snapshot.latestSequence + BigInt(1)) {
    return requireResync(state, "sequence_gap");
  }

  validateBatch(batch);
  const snapshot = clone(SessionSnapshotSchema, state.snapshot);
  for (const event of batch.events) applyAgentEvent(snapshot, event);
  snapshot.revision = batch.revision;
  snapshot.latestSequence = batch.lastSequence;
  validateSessionSnapshot(snapshot);

  const appliedBatches = new Map(state.appliedBatches);
  appliedBatches.set(identity, batch.digest);
  while (appliedBatches.size > MAX_APPLIED_BATCHES) {
    const oldest = appliedBatches.keys().next().value;
    if (oldest === undefined) break;
    appliedBatches.delete(oldest);
  }
  return { snapshot, status: "ready", appliedBatches };
}

function validateBatch(batch: SessionDeltaBatch): void {
  if (!batch.digest || batch.events.length === 0 || batch.revision !== batch.baseRevision + BigInt(1)) {
    throw new AgentSessionReductionError("delta_invalid");
  }
  if (batch.lastSequence < batch.firstSequence) {
    throw new AgentSessionReductionError("delta_sequence_range_invalid");
  }
  const expectedCount = batch.lastSequence - batch.firstSequence + BigInt(1);
  if (expectedCount !== BigInt(batch.events.length)) {
    throw new AgentSessionReductionError("delta_event_count_invalid");
  }

  batch.events.forEach((event, index) => {
    const envelope = event.envelope;
    const sequence = batch.firstSequence + BigInt(index);
    if (
      !envelope ||
      envelope.sessionId !== batch.sessionId ||
      envelope.streamEpoch !== batch.streamEpoch ||
      envelope.revision !== batch.revision ||
      envelope.sequence !== sequence ||
      !envelope.itemId
    ) {
      throw new AgentSessionReductionError("delta_event_envelope_invalid");
    }
  });
}

function requireResync(
  state: AgentSessionState,
  resyncReason: NonNullable<AgentSessionState["resyncReason"]>,
): AgentSessionState {
  return { ...state, status: "resync_required", resyncReason };
}
