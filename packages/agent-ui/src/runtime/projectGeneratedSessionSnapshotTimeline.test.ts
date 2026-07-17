import { create } from "@bufbuild/protobuf";
import { describe, expect, it } from "vitest";

import {
  ContentBlockSchema,
} from "@do-worker/proto/agent_workbench/v2/content_pb";
import {
  MessageTimelineItemSchema,
  SessionSnapshotSchema,
  TimelineItemContentSchema,
  TimelineItemSchema,
} from "@do-worker/proto/agent_workbench/v2/session_pb";
import {
  MessageRole,
  SessionStatus,
  TimelineItemStatus,
} from "@do-worker/proto/agent_workbench/v2/session_state_pb";
import { projectGeneratedSessionSnapshot } from "./projectGeneratedSessionSnapshot";

const encoder = new TextEncoder();

describe("projectGeneratedSessionSnapshot message content", () => {
  it("preserves displayable blocks and emits evidence for unsupported blocks", () => {
    const message = create(MessageTimelineItemSchema, {
      role: MessageRole.USER,
      status: TimelineItemStatus.FAILED,
      content: [
        create(ContentBlockSchema, {
          contentId: "text",
          content: { case: "text", value: { text: "Plain text" } },
        }),
        create(ContentBlockSchema, {
          contentId: "markdown",
          content: {
            case: "markdown",
            value: { markdown: "**Markdown**" },
          },
        }),
        create(ContentBlockSchema, {
          contentId: "code",
          content: {
            case: "code",
            value: {
              code: "const answer = 42;",
              language: "ts",
              filename: "answer.ts",
            },
          },
        }),
        create(ContentBlockSchema, {
          contentId: "log",
          content: {
            case: "log",
            value: {
              level: "warn",
              message: "Watch this",
              createdAt: "2026-07-16T00:00:00Z",
              fields: {
                mediaType: "application/json",
                data: encoder.encode('{"attempt":2}'),
              },
            },
          },
        }),
        create(ContentBlockSchema, {
          contentId: "html",
          content: {
            case: "html",
            value: {
              securityProfile: "static-html-v1",
              payload: { case: "source", value: "<main>Ready</main>" },
            },
          },
        }),
        create(ContentBlockSchema, {
          contentId: "link",
          content: {
            case: "link",
            value: {
              url: "https://example.test/data",
              label: "Data",
              mediaType: "application/json",
            },
          },
        }),
        create(ContentBlockSchema, {
          contentId: "citation",
          content: {
            case: "citation",
            value: {
              citationId: "citation-1",
              label: "Source",
              excerpt: "Quoted evidence",
            },
          },
        }),
        create(ContentBlockSchema, {
          contentId: "error",
          content: {
            case: "error",
            value: {
              code: "content_failed",
              message: "Content failed",
            },
          },
        }),
        create(ContentBlockSchema, {
          contentId: "iframe",
          content: {
            case: "restrictedIframe",
            value: {
              rendererId: "renderer-1",
              protocolVersion: "2",
              url: "https://renderer.example.test",
              origin: "https://renderer.example.test",
              sandboxTokens: ["allow-scripts"],
              permissions: ["clipboard-read"],
              maxInboundBytes: 1024n,
              maxOutboundBytes: 2048n,
            },
          },
        }),
        create(ContentBlockSchema, {
          contentId: "preview",
          content: {
            case: "livePreview",
            value: {
              resourceId: "preview-1",
              securityProfile: "pod-preview-v1",
              sessionUrl: "https://preview.example.test",
            },
          },
        }),
      ],
    });
    const content = create(TimelineItemContentSchema, {
      content: { case: "message", value: message },
    });
    const snapshot = create(SessionSnapshotSchema, {
      sessionId: "session-all",
      streamEpoch: "epoch-all",
      revision: 1n,
      latestSequence: 1n,
      status: SessionStatus.RUNNING,
      history: [
        create(TimelineItemSchema, {
          envelope: {
            sessionId: "session-all",
            streamEpoch: "epoch-all",
            revision: 1n,
            sequence: 1n,
            itemId: "message-all",
            createdAt: "2026-07-16T00:00:01Z",
          },
          content,
        }),
      ],
    });

    const projected = projectGeneratedSessionSnapshot(snapshot, {
      title: "All timeline",
      agentLabel: "Codex",
      connection: "connected",
      interactionMode: "pty",
      hasOlderItems: false,
    });

    expect(projected.items).toEqual(
      expect.arrayContaining([
        expect.objectContaining({
          id: "message-all",
          kind: "message",
          role: "user",
          status: "failed",
          text: expect.stringMatching(
            /Plain text.*\*\*Markdown\*\*.*answer\.ts.*const answer = 42.*warn.*Watch this.*2026-07-16.*attempt.*static-html-v1.*Ready.*application\/json.*citation-1.*Quoted evidence.*content_failed.*Content failed/is,
          ),
        }),
        expect.objectContaining({
          id: "message-all:preview",
          kind: "system",
          title: "Unsupported content",
          detail: expect.stringMatching(
            /livePreview.*preview-1.*pod-preview-v1.*preview\.example\.test/is,
          ),
        }),
        expect.objectContaining({
          id: "message-all:iframe",
          kind: "system",
          title: "Unsupported content",
          detail: expect.stringMatching(
            /renderer-1.*allow-scripts.*clipboard-read.*1024.*2048/is,
          ),
        }),
      ]),
    );
  });
});
