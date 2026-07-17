import { vi } from "vitest";

import type {
  AgentSessionRuntime,
  AgentSessionSnapshot,
  TerminalRuntime,
} from "./contracts";

export function agentWorkspaceSnapshot(): AgentSessionSnapshot {
  return {
    sessionId: "session-1",
    title: "Release audit",
    agentLabel: "Codex",
    status: "running",
    connection: "connected",
    interactionMode: "acp",
    capabilities: {
      sendMessage: true,
      interrupt: true,
      resolvePermission: true,
      updateConfiguration: true,
      terminal: true,
    },
    commands: [
      {
        name: "compact",
        label: "/compact",
        description: "Compact the current context",
      },
    ],
    configuration: [
      {
        id: "permissionMode",
        label: "Permissions",
        value: "default",
        options: [
          { value: "default", label: "Ask before changes" },
          { value: "acceptEdits", label: "Accept edits" },
        ],
      },
      {
        id: "model",
        label: "Model",
        value: "gpt-5.6",
        options: [
          { value: "gpt-5.6", label: "GPT-5.6" },
          { value: "gpt-5.5", label: "GPT-5.5" },
        ],
      },
    ],
    metadata: [
      { id: "runner", label: "Runner", value: "dev-runner-codex" },
    ],
    items: [
      {
        id: "user-1",
        kind: "message",
        role: "user",
        text: "Audit the release.",
        status: "completed",
      },
      {
        id: "tool-1",
        identity: {
          namespace: "agentsmesh.acp",
          schemaVersion: "1",
          semanticKey: "shell",
        },
        kind: "tool",
        results: [],
        title: "shell",
        input: "pnpm test",
        output: "12 tests passed",
        status: "running",
      },
    ],
    plan: [
      { id: "plan-1", title: "Inspect changes", status: "completed" },
      { id: "plan-2", title: "Run verification", status: "running" },
    ],
    permissions: [
      {
        id: "permission-1",
        kind: "approval",
        title: "Run release command",
        description: "pnpm release",
      },
    ],
    terminals: [
      {
        id: "terminal-1",
        label: "Agent terminal",
        status: "connected",
        writable: true,
      },
    ],
    hasOlderItems: false,
    error: null,
  };
}

export function agentWorkspaceRuntime(snapshot: AgentSessionSnapshot) {
  const listeners = new Set<() => void>();
  const agentRuntime: AgentSessionRuntime = {
    open: vi.fn(async () => undefined),
    close: vi.fn(),
    getSnapshot: () => snapshot,
    subscribe: (_sessionId, listener) => {
      listeners.add(listener);
      return () => listeners.delete(listener);
    },
    sendMessage: vi.fn(async () => undefined),
    sendSlashCommand: vi.fn(async () => undefined),
    interrupt: vi.fn(async () => undefined),
    resolvePermission: vi.fn(async () => undefined),
    updateConfiguration: vi.fn(async () => undefined),
    loadArtifact: vi.fn(
      async () => new Blob(["artifact"], { type: "image/png" }),
    ),
    loadOlder: vi.fn(async () => undefined),
  };
  const terminalRuntime: TerminalRuntime = {
    connect: vi.fn(async () => undefined),
    disconnect: vi.fn(),
    subscribeOutput: vi.fn(() => () => undefined),
    subscribeStatus: vi.fn(() => () => undefined),
    write: vi.fn(async () => undefined),
    resize: vi.fn(async () => undefined),
    acquireControl: vi.fn(async () => ({
      leaseId: "lease-1",
      expiresAt: Date.now() + 60_000,
    })),
    renewControl: vi.fn(async () => undefined),
    releaseControl: vi.fn(async () => undefined),
  };
  return { agentRuntime, terminalRuntime };
}
