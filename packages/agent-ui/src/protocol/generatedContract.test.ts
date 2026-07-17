import { create } from "@bufbuild/protobuf";
import { describe, expect, it } from "vitest";

import {
  ArtifactActionCommandSchema,
  CommandEnvelopeSchema,
} from "@do-worker/proto/agent_workbench/v2/command_pb";
import {
  HtmlContentSchema,
} from "@do-worker/proto/agent_workbench/v2/content_pb";
import {
  AgentEventSchema,
  MessageTimelineItemSchema,
  PermissionRequestSchema,
  SessionSnapshotSchema,
  TimelineItemContentSchema,
} from "@do-worker/proto/agent_workbench/v2/session_pb";
import {
  MessageRole,
  PermissionRequestState,
  SessionStatus,
  TimelineItemStatus,
} from "@do-worker/proto/agent_workbench/v2/session_state_pb";

describe("generated Agent Workbench V2 contract", () => {
  it("keeps commands as generated oneofs with explicit artifact action versions", () => {
    const action = create(ArtifactActionCommandSchema, {
      artifactId: "image-1",
      representationId: "source",
      baseRevision: 7n,
      clientActionId: "action-1",
      actionType: "image.edit",
      actionSchemaVersion: "1",
    });
    const command = create(CommandEnvelopeSchema, {
      sessionId: "session-1",
      streamEpoch: "epoch-1",
      commandId: "command-1",
      payloadDigest: "sha256:abc",
      issuedAt: "2026-07-16T00:00:00Z",
      command: {
        case: "artifactAction",
        value: action,
      },
    });

    expect(command.command.case).toBe("artifactAction");
    expect(action.actionSchemaVersion).toBe("1");
  });

  it("uses typed timeline and session states", () => {
    const content = create(TimelineItemContentSchema, {
      content: {
        case: "message",
        value: create(MessageTimelineItemSchema, {
          role: MessageRole.ASSISTANT,
          status: TimelineItemStatus.COMPLETED,
        }),
      },
    });
    const snapshot = create(SessionSnapshotSchema, {
      sessionId: "session-1",
      streamEpoch: "epoch-1",
      revision: 4n,
      latestSequence: 9n,
      status: SessionStatus.RUNNING,
    });

    expect(content.content.case).toBe("message");
    expect(snapshot.status).toBe(SessionStatus.RUNNING);
  });

  it("carries timeline event ordering in one envelope", () => {
    const event = create(AgentEventSchema, {
      envelope: {
        sessionId: "session-1",
        streamEpoch: "epoch-1",
        revision: 5n,
        sequence: 10n,
        itemId: "message-1",
        createdAt: "2026-07-16T00:00:01Z",
      },
      event: {
        case: "timelineItemAppended",
        value: {
          content: create(TimelineItemContentSchema, {
            content: {
              case: "message",
              value: create(MessageTimelineItemSchema, {
                role: MessageRole.ASSISTANT,
                status: TimelineItemStatus.STREAMING,
              }),
            },
          }),
        },
      },
    });

    expect(event.envelope?.sequence).toBe(10n);
    expect(event.event.case).toBe("timelineItemAppended");
    expect(event.event.value).not.toHaveProperty("envelope");
  });

  it("makes inline and artifact HTML payloads mutually exclusive", () => {
    const html = create(HtmlContentSchema, {
      securityProfile: "static-html-v1",
      payload: {
        case: "source",
        value: "<main>ready</main>",
      },
    });

    expect(html.payload.case).toBe("source");
  });

  it("keeps terminal session errors and permission resolutions in snapshots", () => {
    const resolution = {
      permissionRequestId: "permission-1",
      decision: 1,
      resolvedAt: "2026-07-16T00:00:02Z",
    };
    const request = create(PermissionRequestSchema, {
      permissionRequestId: "permission-1",
      state: PermissionRequestState.RESOLVED,
      resolution,
    } as never);
    const snapshot = create(SessionSnapshotSchema, {
      sessionId: "session-1",
      streamEpoch: "epoch-1",
      status: SessionStatus.FAILED,
      error: {
        code: "runner_failed",
        message: "Runner stopped unexpectedly.",
        retryable: true,
      },
    } as never);

    expect(request).toHaveProperty("resolution");
    expect(snapshot).toHaveProperty("error");
  });
});
