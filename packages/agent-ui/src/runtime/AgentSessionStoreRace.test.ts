import { create } from "@bufbuild/protobuf";
import { describe, expect, it, vi } from "vitest";

import {
  CommandEnvelopeSchema,
  CommandReceiptSchema,
  CommandReceiptState,
} from "@agent-cloud/proto/agent_workbench/v2/command_pb";
import {
  AgentEventSchema,
  SessionDeltaBatchSchema,
  SessionSnapshotSchema,
} from "@agent-cloud/proto/agent_workbench/v2/session_pb";
import { SessionStatus } from "@agent-cloud/proto/agent_workbench/v2/session_state_pb";
import { AgentSessionStore, type AgentSessionTransport } from "./AgentSessionStore";

describe("AgentSessionStore transport races", () => {
  it("rejects a delayed snapshot from an epoch that was already replaced", () => {
    const store = new AgentSessionStore(transport());
    store.applySnapshot(snapshot("epoch-a", 5n));
    store.applySnapshot(snapshot("epoch-b", 1n));

    expect(() => store.applySnapshot(snapshot("epoch-a", 6n))).toThrow(
      "snapshot_epoch_stale",
    );
    expect(store.getState()?.snapshot.streamEpoch).toBe("epoch-b");
    expect(store.getState()?.snapshot.revision).toBe(1n);
  });

  it("does not forget replaced epochs after many reconnects", () => {
    const store = new AgentSessionStore(transport());
    for (let index = 0; index < 40; index += 1) {
      store.applySnapshot(snapshot(`epoch-${index}`, 1n));
    }

    expect(() => store.applySnapshot(snapshot("epoch-0", 2n))).toThrow(
      "snapshot_epoch_stale",
    );
    expect(store.getState()?.snapshot.streamEpoch).toBe("epoch-39");
  });

  it("keeps newer SSE metadata when a same-state HTTP receipt arrives late", async () => {
    let resolveReceipt:
      | ((value: ReturnType<typeof receipt>) => void)
      | undefined;
    const execute = vi.fn().mockReturnValue(
      new Promise((resolve) => {
        resolveReceipt = resolve;
      }),
    );
    const store = new AgentSessionStore(transport(execute));
    store.applySnapshot(snapshot("epoch-a", 1n));
    const pending = store.execute(command());

    store.applyDeltaBatch(
      create(SessionDeltaBatchSchema, {
        sessionId: "session-1",
        streamEpoch: "epoch-a",
        baseRevision: 1n,
        revision: 2n,
        firstSequence: 2n,
        lastSequence: 2n,
        digest: "sha256:receipt",
        events: [receiptEvent()],
      }),
    );
    resolveReceipt?.(
      receipt("2026-07-16T00:00:01Z", 1n),
    );
    await pending;

    expect(store.getCommandReceipt("command-1")).toMatchObject({
      state: CommandReceiptState.ACCEPTED,
      updatedAt: "2026-07-16T00:00:02Z",
      resultingRevision: 2n,
    });
  });
});

function snapshot(streamEpoch: string, revision: bigint) {
  return create(SessionSnapshotSchema, {
    sessionId: "session-1",
    streamEpoch,
    revision,
    latestSequence: revision,
    status: SessionStatus.IDLE,
    digest: `sha256:${streamEpoch}:${revision}`,
  });
}

function command() {
  return create(CommandEnvelopeSchema, {
    sessionId: "session-1",
    streamEpoch: "epoch-a",
    commandId: "command-1",
    payloadDigest: "sha256:command",
    issuedAt: "2026-07-16T00:00:00Z",
    command: { case: "interrupt", value: {} },
  });
}

function receipt(updatedAt: string, resultingRevision: bigint) {
  return create(CommandReceiptSchema, {
    sessionId: "session-1",
    commandId: "command-1",
    payloadDigest: "sha256:command",
    state: CommandReceiptState.ACCEPTED,
    updatedAt,
    resultingRevision,
  });
}

function receiptEvent() {
  return create(AgentEventSchema, {
    envelope: {
      sessionId: "session-1",
      streamEpoch: "epoch-a",
      revision: 2n,
      sequence: 2n,
      itemId: "receipt-command-1",
      createdAt: "2026-07-16T00:00:02Z",
    },
    event: {
      case: "commandReceiptChanged",
      value: {
        receipt: receipt("2026-07-16T00:00:02Z", 2n),
      },
    },
  });
}

function transport(
  execute = vi.fn().mockResolvedValue(
    receipt("2026-07-16T00:00:01Z", 1n),
  ),
): AgentSessionTransport {
  return {
    execute,
    requestResync: vi.fn().mockResolvedValue(undefined),
  };
}
