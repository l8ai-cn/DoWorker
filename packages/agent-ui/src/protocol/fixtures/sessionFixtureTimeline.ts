import { create } from "@bufbuild/protobuf";

import {
  ContentBlockSchema,
  UnsupportedReason,
} from "@agent-cloud/proto/agent_workbench/v2/content_pb";
import {
  MessageTimelineItemSchema,
  TimelineItemContentSchema,
  TimelineItemSchema,
  type TimelineItem,
} from "@agent-cloud/proto/agent_workbench/v2/session_pb";
import {
  MessageRole,
  TimelineItemStatus,
} from "@agent-cloud/proto/agent_workbench/v2/session_state_pb";
import {
  ToolExecutionSchema,
  ToolPhase,
} from "@agent-cloud/proto/agent_workbench/v2/tool_pb";

const textEncoder = new TextEncoder();

function fixtureIdentity(semanticKey: string, sourceType?: string) {
  return {
    namespace: "agentcloud.fixture",
    semanticKey,
    schemaVersion: "1",
    sourceType,
  };
}

function createMessageContent() {
  return create(MessageTimelineItemSchema, {
    role: MessageRole.ASSISTANT,
    status: TimelineItemStatus.COMPLETED,
    content: [
      create(ContentBlockSchema, {
        contentId: "markdown-1",
        identity: fixtureIdentity("content.markdown"),
        content: {
          case: "markdown",
          value: { markdown: "## Workbench result\n\nAll media remains typed." },
        },
      }),
      create(ContentBlockSchema, {
        contentId: "image-1",
        identity: fixtureIdentity("content.image"),
        content: {
          case: "image",
          value: {
            artifactId: "image-1",
            representationId: "image-source",
            revision: 1n,
            mediaType: "image/png",
            filename: "result.png",
            altText: "Generated result",
          },
        },
      }),
      create(ContentBlockSchema, {
        contentId: "video-1",
        identity: fixtureIdentity("content.video"),
        content: {
          case: "video",
          value: {
            artifactId: "video-1",
            representationId: "video-playable",
            revision: 2n,
            mediaType: "video/mp4",
            filename: "result.mp4",
          },
        },
      }),
      create(ContentBlockSchema, {
        contentId: "presentation-1",
        identity: fixtureIdentity("content.presentation"),
        content: {
          case: "presentation",
          value: {
            artifactId: "deck-1",
            representationId: "deck-source",
            revision: 3n,
            mediaType:
              "application/vnd.openxmlformats-officedocument.presentationml.presentation",
            filename: "workbench.pptx",
          },
        },
      }),
      create(ContentBlockSchema, {
        contentId: "unknown-1",
        identity: fixtureIdentity("content.future", "acp.future-content"),
        content: {
          case: "unsupported",
          value: {
            identity: fixtureIdentity("content.future", "acp.future-content"),
            reason: UnsupportedReason.UNKNOWN,
            payload: {
              mediaType: "application/x-agent-workbench-future",
              data: Uint8Array.from([0, 255, 16, 128, 64, 10]),
            },
          },
        },
      }),
    ],
  });
}

function createToolContent() {
  return create(ToolExecutionSchema, {
    executionId: "tool-execution-1",
    identity: {
      namespace: "agentcloud.claude",
      semanticKey: "filesystem.read",
      schemaVersion: "1",
      sourceToolName: "Read",
    },
    category: "filesystem",
    phase: ToolPhase.RUNNING,
    input: {
      mediaType: "application/json",
      data: textEncoder.encode('{"path":"/workspace/README.md"}'),
    },
    results: [
      {
        resultId: "tool-result-1",
        primary: true,
        blocks: [
          create(ContentBlockSchema, {
            contentId: "tool-result-markdown-1",
            identity: fixtureIdentity("content.markdown"),
            content: {
              case: "markdown",
              value: { markdown: "README content is available." },
            },
          }),
        ],
      },
    ],
    title: "Read README",
  });
}

export function createFixtureHistory(): TimelineItem[] {
  return [
    create(TimelineItemSchema, {
      envelope: {
        sessionId: "session-lossless-1",
        streamEpoch: "epoch-lossless-1",
        revision: 9_007_199_254_740_992n,
        sequence: 9_007_199_254_740_994n,
        turnId: "turn-1",
        itemId: "message-1",
        createdAt: "2026-07-16T00:00:00Z",
      },
      content: create(TimelineItemContentSchema, {
        content: { case: "message", value: createMessageContent() },
      }),
    }),
    create(TimelineItemSchema, {
      envelope: {
        sessionId: "session-lossless-1",
        streamEpoch: "epoch-lossless-1",
        revision: 9_007_199_254_740_993n,
        sequence: 9_007_199_254_740_995n,
        turnId: "turn-1",
        itemId: "tool-1",
        parentId: "message-1",
        causationCommandId: "running-command-1",
        createdAt: "2026-07-16T00:00:01Z",
      },
      content: create(TimelineItemContentSchema, {
        content: { case: "toolExecution", value: createToolContent() },
      }),
    }),
  ];
}
