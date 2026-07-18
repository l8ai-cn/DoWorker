import { create } from "@bufbuild/protobuf";
import { describe, expect, it } from "vitest";

import { ArtifactReferenceSchema } from "@do-worker/proto/agent_workbench/v2/artifact_pb";
import { ContentBlockSchema } from "@do-worker/proto/agent_workbench/v2/content_pb";
import {
  ApprovalTimelineItemSchema,
  TimelineItemContentSchema,
  TimelineItemSchema,
} from "@do-worker/proto/agent_workbench/v2/session_pb";
import { SessionStatus } from "@do-worker/proto/agent_workbench/v2/session_state_pb";
import { createLosslessSessionFixture } from "../protocol";
import { projectGeneratedSessionSnapshot as rootProjection } from "../index";
import { projectGeneratedSessionSnapshot as runtimeProjection } from "./index";

const options = {
  title: "Lossless workbench",
  agentLabel: "Claude",
  connection: "connected" as const,
  interactionMode: "acp" as const,
  hasOlderItems: true,
};

describe("projectGeneratedSessionSnapshot lossless fixture", () => {
  it("projects rich media, tool data, unsupported evidence, permissions, terminal, and error", () => {
    const { snapshot } = createLosslessSessionFixture();
    const toolContent = snapshot.history[1]?.content?.content;
    if (toolContent?.case !== "toolExecution") {
      throw new Error("lossless fixture tool is missing");
    }
    toolContent.value.results[0]?.blocks.push(
      create(ContentBlockSchema, {
        contentId: "tool-result-json-1",
        content: {
          case: "json",
          value: {
            value: {
              mediaType: "application/json",
              data: new TextEncoder().encode('{"lines":3}'),
            },
          },
        },
      }),
    );
    toolContent.value.results[0]?.artifacts.push(
      create(ArtifactReferenceSchema, {
        artifactId: "image-1",
        representationId: "image-source",
        revision: 1n,
        role: "preview",
      }),
    );
    const projected = runtimeProjection(snapshot, options);

    expect(projected).toMatchObject({
      sessionId: "session-lossless-1",
      title: "Lossless workbench",
      agentLabel: "Claude",
      connection: "connected",
      interactionMode: "acp",
      hasOlderItems: true,
      status: "failed",
      error: expect.stringContaining("fixture_runner_failed"),
      permissions: [
        {
          id: "permission-1",
          kind: "approval",
          title: "Publish presentation",
          description: "Allow the agent to publish revision 3.",
        },
      ],
      terminals: [
        {
          id: "terminal-1",
          label: "Workbench terminal",
          status: "connected",
          writable: true,
        },
      ],
    });

    expect(projected.items).toEqual(
      expect.arrayContaining([
        expect.objectContaining({
          id: "message-1",
          kind: "message",
          role: "assistant",
          status: "completed",
          text: expect.stringContaining("All media remains typed."),
        }),
        expect.objectContaining({
          kind: "artifact",
          artifactId: "image-1",
          filename: "result.png",
          mimeType: "image/png",
          schemaVersion: "1",
          status: "completed",
        }),
        expect.objectContaining({
          kind: "artifact",
          artifactId: "video-1",
          filename: "result.mp4",
          mimeType: "video/mp4",
          schemaVersion: "1",
          status: "completed",
        }),
        expect.objectContaining({
          kind: "artifact",
          artifactId: "deck-1",
          filename: "workbench.pptx",
          mimeType:
            "application/vnd.openxmlformats-officedocument.presentationml.presentation",
          schemaVersion: "1",
          status: "completed",
        }),
        expect.objectContaining({
          kind: "system",
          title: "Unsupported content",
          detail: expect.stringMatching(
            /agentsmesh\.fixture.*content\.future.*unknown.*00 ff 10 80 40 0a/is,
          ),
        }),
      ]),
    );

    const tool = projected.items.find((item) => item.kind === "tool");
    expect(tool).toMatchObject({
      id: "tool-1",
      identity: {
        namespace: "agentsmesh.claude",
        semanticKey: "filesystem.read",
        schemaVersion: "1",
      },
      input: '{\n  "path": "/workspace/README.md"\n}',
      inputValue: { path: "/workspace/README.md" },
      results: expect.arrayContaining([
        {
          id: "tool-result-1:tool-result-markdown-1",
          kind: "text",
          text: "README content is available.",
        },
        {
          id: "tool-result-1:tool-result-json-1",
          kind: "data",
          value: { lines: 3 },
        },
        {
          id: "tool-result-1:artifact:0",
          kind: "artifact",
          artifactId: "image-1",
          mediaType: "image/png",
          representationId: "image-source",
          revision: 1n,
          role: "preview",
          schemaVersion: "1",
        },
      ]),
      status: "running",
      title: "Read README",
    });
  });

  it("exports the same projection from runtime and package roots", () => {
    expect(rootProjection).toBe(runtimeProjection);
  });

  it("makes resync and resolved approval state explicit", () => {
    const { snapshot } = createLosslessSessionFixture();
    snapshot.status = SessionStatus.RESYNC_REQUIRED;
    snapshot.error = undefined;
    snapshot.history.push(
      create(TimelineItemSchema, {
        envelope: {
          sessionId: snapshot.sessionId,
          streamEpoch: snapshot.streamEpoch,
          revision: snapshot.revision,
          sequence: snapshot.latestSequence + 1n,
          itemId: "approval-resolved",
          createdAt: "2026-07-16T00:00:05Z",
        },
        content: create(TimelineItemContentSchema, {
          content: {
            case: "approval",
            value: create(ApprovalTimelineItemSchema, {
              permissionRequestId: "permission-2",
            }),
          },
        }),
      }),
    );

    const projected = runtimeProjection(snapshot, options);

    expect(projected.status).toBe("failed");
    expect(projected.error).toMatch(/resync/i);
    expect(projected.items).toContainEqual(
      expect.objectContaining({
        id: "approval-resolved",
        title: "Permission requested",
        status: "completed",
      }),
    );
  });
});
