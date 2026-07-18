import { describe, expect, it } from "vitest";

import type { AnyBlock } from "@/lib/blocks";
import type { EmbeddedSession } from "@/embed-session-api";
import { projectEmbeddedWorkspaceSnapshot } from "./embeddedWorkspaceProjection";

const session: EmbeddedSession = {
  agentLabel: "codex-cli",
  id: "session-1",
  interactionMode: "pty",
  podKey: "pod-1",
  status: "running",
  title: "Review the repository",
};

const ctx = {
  agent: null,
  depth: 0,
  turn: 1,
  timestamp: 1,
  responseId: "response-1",
  itemId: null,
};

function project(
  input: Omit<
    Parameters<typeof projectEmbeddedWorkspaceSnapshot>[0],
    "configuration" | "configurationConnected"
  >,
) {
  return projectEmbeddedWorkspaceSnapshot({
    ...input,
    configuration: {
      model: "",
      permissionMode: "",
      supportedPermissionModes: [],
    },
    configurationConnected: false,
  });
}

describe("projectEmbeddedWorkspaceSnapshot", () => {
  it("projects conversation, activity, approvals, capabilities, and terminals", () => {
    const blocks = [
      {
        type: "user_message",
        ctx: { ...ctx, itemId: "user-1", responseId: "" },
        content: [{ type: "input_text", text: "Inspect authentication" }],
      },
      { type: "reasoning_chunk", ctx, text: "Tracing the session boundary." },
      {
        type: "tool_group",
        ctx: { ...ctx, itemId: "tool-1" },
        executions: [
          {
            name: "read_file",
            arguments: { path: "auth.ts" },
            argsSummary: "auth.ts",
            callId: "call-1",
            agentName: "codex-cli",
            executedBy: "server",
            output: "export class AuthManager {}",
          },
        ],
        iteration: 1,
      },
      {
        type: "text_done",
        ctx: { ...ctx, itemId: "assistant-1" },
        fullText: "The authentication boundary is duplicated.",
        hasCodeBlocks: false,
      },
      {
        type: "file",
        ctx: { ...ctx, itemId: "file-item-1" },
        fileId: "file-1",
        filename: "architecture.png",
        contentType: "image/png",
      },
      {
        type: "elicitation",
        ctx: { ...ctx, itemId: "approval-1" },
        elicitationId: "permission-1",
        message: "Allow editing auth.ts?",
        phase: "tool_call",
        policyName: "workspace-write",
        contentPreview: "auth.ts",
        requestedSchema: {},
        status: "pending",
        response: null,
      },
    ] as AnyBlock[];

    const snapshot = project({
      activeResponse: {
        responseId: "response-1",
        state: "streaming",
        error: null,
      },
      blocks,
      capabilities: {
        interrupt: true,
        resolvePermission: true,
        sendMessage: true,
        terminal: true,
        updateConfiguration: false,
      },
      connection: "connected",
      error: null,
      hasOlderItems: true,
      session,
      status: "running",
      terminals: [
        {
          id: "terminal-main",
          label: "main",
          status: "connected",
          writable: true,
        },
      ],
    });

    expect(snapshot).toMatchObject({
      sessionId: "session-1",
      title: "Review the repository",
      agentLabel: "codex-cli",
      status: "running",
      connection: "connected",
      interactionMode: "pty",
      hasOlderItems: true,
      capabilities: {
        sendMessage: true,
        interrupt: true,
        resolvePermission: true,
        terminal: true,
      },
      terminals: [{ id: "terminal-main", writable: true }],
      permissions: [
        {
          id: "permission-1",
          kind: "approval",
          title: "Allow editing auth.ts?",
          description: "auth.ts",
        },
      ],
    });
    expect(snapshot.items).toEqual(
      expect.arrayContaining([
        expect.objectContaining({
          kind: "message",
          role: "user",
          text: "Inspect authentication",
        }),
        expect.objectContaining({
          kind: "reasoning",
          detail: "Tracing the session boundary.",
          status: "running",
        }),
        expect.objectContaining({
          kind: "tool",
          title: "read_file",
          input: expect.stringContaining("auth.ts"),
          output: "export class AuthManager {}",
          status: "completed",
        }),
        expect.objectContaining({
          kind: "message",
          role: "assistant",
          text: "The authentication boundary is duplicated.",
        }),
        expect.objectContaining({
          kind: "artifact",
          artifactId: "file-1",
          filename: "architecture.png",
          mimeType: "image/png",
        }),
      ]),
    );
  });

  it("projects structured agent questions without reducing them to approval", () => {
    const snapshot = project({
      activeResponse: null,
      blocks: [
        {
          type: "elicitation",
          ctx: { ...ctx, itemId: "question-1" },
          elicitationId: "permission-question-1",
          message: "Choose the delivery format",
          phase: "tool_call",
          policyName: "request_user_input",
          contentPreview: "",
          requestedSchema: {},
          askUserQuestion: {
            questions: [
              {
                id: "format",
                header: "Format",
                question: "Which result should be generated?",
                options: [
                  { label: "PPTX", description: "Editable presentation" },
                  { label: "MP4", description: "Rendered video" },
                ],
                multiSelect: false,
                isOther: true,
              },
            ],
          },
          status: "pending",
          response: null,
        },
      ] as AnyBlock[],
      capabilities: {
        interrupt: true,
        resolvePermission: true,
        sendMessage: true,
        terminal: false,
        updateConfiguration: false,
      },
      connection: "connected",
      error: null,
      hasOlderItems: false,
      session,
      status: "waiting",
      terminals: [],
    });

    expect(snapshot.permissions).toEqual([
      {
        id: "permission-question-1",
        kind: "question",
        title: "Choose the delivery format",
        questions: [
          {
            id: "format",
            prompt: "Which result should be generated?",
            header: "Format",
            options: [
              { label: "PPTX", description: "Editable presentation" },
              { label: "MP4", description: "Rendered video" },
            ],
            multiple: false,
            allowCustom: true,
            secret: false,
          },
        ],
      },
    ]);
  });

  it("projects generated media and presentation files from file changes", () => {
    const snapshot = project({
      activeResponse: null,
      blocks: [
        {
          type: "tool_group",
          ctx: { ...ctx, itemId: "tool-file-change" },
          executions: [
            {
              name: "fileChange",
              arguments: {
                changes: [
                  { path: "deliverables/preview.png", kind: { type: "create" }, diff: "" },
                  { path: "deliverables/deck.pptx", kind: { type: "create" }, diff: "" },
                  { path: "src/main.ts", kind: { type: "create" }, diff: "" },
                ],
              },
              argsSummary: "3 files",
              callId: "file-change-1",
              agentName: "codex-cli",
              executedBy: "server",
              output: "Created deliverables",
            },
          ],
          iteration: 1,
        },
      ] as AnyBlock[],
      capabilities: {
        interrupt: true,
        resolvePermission: true,
        sendMessage: true,
        terminal: false,
        updateConfiguration: false,
      },
      connection: "connected",
      error: null,
      hasOlderItems: false,
      session,
      status: "idle",
      terminals: [],
    });

    expect(snapshot.items).toEqual(
      expect.arrayContaining([
        expect.objectContaining({
          kind: "artifact",
          artifactId: "workspace:deliverables/preview.png",
          filename: "preview.png",
          mimeType: "image/png",
        }),
        expect.objectContaining({
          kind: "artifact",
          artifactId: "workspace:deliverables/deck.pptx",
          filename: "deck.pptx",
          mimeType:
            "application/vnd.openxmlformats-officedocument.presentationml.presentation",
        }),
      ]),
    );
    expect(snapshot.items).not.toEqual(
      expect.arrayContaining([
        expect.objectContaining({ artifactId: "workspace:src/main.ts" }),
      ]),
    );
  });

  it("does not advertise a terminal for ACP sessions", () => {
    const snapshot = project({
      activeResponse: null,
      blocks: [],
      capabilities: {
        interrupt: true,
        resolvePermission: true,
        sendMessage: true,
        terminal: true,
        updateConfiguration: false,
      },
      connection: "connected",
      error: null,
      hasOlderItems: false,
      session: { ...session, interactionMode: "acp" },
      status: "idle",
      terminals: [
        {
          id: "terminal-main",
          label: "main",
          status: "connected",
          writable: true,
        },
      ],
    });

    expect(snapshot.capabilities.terminal).toBe(false);
    expect(snapshot.terminals).toEqual([]);
  });

  it("preserves an unselected permission mode until the Runner confirms one", () => {
    const snapshot = projectEmbeddedWorkspaceSnapshot({
      activeResponse: null,
      blocks: [],
      capabilities: {
        interrupt: true,
        resolvePermission: true,
        sendMessage: true,
        terminal: false,
        updateConfiguration: false,
      },
      configuration: {
        model: "",
        permissionMode: "",
        supportedPermissionModes: ["bypass", "ask_dangerous", "ask_any_write"],
      },
      configurationConnected: true,
      connection: "connected",
      error: null,
      hasOlderItems: false,
      session: { ...session, interactionMode: "acp" },
      status: "idle",
      terminals: [],
    });

    expect(snapshot.capabilities.updateConfiguration).toBe(true);
    expect(snapshot.configuration).toEqual([
      expect.objectContaining({
        id: "permissionMode",
        value: "",
      }),
    ]);
  });
});
