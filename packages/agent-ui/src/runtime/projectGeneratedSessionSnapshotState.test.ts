import { create } from "@bufbuild/protobuf";
import { describe, expect, it } from "vitest";
import { ArtifactDescriptorSchema, ArtifactReferenceSchema, ArtifactStatus } from "@agent-cloud/proto/agent_workbench/v2/artifact_pb";
import { AgentErrorSchema } from "@agent-cloud/proto/agent_workbench/v2/command_pb";
import {
  ContentBlockSchema,
  UnsupportedReason,
  UnsupportedValueSchema,
} from "@agent-cloud/proto/agent_workbench/v2/content_pb";
import {
  ApprovalTimelineItemSchema,
  ArtifactReferenceTimelineItemSchema,
  ErrorTimelineItemSchema,
  PlanStepSchema,
  PlanTimelineItemSchema,
  ReasoningTimelineItemSchema,
  SessionSnapshotSchema,
  StatusTimelineItemSchema,
  SystemTimelineItemSchema,
  TimelineItemContentSchema,
  TimelineItemSchema,
  type TimelineItemContent,
} from "@agent-cloud/proto/agent_workbench/v2/session_pb";
import {
  PermissionRequestState,
  PlanStepStatus,
  SessionResourceStatus,
  SessionStatus,
  TerminalControlMode,
  TimelineItemStatus,
} from "@agent-cloud/proto/agent_workbench/v2/session_state_pb";
import { projectGeneratedSessionSnapshot } from "./projectGeneratedSessionSnapshot";
const encoder = new TextEncoder();
function timelineItem(itemId: string, sequence: bigint, content: TimelineItemContent) {
  return create(TimelineItemSchema, {
    envelope: {
      sessionId: "session-all",
      streamEpoch: "epoch-all",
      revision: sequence,
      sequence,
      itemId,
      createdAt: `2026-07-16T00:00:${sequence.toString().padStart(2, "0")}Z`,
    },
    content,
  });
}
function timelineContent(content: TimelineItemContent["content"]) {
  return create(TimelineItemContentSchema, { content });
}
describe("projectGeneratedSessionSnapshot state coverage", () => {
  it("projects every timeline case, plan, capabilities, and manifest artifact", () => {
    const reasoningBlock = create(ContentBlockSchema, {
      contentId: "reasoning-text",
      content: {
        case: "text",
        value: { text: "Checking constraints." },
      },
    });
    const systemBlock = create(ContentBlockSchema, {
      contentId: "system-text",
      content: { case: "text", value: { text: "System notice" } },
    });
    const steps = [
      create(PlanStepSchema, {
        stepId: "step-1",
        title: "Inspect",
        detail: "Read the protocol.",
        status: PlanStepStatus.COMPLETED,
      }),
      create(PlanStepSchema, {
        stepId: "step-2",
        title: "Implement",
        status: PlanStepStatus.RUNNING,
      }),
    ];
    const timelineError = create(AgentErrorSchema, {
      code: "timeline_failed",
      message: "A timeline operation failed.",
      retryable: false,
    });
    const unsupported = create(UnsupportedValueSchema, {
      identity: {
        namespace: "agentcloud.future",
        semanticKey: "timeline.future",
        schemaVersion: "9",
      },
      reason: UnsupportedReason.UNSUPPORTED,
      payload: {
        mediaType: "application/json",
        data: encoder.encode('{"future":true}'),
      },
    });
    const artifactReference = create(ArtifactReferenceSchema, {
      artifactId: "video-manifest",
      representationId: "playable",
      role: "preview",
    });
    const history = [
      timelineItem(
        "reasoning-1",
        1n,
        timelineContent({
          case: "reasoning",
          value: create(ReasoningTimelineItemSchema, {
            status: TimelineItemStatus.RUNNING,
            content: [reasoningBlock],
          }),
        }),
      ),
      timelineItem(
        "plan-1",
        2n,
        timelineContent({
          case: "plan",
          value: create(PlanTimelineItemSchema, { steps }),
        }),
      ),
      timelineItem(
        "artifact-reference-1",
        3n,
        timelineContent({
          case: "artifactReference",
          value: create(ArtifactReferenceTimelineItemSchema, {
            artifact: artifactReference,
            label: "Rendered video",
          }),
        }),
      ),
      timelineItem(
        "approval-1",
        4n,
        timelineContent({
          case: "approval",
          value: create(ApprovalTimelineItemSchema, {
            permissionRequestId: "question-1",
          }),
        }),
      ),
      timelineItem(
        "status-1",
        5n,
        timelineContent({
          case: "status",
          value: create(StatusTimelineItemSchema, {
            status: SessionStatus.WAITING,
            detail: "Waiting for an answer.",
          }),
        }),
      ),
      timelineItem(
        "error-1",
        6n,
        timelineContent({
          case: "error",
          value: create(ErrorTimelineItemSchema, { error: timelineError }),
        }),
      ),
      timelineItem(
        "system-1",
        7n,
        timelineContent({
          case: "system",
          value: create(SystemTimelineItemSchema, {
            content: [systemBlock],
          }),
        }),
      ),
      timelineItem(
        "unsupported-1",
        8n,
        timelineContent({ case: "unsupported", value: unsupported }),
      ),
      timelineItem("empty-1", 9n, create(TimelineItemContentSchema)),
    ];
    const artifact = create(ArtifactDescriptorSchema, {
      artifactId: "video-manifest",
      revision: 2n,
      filename: "source.mov",
      mediaType: "video/quicktime",
      role: "source",
      status: ArtifactStatus.READY,
      representations: [
        {
          representationId: "original",
          revision: 2n,
          filename: "source.mov",
          mediaType: "video/quicktime",
          role: "source",
          status: ArtifactStatus.READY,
        },
        {
          representationId: "playable",
          revision: 2n,
          filename: "result.mp4",
          mediaType: "video/mp4",
          role: "preview",
          status: ArtifactStatus.READY,
        },
      ],
      manifest: {
        manifest: {
          case: "video",
          value: { playableRepresentationId: "playable" },
        },
      },
    });
    const snapshot = create(SessionSnapshotSchema, {
      sessionId: "session-all",
      streamEpoch: "epoch-all",
      revision: 9n,
      latestSequence: 9n,
      status: SessionStatus.WAITING,
      history,
      permissionRequests: [
        {
          permissionRequestId: "question-1",
          state: PermissionRequestState.PENDING,
          request: {
            case: "questionnaire",
            value: {
              title: "Choose output",
              questions: [
                {
                  questionId: "format",
                  prompt: "Which format?",
                  header: "Format",
                  options: [
                    { label: "PPTX", description: "Editable deck" },
                  ],
                  multiple: false,
                  allowCustom: true,
                  secret: false,
                },
              ],
            },
          },
        },
      ],
      artifacts: [artifact],
      capabilities: {
        protocolVersion: "2",
        commandSchemas: [
          {
            namespace: "proto.agent_workbench.v2",
            semanticKey: "send_prompt",
            schemaVersion: "2",
            actions: ["session.send"],
          },
          {
            namespace: "proto.agent_workbench.v2",
            semanticKey: "interrupt",
            schemaVersion: "2",
            actions: ["session.interrupt"],
          },
          {
            namespace: "proto.agent_workbench.v2",
            semanticKey: "change_configuration",
            schemaVersion: "2",
            actions: ["session.configure"],
          },
          {
            namespace: "proto.agent_workbench.v2",
            semanticKey: "resolve_permission",
            schemaVersion: "2",
            actions: ["session.permission.resolve"],
          },
        ],
        terminalOperations: ["terminal.input"],
      },
      grants: [
        {
          grantId: "grant-1",
          issuer: "backend",
          subject: "user-1",
          sessionId: "session-all",
          actions: [
            "session.send",
            "session.interrupt",
            "session.permission.resolve",
            "terminal.input",
          ],
          issuedAt: "2026-07-16T00:00:00Z",
        },
      ],
      resources: [
        {
          resourceId: "terminal-1",
          label: "Main terminal",
          status: SessionResourceStatus.READY,
          resource: {
            case: "terminal",
            value: {
              writable: true,
              controlMode: TerminalControlMode.SURFACE,
            },
          },
        },
      ],
      error: {
        code: "session_waiting",
        message: "The session needs input.",
        retryable: true,
        details: {
          mediaType: "application/json",
          data: encoder.encode('{"permission":"question-1"}'),
        },
      },
    });
    const projected = projectGeneratedSessionSnapshot(snapshot, {
      title: "All timeline",
      agentLabel: "Codex",
      connection: "connected",
      interactionMode: "pty",
      hasOlderItems: false,
    });
    expect(projected.status).toBe("waiting");
    expect(projected.capabilities).toEqual({
      sendMessage: true,
      interrupt: true,
      resolvePermission: true,
      updateConfiguration: false,
      terminal: true,
    });
    expect(projected.plan).toEqual([
      { id: "step-1", title: "Inspect", status: "completed" },
      { id: "step-2", title: "Implement", status: "running" },
    ]);
    expect(projected.permissions[0]).toMatchObject({
      id: "question-1",
      kind: "question",
      title: "Choose output",
      questions: [
        {
          id: "format",
          prompt: "Which format?",
          options: [{ label: "PPTX", description: "Editable deck" }],
        },
      ],
    });
    expect(projected.items).toEqual(
      expect.arrayContaining([
        expect.objectContaining({
          id: "reasoning-1",
          kind: "reasoning",
          detail: "Checking constraints.",
          status: "running",
        }),
        expect.objectContaining({
          id: "plan-1",
          title: "Plan updated",
          detail: expect.stringContaining("Inspect"),
        }),
        expect.objectContaining({
          artifactId: "video-manifest",
          filename: "result.mp4",
          mimeType: "video/mp4",
          role: "preview",
        }),
        expect.objectContaining({
          id: "approval-1",
          title: "Permission requested",
          detail: expect.stringContaining("question-1"),
        }),
        expect.objectContaining({
          id: "status-1",
          title: "Session waiting",
          detail: "Waiting for an answer.",
        }),
        expect.objectContaining({
          id: "error-1",
          kind: "error",
          title: "timeline_failed",
          detail: "A timeline operation failed.",
        }),
        expect.objectContaining({
          id: "system-1",
          title: "System",
          detail: "System notice",
        }),
        expect.objectContaining({
          id: "unsupported-1",
          title: "Unsupported timeline item",
          detail: expect.stringMatching(/timeline\.future.*unsupported.*future/is),
        }),
        expect.objectContaining({
          id: "empty-1",
          title: "Unsupported timeline item",
          detail: expect.stringContaining("missing timeline content"),
        }),
      ]),
    );
    expect(projected.error).toMatch(
      /session_waiting.*The session needs input.*permission.*question-1/is,
    );
  });
});
