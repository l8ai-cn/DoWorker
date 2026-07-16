import type { AcpSessionState } from "@/stores/acpSessionTypes";

import { projectWebAcpSnapshot } from "./webAcpSnapshot";

const session: AcpSessionState = {
  messages: [
    { text: "Inspect the release", role: "user", timestamp: 1, complete: true },
    { text: "I am checking it.", role: "assistant", timestamp: 2, complete: false },
  ],
  toolCalls: {
    tool: {
      toolCallId: "tool",
      toolName: "Bash",
      status: "in_progress",
      argumentsJson: "{\"command\":\"pnpm test\"}",
      timestamp: 3,
    },
  },
  plan: [
    { title: "Read changes", status: "completed" },
    { title: "Run tests", status: "in_progress" },
  ],
  thinkings: [{ text: "Finding affected packages", timestamp: 2.5, complete: true }],
  logs: [{ level: "error", message: "One check failed", timestamp: 4 }],
  state: "processing",
  pendingPermissions: [
    {
      requestId: "permission-1",
      toolName: "Bash",
      argumentsJson: "{\"command\":\"pnpm release\"}",
      description: "Run release command",
    },
    {
      requestId: "permission-2",
      toolName: "requestUserInput",
      argumentsJson: JSON.stringify({
        questions: [
          {
            id: "format",
            header: "Format",
            question: "Which deliverables should be generated?",
            options: [
              { label: "PPTX", description: "Editable presentation" },
              { label: "MP4", description: "Rendered video" },
            ],
            multiSelect: true,
            isOther: true,
            isSecret: false,
          },
        ],
      }),
      description: "Choose deliverables",
    },
  ],
  configuration: {
    permissionMode: "default",
    model: "gpt-5",
    supportedPermissionModes: ["default"],
  },
};

it("projects the Web ACP session into the shared workspace snapshot", () => {
  const snapshot = projectWebAcpSnapshot({
    agentLabel: "Codex",
    connection: "connected",
    sessionId: "web-acp:pod-1",
    session,
    title: "Release audit",
  });

  expect(snapshot.sessionId).toBe("web-acp:pod-1");
  expect(snapshot.status).toBe("running");
  expect(snapshot.items.map((item) => item.kind)).toEqual([
    "message",
    "message",
    "reasoning",
    "tool",
    "error",
  ]);
  expect(snapshot.plan[1]).toMatchObject({ title: "Run tests", status: "running" });
  expect(snapshot.permissions[0]).toMatchObject({
    id: "permission-1",
    kind: "approval",
    title: "Bash",
  });
  expect(snapshot.permissions[1]).toEqual({
    id: "permission-2",
    kind: "question",
    title: "Choose deliverables",
    questions: [
      {
        id: "format",
        header: "Format",
        prompt: "Which deliverables should be generated?",
        options: [
          { label: "PPTX", description: "Editable presentation" },
          { label: "MP4", description: "Rendered video" },
        ],
        multiple: true,
        allowCustom: true,
        secret: false,
      },
    ],
  });
  expect(snapshot.capabilities.terminal).toBe(false);
  expect(snapshot.terminals).toEqual([]);
});

it("does not invent configuration controls when the agent declares none", () => {
  const snapshot = projectWebAcpSnapshot({
    agentLabel: "Codex",
    connection: "connected",
    sessionId: "web-acp:pod-1",
    session: {
      ...session,
      configuration: {
        permissionMode: "",
        model: "gpt-5",
        supportedPermissionModes: [],
      },
    },
    title: "Release audit",
  });

  expect(snapshot.capabilities.updateConfiguration).toBe(false);
  expect(snapshot.configuration).toEqual([]);
});

it("preserves an explicit empty permission choice and presents ACP mode labels", () => {
  const snapshot = projectWebAcpSnapshot({
    agentLabel: "Codex",
    connection: "connected",
    sessionId: "web-acp:pod-1",
    session: {
      ...session,
      configuration: {
        permissionMode: "",
        model: "",
        supportedPermissionModes: [
          "bypass",
          "ask_dangerous",
          "ask_any_write",
        ],
      },
    },
    title: "Release audit",
  });

  expect(snapshot.configuration).toEqual([
    {
      id: "permissionMode",
      label: "Permissions",
      value: "",
      options: [
        { value: "bypass", label: "Full access" },
        {
          value: "ask_dangerous",
          label: "Ask for dangerous actions",
        },
        {
          value: "ask_any_write",
          label: "Ask before writes",
        },
      ],
    },
  ]);
});

it("projects generated workspace deliverables from completed file changes", () => {
  const snapshot = projectWebAcpSnapshot({
    agentLabel: "Codex",
    connection: "connected",
    sessionId: "web-acp:pod-1",
    session: {
      ...session,
      toolCalls: {
        files: {
          toolCallId: "files",
          toolName: "fileChange",
          status: "completed",
          success: true,
          argumentsJson: JSON.stringify({
            changes: [
              { path: "deliverables/demo.mp4" },
              { path: "src/player.ts" },
            ],
          }),
          resultText: "Created deliverables/demo.mp4",
          timestamp: 3,
        },
      },
    },
    title: "Generate a demo",
  });

  expect(snapshot.items).toEqual(
    expect.arrayContaining([
      expect.objectContaining({
        kind: "artifact",
        artifactId: "workspace:deliverables/demo.mp4",
        filename: "demo.mp4",
        mimeType: "video/mp4",
      }),
    ]),
  );
});

it("merges discovered Worker artifacts without duplicating tool artifacts", () => {
  const snapshot = projectWebAcpSnapshot({
    agentLabel: "Seedance Expert",
    connection: "connected",
    sessionId: "web-acp:pod-1",
    session: {
      ...session,
      toolCalls: {
        files: {
          toolCallId: "files",
          toolName: "fileChange",
          status: "completed",
          success: true,
          argumentsJson: JSON.stringify({
            changes: [{ path: "output/demo.mp4" }],
          }),
          timestamp: 3,
        },
      },
    },
    title: "Generate a demo",
    workspaceArtifacts: [
      {
        id: "workspace-discovery:artifact:0",
        kind: "artifact",
        artifactId: "workspace:output/demo.mp4",
        filename: "demo.mp4",
        mimeType: "video/mp4",
        status: "completed",
      },
      {
        id: "workspace-discovery:artifact:1",
        kind: "artifact",
        artifactId: "workspace:output/poster.png",
        filename: "poster.png",
        mimeType: "image/png",
        status: "completed",
      },
    ],
  });

  const artifacts = snapshot.items.filter((item) => item.kind === "artifact");
  expect(artifacts).toHaveLength(2);
  expect(artifacts).toEqual(
    expect.arrayContaining([
      expect.objectContaining({ artifactId: "workspace:output/demo.mp4" }),
      expect.objectContaining({ artifactId: "workspace:output/poster.png" }),
    ]),
  );
});
