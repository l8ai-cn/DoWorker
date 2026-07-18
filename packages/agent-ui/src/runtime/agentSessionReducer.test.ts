import { create } from "@bufbuild/protobuf";
import { describe, expect, it } from "vitest";

import {
  ArtifactDescriptorSchema,
  ArtifactRevisionSchema,
  ArtifactStatus,
  type ArtifactDescriptor,
} from "@do-worker/proto/agent_workbench/v2/artifact_pb";
import {
  CommandReceiptState,
  CommandReceiptSchema,
  type CommandReceipt,
} from "@do-worker/proto/agent_workbench/v2/command_pb";
import {
  AgentEventSchema,
  MessageTimelineItemSchema,
  SessionDeltaBatchSchema,
  SessionSnapshotSchema,
  type TimelineItem,
  TimelineItemContentSchema,
} from "@do-worker/proto/agent_workbench/v2/session_pb";
import {
  MessageRole,
  SessionStatus,
  TimelineItemStatus,
} from "@do-worker/proto/agent_workbench/v2/session_state_pb";
import {
  AgentSessionReductionError,
  applyDeltaBatch,
  applySessionSnapshot,
} from "./agentSessionReducer";

describe("agentSessionReducer", () => {
  it("applies a valid delta batch atomically", () => {
    const state = applySessionSnapshot(
      snapshot({
        commandReceipts: [receipt(CommandReceiptState.ACCEPTED)],
      }),
    );
    const batch = deltaBatch({
      events: [
        messageAppended(10n, "message-1"),
        receiptChanged(11n, CommandReceiptState.RUNNING),
      ],
      firstSequence: 10n,
      lastSequence: 11n,
      revision: 5n,
    });

    const next = applyDeltaBatch(state, batch);

    expect(next.status).toBe("ready");
    expect(next.snapshot.revision).toBe(5n);
    expect(next.snapshot.latestSequence).toBe(11n);
    expect(next.snapshot.history).toHaveLength(1);
    expect(next.snapshot.commandReceipts[0]?.state).toBe(CommandReceiptState.RUNNING);
  });

  it("returns the same state for an identical duplicate batch", () => {
    const batch = deltaBatch({
      events: [messageAppended(10n, "message-1")],
      firstSequence: 10n,
      lastSequence: 10n,
      revision: 5n,
    });
    const applied = applyDeltaBatch(applySessionSnapshot(snapshot()), batch);

    expect(applyDeltaBatch(applied, batch)).toBe(applied);
  });

  it("requires resync for a duplicate range with a different digest", () => {
    const original = deltaBatch({
      digest: "sha256:original",
      events: [messageAppended(10n, "message-1")],
      firstSequence: 10n,
      lastSequence: 10n,
      revision: 5n,
    });
    const applied = applyDeltaBatch(applySessionSnapshot(snapshot()), original);
    const conflict = create(SessionDeltaBatchSchema, {
      ...original,
      digest: "sha256:conflict",
    });

    const next = applyDeltaBatch(applied, conflict);

    expect(next.status).toBe("resync_required");
    expect(next.resyncReason).toBe("digest_conflict");
    expect(next.snapshot).toBe(applied.snapshot);
  });

  it.each([
    ["sequence_gap", { firstSequence: 11n }],
    ["stream_epoch_changed", { streamEpoch: "epoch-2" }],
    ["base_revision_mismatch", { baseRevision: 3n }],
  ] as const)("keeps readable state and requests resync for %s", (reason, patch) => {
    const state = applySessionSnapshot(snapshot());
    const batch = create(SessionDeltaBatchSchema, {
      ...deltaBatch({
        events: [messageAppended(10n, "message-1")],
        firstSequence: 10n,
        lastSequence: 10n,
        revision: 5n,
      }),
      ...patch,
    });

    const next = applyDeltaBatch(state, batch);

    expect(next.status).toBe("resync_required");
    expect(next.resyncReason).toBe(reason);
    expect(next.snapshot).toBe(state.snapshot);
  });

  it("does not mutate state when an event is invalid", () => {
    const state = applySessionSnapshot(snapshot());
    const invalidUpdate = create(AgentEventSchema, {
      envelope: envelope(10n, "missing-item", 5n),
      event: {
        case: "timelineItemUpdated",
        value: { content: assistantMessage("replacement") },
      },
    });

    expect(() =>
      applyDeltaBatch(
        state,
        deltaBatch({
          events: [invalidUpdate],
          firstSequence: 10n,
          lastSequence: 10n,
          revision: 5n,
        }),
      ),
    ).toThrowError(new AgentSessionReductionError("timeline_item_missing"));
    expect(state.snapshot.history).toHaveLength(0);
    expect(state.snapshot.revision).toBe(4n);
  });

  it("rejects command receipt changes after a terminal state", () => {
    const terminalSnapshot = snapshot({
      commandReceipts: [receipt(CommandReceiptState.SUCCEEDED)],
    });
    const state = applySessionSnapshot(terminalSnapshot);

    expect(() =>
      applyDeltaBatch(
        state,
        deltaBatch({
          events: [receiptChanged(10n, CommandReceiptState.RUNNING)],
          firstSequence: 10n,
          lastSequence: 10n,
          revision: 5n,
        }),
      ),
    ).toThrowError(new AgentSessionReductionError("receipt_terminal"));
  });

  it("keeps receipt metadata updates without inventing a state transition", () => {
    const state = applySessionSnapshot(
      snapshot({
        commandReceipts: [
          receipt(CommandReceiptState.RECEIVED, {
            updatedAt: "2026-07-16T00:00:00Z",
          }),
        ],
      }),
    );

    const next = applyDeltaBatch(
      state,
      deltaBatch({
        events: [
          receiptChanged(10n, CommandReceiptState.RECEIVED, {
            updatedAt: "2026-07-16T00:00:01Z",
          }),
        ],
        firstSequence: 10n,
        lastSequence: 10n,
        revision: 5n,
      }),
    );

    expect(next.snapshot.commandReceipts[0]?.updatedAt).toBe(
      "2026-07-16T00:00:01Z",
    );
  });

  it("preserves artifact revision history across live deltas", () => {
    const state = applySessionSnapshot(
      snapshot({ artifacts: [artifact(1n, "tool-one")] }),
    );

    const next = applyDeltaBatch(
      state,
      deltaBatch({
        events: [artifactChanged(10n, artifact(2n, "tool-two"))],
        firstSequence: 10n,
        lastSequence: 10n,
        revision: 5n,
      }),
    );

    expect(next.snapshot.artifacts[0]?.revision).toBe(2n);
    expect(
      next.snapshot.artifacts[0]?.revisions.map(
        (revision) => revision.provenance?.toolExecutionId,
      ),
    ).toEqual(["tool-one", "tool-two"]);
  });

});

function snapshot(
  patch: {
    artifacts?: ArtifactDescriptor[];
    commandReceipts?: CommandReceipt[];
    history?: TimelineItem[];
  } = {},
) {
  return create(SessionSnapshotSchema, {
    sessionId: "session-1",
    streamEpoch: "epoch-1",
    revision: 4n,
    latestSequence: 9n,
    status: SessionStatus.RUNNING,
    artifacts: patch.artifacts ?? [],
    commandReceipts: patch.commandReceipts ?? [],
    history: patch.history ?? [],
  });
}

function deltaBatch(input: {
  digest?: string;
  events: ReturnType<typeof create<typeof AgentEventSchema>>[];
  firstSequence: bigint;
  lastSequence: bigint;
  revision: bigint;
}) {
  return create(SessionDeltaBatchSchema, {
    sessionId: "session-1",
    streamEpoch: "epoch-1",
    baseRevision: 4n,
    digest: input.digest ?? "sha256:batch-1",
    ...input,
  });
}

function messageAppended(sequence: bigint, itemId: string) {
  return create(AgentEventSchema, {
    envelope: envelope(sequence, itemId, 5n),
    event: {
      case: "timelineItemAppended",
      value: { content: assistantMessage("Done") },
    },
  });
}

function artifactChanged(sequence: bigint, value: ArtifactDescriptor) {
  return create(AgentEventSchema, {
    envelope: envelope(sequence, `artifact-${sequence}`, 5n),
    event: {
      case: "artifactChanged",
      value: { artifact: value },
    },
  });
}

function artifact(revision: bigint, toolExecutionId: string) {
  return create(ArtifactDescriptorSchema, {
    artifactId: "artifact-1",
    revision,
    filename: "result.png",
    mediaType: "image/png",
    status: ArtifactStatus.READY,
    revisions: [
      create(ArtifactRevisionSchema, {
        revision,
        provenance: { toolExecutionId },
      }),
    ],
  });
}

function assistantMessage(text: string) {
  return create(TimelineItemContentSchema, {
    content: {
      case: "message",
      value: create(MessageTimelineItemSchema, {
        role: MessageRole.ASSISTANT,
        status: TimelineItemStatus.COMPLETED,
        content: [],
      }),
    },
  });
}

function receiptChanged(
  sequence: bigint,
  state: CommandReceiptState,
  patch: { updatedAt?: string } = {},
) {
  return create(AgentEventSchema, {
    envelope: envelope(sequence, `receipt-${sequence}`, 5n),
    event: {
      case: "commandReceiptChanged",
      value: { receipt: receipt(state, patch) },
    },
  });
}

function receipt(
  state: CommandReceiptState,
  patch: { updatedAt?: string } = {},
) {
  return create(CommandReceiptSchema, {
    sessionId: "session-1",
    commandId: "command-1",
    payloadDigest: "sha256:command-1",
    state,
    ...patch,
  });
}

function envelope(sequence: bigint, itemId: string, revision: bigint) {
  return {
    sessionId: "session-1",
    streamEpoch: "epoch-1",
    revision,
    sequence,
    itemId,
    createdAt: "2026-07-16T00:00:00Z",
  };
}
