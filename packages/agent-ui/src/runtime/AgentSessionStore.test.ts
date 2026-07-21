import { create } from "@bufbuild/protobuf";
import { describe, expect, it, vi } from "vitest";

import {
  CommandEnvelopeSchema,
  CommandReceiptState,
  CommandReceiptSchema,
} from "@agent-cloud/proto/agent_workbench/v2/command_pb";
import {
  AgentEventSchema,
  SessionDeltaBatchSchema,
  SessionSnapshotSchema,
} from "@agent-cloud/proto/agent_workbench/v2/session_pb";
import { SessionStatus } from "@agent-cloud/proto/agent_workbench/v2/session_state_pb";
import { AgentSessionStore } from "./AgentSessionStore";

describe("AgentSessionStore", () => {
  it("publishes once after a multi-event batch", () => {
    const store = new AgentSessionStore(transport());
    store.applySnapshot(snapshot());
    const listener = vi.fn();
    store.subscribe(listener);

    store.applyDeltaBatch(
      create(SessionDeltaBatchSchema, {
        sessionId: "session-1",
        streamEpoch: "epoch-1",
        baseRevision: 1n,
        revision: 2n,
        firstSequence: 2n,
        lastSequence: 3n,
        digest: "sha256:batch",
        events: [
          statusChanged(2n, SessionStatus.RUNNING),
          statusChanged(3n, SessionStatus.IDLE),
        ],
      }),
    );

    expect(listener).toHaveBeenCalledTimes(1);
    expect(store.getState()?.snapshot.status).toBe(SessionStatus.IDLE);
  });

  it("does not publish an identical duplicate batch", () => {
    const store = new AgentSessionStore(transport());
    store.applySnapshot(snapshot());
    const batch = create(SessionDeltaBatchSchema, {
      sessionId: "session-1",
      streamEpoch: "epoch-1",
      baseRevision: 1n,
      revision: 2n,
      firstSequence: 2n,
      lastSequence: 2n,
      digest: "sha256:batch",
      events: [statusChanged(2n, SessionStatus.RUNNING)],
    });
    store.applyDeltaBatch(batch);
    const listener = vi.fn();
    store.subscribe(listener);

    store.applyDeltaBatch(batch);

    expect(listener).not.toHaveBeenCalled();
  });

  it("stores the receipt returned by execute", async () => {
    const execute = vi.fn().mockResolvedValue(receipt(CommandReceiptState.RECEIVED));
    const store = new AgentSessionStore(transport({ execute }));
    store.applySnapshot(snapshot());
    const command = create(CommandEnvelopeSchema, {
      sessionId: "session-1",
      streamEpoch: "epoch-1",
      commandId: "command-1",
      payloadDigest: "sha256:command",
      issuedAt: "2026-07-16T00:00:00Z",
      command: { case: "interrupt", value: {} },
    });

    const result = await store.execute(command);

    expect(result.state).toBe(CommandReceiptState.RECEIVED);
    expect(store.getCommandReceipt("command-1")).toEqual(result);
  });

  it("merges an async receipt into the latest state", async () => {
    let resolveReceipt: ((value: ReturnType<typeof receipt>) => void) | undefined;
    const execute = vi.fn().mockReturnValue(
      new Promise((resolve) => {
        resolveReceipt = resolve;
      }),
    );
    const store = new AgentSessionStore(transport({ execute }));
    store.applySnapshot(snapshot());
    const pending = store.execute(command());

    store.applyDeltaBatch(
      create(SessionDeltaBatchSchema, {
        sessionId: "session-1",
        streamEpoch: "epoch-1",
        baseRevision: 1n,
        revision: 2n,
        firstSequence: 2n,
        lastSequence: 2n,
        digest: "sha256:batch",
        events: [statusChanged(2n, SessionStatus.RUNNING)],
      }),
    );
    resolveReceipt?.(receipt(CommandReceiptState.ACCEPTED));
    await pending;

    expect(store.getState()?.snapshot.revision).toBe(2n);
    expect(store.getState()?.snapshot.status).toBe(SessionStatus.RUNNING);
    expect(store.getCommandReceipt("command-1")?.state).toBe(
      CommandReceiptState.ACCEPTED,
    );
  });

  it("preserves resync state when an async receipt arrives", async () => {
    let resolveReceipt: ((value: ReturnType<typeof receipt>) => void) | undefined;
    const execute = vi.fn().mockReturnValue(
      new Promise((resolve) => {
        resolveReceipt = resolve;
      }),
    );
    const store = new AgentSessionStore(transport({ execute }));
    store.applySnapshot(snapshot());
    const pending = store.execute(command());

    const resync = store.requestResync("transport_reconnect");
    resolveReceipt?.(receipt(CommandReceiptState.SUCCEEDED));
    await Promise.all([pending, resync]);

    expect(store.getState()?.status).toBe("resync_required");
    expect(store.getState()?.resyncReason).toBe("transport_reconnect");
    expect(store.getCommandReceipt("command-1")?.state).toBe(
      CommandReceiptState.SUCCEEDED,
    );
  });

  it("marks commands unavailable before requesting a resync", async () => {
    const requestResync = vi.fn().mockResolvedValue(undefined);
    const store = new AgentSessionStore(transport({ requestResync }));
    store.applySnapshot(snapshot());
    const listener = vi.fn();
    store.subscribe(listener);

    await store.requestResync("manual_reconnect");

    expect(store.getState()?.status).toBe("resync_required");
    expect(store.getState()?.resyncReason).toBe("manual_reconnect");
    expect(requestResync).toHaveBeenCalledWith("session-1", "manual_reconnect");
    expect(listener).toHaveBeenCalledTimes(1);
  });

  it("does not expose the internal session state for mutation", () => {
    const store = new AgentSessionStore(transport());
    store.applySnapshot(snapshot());

    const exposed = store.getState();
    if (!exposed) throw new Error("session_state_missing");
    exposed.snapshot.revision = 99n;
    (exposed.appliedBatches as Map<string, string>).set(
      "mutated",
      "sha256:mutated",
    );

    expect(store.getState()?.snapshot.revision).toBe(1n);
    expect(store.getState()?.appliedBatches.size).toBe(0);

    store.applyDeltaBatch(
      create(SessionDeltaBatchSchema, {
        sessionId: "session-1",
        streamEpoch: "epoch-1",
        baseRevision: 1n,
        revision: 2n,
        firstSequence: 2n,
        lastSequence: 2n,
        digest: "sha256:batch",
        events: [statusChanged(2n, SessionStatus.RUNNING)],
      }),
    );

    expect(store.getState()?.snapshot.revision).toBe(2n);
  });

  it("returns a detached command receipt", async () => {
    const execute = vi.fn().mockResolvedValue(receipt(CommandReceiptState.RUNNING));
    const store = new AgentSessionStore(transport({ execute }));
    store.applySnapshot(snapshot());
    await store.execute(command());

    const exposed = store.getCommandReceipt("command-1");
    if (!exposed) throw new Error("command_receipt_missing");
    exposed.state = CommandReceiptState.FAILED;

    expect(store.getCommandReceipt("command-1")?.state).toBe(
      CommandReceiptState.RUNNING,
    );
  });

  it("rejects a stale snapshot without rolling state back", () => {
    const store = new AgentSessionStore(transport());
    store.applySnapshot(snapshot({ revision: 2n, latestSequence: 2n }));

    expect(() => store.applySnapshot(snapshot())).toThrow("snapshot_stale");
    expect(store.getState()?.snapshot.revision).toBe(2n);
    expect(store.getState()?.snapshot.latestSequence).toBe(2n);
  });

  it("does not publish an identical snapshot twice", () => {
    const store = new AgentSessionStore(transport());
    store.applySnapshot(snapshot());
    const listener = vi.fn();
    store.subscribe(listener);

    store.applySnapshot(snapshot());

    expect(listener).not.toHaveBeenCalled();
  });

  it("preserves duplicate detection after a same-cursor metadata refresh", () => {
    const store = new AgentSessionStore(transport());
    store.applySnapshot(snapshot());
    const batch = create(SessionDeltaBatchSchema, {
      sessionId: "session-1",
      streamEpoch: "epoch-1",
      baseRevision: 1n,
      revision: 2n,
      firstSequence: 2n,
      lastSequence: 2n,
      digest: "sha256:batch",
      events: [statusChanged(2n, SessionStatus.RUNNING)],
    });
    store.applyDeltaBatch(batch);
    const authoritative = store.getState()!.snapshot;
    authoritative.digest = "sha256:authoritative";

    store.applySnapshot(authoritative);
    store.applyDeltaBatch(batch);

    expect(store.getState()?.snapshot.digest).toBe("sha256:authoritative");
    expect(store.getState()?.appliedBatches.size).toBe(1);
    expect(store.getState()?.status).toBe("ready");
  });

  it("disables commands after a duplicate range digest conflict", async () => {
    const store = new AgentSessionStore(transport());
    store.applySnapshot(snapshot());
    const original = create(SessionDeltaBatchSchema, {
      sessionId: "session-1",
      streamEpoch: "epoch-1",
      baseRevision: 1n,
      revision: 2n,
      firstSequence: 2n,
      lastSequence: 2n,
      digest: "sha256:original",
      events: [statusChanged(2n, SessionStatus.RUNNING)],
    });
    store.applyDeltaBatch(original);

    store.applyDeltaBatch(create(SessionDeltaBatchSchema, {
      ...original,
      digest: "sha256:conflict",
    }));

    expect(store.getState()?.status).toBe("resync_required");
    expect(store.getState()?.resyncReason).toBe("digest_conflict");
    await expect(store.execute(command())).rejects.toThrow(
      "session_resync_required",
    );
  });

  it("rejects different content at the same snapshot cursor", () => {
    const store = new AgentSessionStore(transport());
    store.applySnapshot(snapshot());

    expect(() =>
      store.applySnapshot(snapshot({ status: SessionStatus.RUNNING })),
    ).toThrow("snapshot_cursor_conflict");
    expect(store.getState()?.snapshot.status).toBe(SessionStatus.IDLE);
  });

  it("rejects a receipt that does not match the executed command", async () => {
    const execute = vi.fn().mockResolvedValue(
      create(CommandReceiptSchema, {
        sessionId: "session-1",
        commandId: "different-command",
        payloadDigest: "sha256:different",
        state: CommandReceiptState.RECEIVED,
      }),
    );
    const store = new AgentSessionStore(transport({ execute }));
    store.applySnapshot(snapshot());

    await expect(store.execute(command())).rejects.toThrow(
      "command_receipt_mismatch",
    );
    expect(store.getState()?.snapshot.commandReceipts).toHaveLength(0);
  });

  it("ignores a delayed HTTP receipt that precedes the SSE state", async () => {
    let resolveReceipt: ((value: ReturnType<typeof receipt>) => void) | undefined;
    const execute = vi.fn().mockReturnValue(
      new Promise((resolve) => {
        resolveReceipt = resolve;
      }),
    );
    const store = new AgentSessionStore(transport({ execute }));
    store.applySnapshot(snapshot());
    const pending = store.execute(command());

    store.applyDeltaBatch(
      create(SessionDeltaBatchSchema, {
        sessionId: "session-1",
        streamEpoch: "epoch-1",
        baseRevision: 1n,
        revision: 2n,
        firstSequence: 2n,
        lastSequence: 2n,
        digest: "sha256:accepted",
        events: [receiptChanged(2n, CommandReceiptState.ACCEPTED)],
      }),
    );
    resolveReceipt?.(receipt(CommandReceiptState.RECEIVED));
    await expect(pending).resolves.toMatchObject({
      state: CommandReceiptState.RECEIVED,
    });

    expect(store.getCommandReceipt("command-1")?.state).toBe(
      CommandReceiptState.ACCEPTED,
    );
  });

  it("isolates listener failures from protocol state updates", () => {
    const store = new AgentSessionStore(transport());
    store.applySnapshot(snapshot());
    const second = vi.fn();
    const consoleError = vi.spyOn(console, "error").mockImplementation(() => {});
    store.subscribe(() => {
      throw new Error("listener_failed");
    });
    store.subscribe(second);

    store.applyDeltaBatch(
      create(SessionDeltaBatchSchema, {
        sessionId: "session-1",
        streamEpoch: "epoch-1",
        baseRevision: 1n,
        revision: 2n,
        firstSequence: 2n,
        lastSequence: 2n,
        digest: "sha256:batch",
        events: [statusChanged(2n, SessionStatus.RUNNING)],
      }),
    );

    expect(second).toHaveBeenCalledTimes(1);
    expect(consoleError).toHaveBeenCalledWith(
      "agent_session_listener_failed",
      expect.any(Error),
    );
    expect(store.getState()?.snapshot.revision).toBe(2n);
    consoleError.mockRestore();
  });
});

function snapshot(
  patch: {
    revision?: bigint;
    latestSequence?: bigint;
    status?: SessionStatus;
    digest?: string;
  } = {},
) {
  return create(SessionSnapshotSchema, {
    sessionId: "session-1",
    streamEpoch: "epoch-1",
    revision: patch.revision ?? 1n,
    latestSequence: patch.latestSequence ?? 1n,
    status: patch.status ?? SessionStatus.IDLE,
    digest: patch.digest,
  });
}

function statusChanged(sequence: bigint, status: SessionStatus) {
  return create(AgentEventSchema, {
    envelope: {
      sessionId: "session-1",
      streamEpoch: "epoch-1",
      revision: 2n,
      sequence,
      itemId: `status-${sequence}`,
      createdAt: "2026-07-16T00:00:00Z",
    },
    event: {
      case: "sessionStatusChanged",
      value: { status },
    },
  });
}

function receipt(state: CommandReceiptState) {
  return create(CommandReceiptSchema, {
    sessionId: "session-1",
    commandId: "command-1",
    payloadDigest: "sha256:command",
    state,
  });
}

function receiptChanged(sequence: bigint, state: CommandReceiptState) {
  return create(AgentEventSchema, {
    envelope: {
      sessionId: "session-1",
      streamEpoch: "epoch-1",
      revision: 2n,
      sequence,
      itemId: `receipt-${sequence}`,
      createdAt: "2026-07-16T00:00:00Z",
    },
    event: {
      case: "commandReceiptChanged",
      value: { receipt: receipt(state) },
    },
  });
}

function command() {
  return create(CommandEnvelopeSchema, {
    sessionId: "session-1",
    streamEpoch: "epoch-1",
    commandId: "command-1",
    payloadDigest: "sha256:command",
    issuedAt: "2026-07-16T00:00:00Z",
    command: { case: "interrupt", value: {} },
  });
}

function transport(overrides: Record<string, unknown> = {}) {
  return {
    execute: vi.fn(),
    requestResync: vi.fn(),
    ...overrides,
  };
}
