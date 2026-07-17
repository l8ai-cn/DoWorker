import { create } from "@bufbuild/protobuf";
import { describe, expect, it } from "vitest";

import { UnsupportedValueSchema } from "@do-worker/proto/agent_workbench/v2/content_pb";
import {
  AgentEventSchema,
  MessageTimelineItemSchema,
  SessionDeltaBatchSchema,
  SessionSnapshotSchema,
  TimelineItemContentSchema,
  TimelineItemSchema,
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

describe("agent session timeline event reduction", () => {
  it("preserves position and records the update envelope", () => {
    const state = applySessionSnapshot(
      snapshot([
        create(TimelineItemSchema, {
          envelope: envelope(5n, "message-1", 3n),
          content: assistantMessage(),
        }),
      ]),
    );
    const update = create(AgentEventSchema, {
      envelope: envelope(10n, "message-1", 5n),
      event: {
        case: "timelineItemUpdated",
        value: { content: assistantMessage() },
      },
    });

    const next = applyDeltaBatch(state, batch(update));

    expect(next.snapshot.history).toHaveLength(1);
    expect(next.snapshot.history[0]?.envelope).toMatchObject(update.envelope!);
  });

  it("clones appended event data before committing it", () => {
    const event = appended("message-1");
    const next = applyDeltaBatch(applySessionSnapshot(snapshot()), batch(event));

    event.envelope!.itemId = "mutated";

    expect(next.snapshot.history[0]?.envelope?.itemId).toBe("message-1");
  });

  it("clones unsupported payloads before committing them", () => {
    const unsupported = create(UnsupportedValueSchema, {
      identity: {
        namespace: "runner.event",
        semanticKey: "custom",
        schemaVersion: "1",
      },
      payload: {
        mediaType: "application/octet-stream",
        data: Uint8Array.from([1, 2, 3]),
      },
    });
    const event = create(AgentEventSchema, {
      envelope: envelope(10n, "unsupported-1", 5n),
      event: { case: "unsupported", value: unsupported },
    });
    const next = applyDeltaBatch(applySessionSnapshot(snapshot()), batch(event));

    unsupported.identity!.namespace = "mutated";
    unsupported.payload!.data[0] = 9;

    const content = next.snapshot.history[0]?.content?.content;
    expect(content?.case).toBe("unsupported");
    if (content?.case !== "unsupported") throw new Error("unsupported_missing");
    expect(content.value.identity?.namespace).toBe("runner.event");
    expect([...(content.value.payload?.data ?? [])]).toEqual([1, 2, 3]);
  });

  it("rejects unsupported reuse of a timeline item id", () => {
    const state = applySessionSnapshot(
      snapshot([
        create(TimelineItemSchema, {
          envelope: envelope(5n, "duplicate", 3n),
          content: assistantMessage(),
        }),
      ]),
    );
    const event = create(AgentEventSchema, {
      envelope: envelope(10n, "duplicate", 5n),
      event: {
        case: "unsupported",
        value: create(UnsupportedValueSchema, {
          identity: {
            namespace: "runner.event",
            semanticKey: "custom",
            schemaVersion: "1",
          },
        }),
      },
    });

    expect(() => applyDeltaBatch(state, batch(event))).toThrowError(
      new AgentSessionReductionError("timeline_item_conflict"),
    );
  });
});

function snapshot(
  history: ReturnType<typeof create<typeof TimelineItemSchema>>[] = [],
) {
  return create(SessionSnapshotSchema, {
    sessionId: "session-1",
    streamEpoch: "epoch-1",
    revision: 4n,
    latestSequence: 9n,
    status: SessionStatus.RUNNING,
    history,
  });
}

function batch(event: ReturnType<typeof create<typeof AgentEventSchema>>) {
  return create(SessionDeltaBatchSchema, {
    sessionId: "session-1",
    streamEpoch: "epoch-1",
    baseRevision: 4n,
    revision: 5n,
    firstSequence: 10n,
    lastSequence: 10n,
    digest: "sha256:batch",
    events: [event],
  });
}

function appended(itemId: string) {
  return create(AgentEventSchema, {
    envelope: envelope(10n, itemId, 5n),
    event: {
      case: "timelineItemAppended",
      value: { content: assistantMessage() },
    },
  });
}

function assistantMessage() {
  return create(TimelineItemContentSchema, {
    content: {
      case: "message",
      value: create(MessageTimelineItemSchema, {
        role: MessageRole.ASSISTANT,
        status: TimelineItemStatus.COMPLETED,
      }),
    },
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
