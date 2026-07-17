import { create } from "@bufbuild/protobuf";
import { describe, expect, it, vi } from "vitest";

import type {
  CommandEnvelope,
  CommandReceipt,
} from "@do-worker/proto/agent_workbench/v2/command_pb";
import {
  AgentEventSchema,
  SessionDeltaBatchSchema,
  SessionSnapshotSchema,
  type SessionCursor,
  type SessionDeltaBatch,
} from "@do-worker/proto/agent_workbench/v2/session_pb";
import { SessionStatus } from "@do-worker/proto/agent_workbench/v2/session_state_pb";
import type { AgentWorkbenchSessionTransport } from "./AgentWorkbenchConnectTransport";
import { AgentSessionConnection } from "./AgentSessionConnection";

describe("AgentSessionConnection", () => {
  it("resynchronizes from a fresh snapshot after a sequence gap", async () => {
    const snapshots = [
      snapshot({ revision: 1n, latestSequence: 1n }),
      snapshot({ revision: 2n, latestSequence: 2n }),
    ];
    const cursors: SessionCursor[] = [];
    let streamCall = 0;
    const transport = fakeTransport({
      getSnapshot: vi.fn(async () => snapshots.shift() ?? snapshots[0]!),
      streamDeltas: (cursor, signal) => {
        cursors.push(cursor);
        streamCall += 1;
        return streamCall === 1
          ? oneBatch(sequenceGapBatch())
          : waitForAbort(signal);
      },
    });
    const connection = new AgentSessionConnection(transport, {
      retryDelay: async () => {},
    });

    await connection.open();
    await vi.waitFor(() => {
      expect(connection.getStore().getState()?.snapshot.revision).toBe(2n);
    });

    expect(cursors).toEqual([
      expect.objectContaining({ revision: 1n, sequence: 1n }),
      expect.objectContaining({ revision: 2n, sequence: 2n }),
    ]);
    expect(connection.getStore().getState()?.status).toBe("ready");
    connection.close();
  });

  it("stops after repeated streams make no progress", async () => {
    const transport = fakeTransport({
      getSnapshot: vi.fn(async () => snapshot()),
      streamDeltas: () => emptyStream(),
    });
    const connection = new AgentSessionConnection(transport, {
      maxConsecutiveFailures: 2,
      retryDelay: async () => {},
    });

    await connection.open();
    await vi.waitFor(() => {
      expect(connection.getStatus()).toBe("failed");
    });

    expect(connection.getError()?.message).toBe(
      "agent_workbench_stream_ended",
    );
    expect(transport.getSnapshot).toHaveBeenCalledTimes(3);
  });
});

function snapshot(
  patch: { revision?: bigint; latestSequence?: bigint } = {},
) {
  return create(SessionSnapshotSchema, {
    sessionId: "session-1",
    streamEpoch: "epoch-1",
    revision: patch.revision ?? 1n,
    latestSequence: patch.latestSequence ?? 1n,
    status: SessionStatus.IDLE,
  });
}

function sequenceGapBatch(): SessionDeltaBatch {
  return create(SessionDeltaBatchSchema, {
    sessionId: "session-1",
    streamEpoch: "epoch-1",
    baseRevision: 1n,
    revision: 2n,
    firstSequence: 3n,
    lastSequence: 3n,
    digest: "sha256:gap",
    events: [
      create(AgentEventSchema, {
        envelope: {
          sessionId: "session-1",
          streamEpoch: "epoch-1",
          revision: 2n,
          sequence: 3n,
          itemId: "status-3",
          createdAt: "2026-07-16T10:00:00Z",
        },
        event: {
          case: "sessionStatusChanged",
          value: { status: SessionStatus.RUNNING },
        },
      }),
    ],
  });
}

function fakeTransport(
  overrides: Partial<AgentWorkbenchSessionTransport>,
): AgentWorkbenchSessionTransport {
  return {
    execute: vi.fn(
      async (_command: CommandEnvelope): Promise<CommandReceipt> => {
        throw new Error("not_implemented");
      },
    ),
    getSnapshot: vi.fn(async () => snapshot()),
    streamDeltas: () => emptyStream(),
    ...overrides,
  };
}

async function* oneBatch(batch: SessionDeltaBatch) {
  yield batch;
}

async function* emptyStream() {}

async function* waitForAbort(signal: AbortSignal) {
  await new Promise<void>((resolve) => {
    if (signal.aborted) resolve();
    else signal.addEventListener("abort", () => resolve(), { once: true });
  });
}
