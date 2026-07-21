import { clone, equals } from "@bufbuild/protobuf";

import {
  AgentErrorSchema,
  CommandReceiptSchema,
  CommandReceiptState,
  type CommandReceipt,
} from "@agent-cloud/proto/agent_workbench/v2/command_pb";
import { AgentSessionReductionError } from "./agentSessionState";

const allowedTransitions = new Map<CommandReceiptState, ReadonlySet<CommandReceiptState>>([
  [
    CommandReceiptState.RECEIVED,
    new Set([CommandReceiptState.ACCEPTED, CommandReceiptState.REJECTED]),
  ],
  [
    CommandReceiptState.ACCEPTED,
    new Set([
      CommandReceiptState.RUNNING,
      CommandReceiptState.SUCCEEDED,
      CommandReceiptState.FAILED,
      CommandReceiptState.CANCELLED,
    ]),
  ],
  [
    CommandReceiptState.RUNNING,
    new Set([
      CommandReceiptState.SUCCEEDED,
      CommandReceiptState.FAILED,
      CommandReceiptState.CANCELLED,
    ]),
  ],
]);

export function upsertCommandReceipt(
  receipts: readonly CommandReceipt[],
  nextReceipt: CommandReceipt,
  sessionId: string,
): CommandReceipt[] {
  if (nextReceipt.sessionId !== sessionId || !nextReceipt.commandId || !nextReceipt.payloadDigest) {
    throw new AgentSessionReductionError("receipt_invalid");
  }

  const index = receipts.findIndex((receipt) => receipt.commandId === nextReceipt.commandId);
  if (index < 0) {
    if (nextReceipt.state === CommandReceiptState.UNSPECIFIED) {
      throw new AgentSessionReductionError("receipt_initial_state");
    }
    return [...receipts, clone(CommandReceiptSchema, nextReceipt)];
  }

  const current = receipts[index]!;
  if (current.payloadDigest !== nextReceipt.payloadDigest) {
    throw new AgentSessionReductionError("command_id_conflict");
  }
  if (current.state === nextReceipt.state) {
    if (isTerminal(current.state)) {
      if (!equals(CommandReceiptSchema, current, nextReceipt)) {
        throw new AgentSessionReductionError("receipt_terminal");
      }
      return receipts.slice();
    }
    const updated = receipts.slice();
    updated[index] = mergeSameStateReceipt(current, nextReceipt);
    return updated;
  }
  if (!allowedTransitions.get(current.state)?.has(nextReceipt.state)) {
    if (isTerminal(current.state)) {
      throw new AgentSessionReductionError("receipt_terminal");
    }
    throw new AgentSessionReductionError("receipt_transition_invalid");
  }

  const updated = receipts.slice();
  updated[index] = clone(CommandReceiptSchema, nextReceipt);
  return updated;
}

export function mergeTransportCommandReceipt(
  receipts: readonly CommandReceipt[],
  nextReceipt: CommandReceipt,
  sessionId: string,
): CommandReceipt[] | undefined {
  const current = receipts.find(
    (receipt) => receipt.commandId === nextReceipt.commandId,
  );
  if (
    current &&
    current.state === nextReceipt.state &&
    current.payloadDigest === nextReceipt.payloadDigest &&
    isReceiptOlder(nextReceipt, current)
  ) {
    return undefined;
  }
  if (
    current &&
    current.state !== nextReceipt.state &&
    current.payloadDigest === nextReceipt.payloadDigest &&
    canReach(nextReceipt.state, current.state)
  ) {
    return undefined;
  }
  return upsertCommandReceipt(receipts, nextReceipt, sessionId);
}

function mergeSameStateReceipt(
  current: CommandReceipt,
  next: CommandReceipt,
): CommandReceipt {
  if (isReceiptOlder(next, current)) {
    return clone(CommandReceiptSchema, current);
  }
  const merged = clone(CommandReceiptSchema, next);
  merged.resultingRevision = maximumRevision(
    current.resultingRevision,
    next.resultingRevision,
  );
  merged.error ??= current.error
    ? clone(AgentErrorSchema, current.error)
    : undefined;
  merged.receivedAt ||= current.receivedAt;
  merged.updatedAt = laterTimestamp(current.updatedAt, next.updatedAt);
  return merged;
}

function isReceiptOlder(
  candidate: CommandReceipt,
  current: CommandReceipt,
): boolean {
  if (
    candidate.resultingRevision !== undefined &&
    current.resultingRevision !== undefined
  ) {
    if (candidate.resultingRevision !== current.resultingRevision) {
      return candidate.resultingRevision < current.resultingRevision;
    }
  }
  const candidateTime = timestampMillis(candidate.updatedAt);
  const currentTime = timestampMillis(current.updatedAt);
  return candidateTime !== undefined &&
    currentTime !== undefined &&
    candidateTime < currentTime;
}

function maximumRevision(
  left: bigint | undefined,
  right: bigint | undefined,
): bigint | undefined {
  if (left === undefined) return right;
  if (right === undefined) return left;
  return left > right ? left : right;
}

function laterTimestamp(
  left: string | undefined,
  right: string | undefined,
): string | undefined {
  const leftTime = timestampMillis(left);
  const rightTime = timestampMillis(right);
  if (leftTime === undefined) return right ?? left;
  if (rightTime === undefined) return left;
  return rightTime >= leftTime ? right : left;
}

function timestampMillis(value: string | undefined): number | undefined {
  if (!value) return undefined;
  const parsed = Date.parse(value);
  return Number.isFinite(parsed) ? parsed : undefined;
}

function canReach(
  from: CommandReceiptState,
  target: CommandReceiptState,
  visited = new Set<CommandReceiptState>(),
): boolean {
  if (from === target) return true;
  if (visited.has(from)) return false;
  visited.add(from);
  for (const next of allowedTransitions.get(from) ?? []) {
    if (canReach(next, target, visited)) return true;
  }
  return false;
}

function isTerminal(state: CommandReceiptState): boolean {
  return (
    state === CommandReceiptState.SUCCEEDED ||
    state === CommandReceiptState.FAILED ||
    state === CommandReceiptState.REJECTED ||
    state === CommandReceiptState.CANCELLED
  );
}
