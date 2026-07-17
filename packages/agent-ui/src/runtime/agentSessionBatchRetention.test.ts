import { create } from "@bufbuild/protobuf";
import { describe, expect, it } from "vitest";

import {
  AgentEventSchema,
  SessionDeltaBatchSchema,
  SessionSnapshotSchema,
} from "@do-worker/proto/agent_workbench/v2/session_pb";
import { SessionStatus } from "@do-worker/proto/agent_workbench/v2/session_state_pb";
import {
  applyDeltaBatch,
  applySessionSnapshot,
} from "./agentSessionReducer";

describe("agent session batch retention", () => {
  it("bounds duplicate signatures and resyncs for an evicted old batch", () => {
    let state = applySessionSnapshot(
      create(SessionSnapshotSchema, {
        sessionId: "session-1",
        streamEpoch: "epoch-1",
        status: SessionStatus.RUNNING,
      }),
    );
    const first = batch(1n);

    for (let revision = 1n; revision <= 257n; revision += 1n) {
      state = applyDeltaBatch(state, batch(revision));
    }

    expect(state.appliedBatches.size).toBe(256);
    const replay = applyDeltaBatch(state, first);
    expect(replay.status).toBe("resync_required");
    expect(replay.resyncReason).toBe("base_revision_mismatch");
  });
});

function batch(revision: bigint) {
  return create(SessionDeltaBatchSchema, {
    sessionId: "session-1",
    streamEpoch: "epoch-1",
    baseRevision: revision - 1n,
    revision,
    firstSequence: revision,
    lastSequence: revision,
    digest: `sha256:batch-${revision}`,
    events: [
      create(AgentEventSchema, {
        envelope: {
          sessionId: "session-1",
          streamEpoch: "epoch-1",
          revision,
          sequence: revision,
          itemId: `status-${revision}`,
          createdAt: "2026-07-16T00:00:00Z",
        },
        event: {
          case: "sessionStatusChanged",
          value: { status: SessionStatus.RUNNING },
        },
      }),
    ],
  });
}
