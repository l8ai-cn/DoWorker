import { vi } from "vitest";

import { EMPTY_SESSION } from "@/stores/acpSessionTypes";
import { WebAcpSessionRuntime } from "./WebAcpSessionRuntime";

function dependencies() {
  const listeners = new Set<() => void>();
  const relay = {
    subscribe: vi.fn(async () => ({ send: vi.fn(), unsubscribe: vi.fn() })),
    unsubscribe: vi.fn(),
    onAcpMessage: vi.fn(() => vi.fn()),
    onStatusChange: vi.fn((_podKey: string, listener: (value: unknown) => void) => {
      listener({
        status: "connected",
        runnerDisconnected: false,
        controlLease: { status: "observer" },
      });
      return vi.fn();
    }),
    sendAcpCommand: vi.fn(async () => undefined),
  };
  return {
    relay,
    readSession: vi.fn(() => EMPTY_SESSION),
    subscribeSession: vi.fn((listener: () => void) => {
      listeners.add(listener);
      return () => listeners.delete(listener);
    }),
    dispatchRelayEvent: vi.fn(),
    removePermission: vi.fn(),
    listWorkspaceArtifacts: vi.fn(async () => []),
    loadWorkspaceArtifact: vi.fn(async () => new Blob(["artifact"], { type: "image/png" })),
    notifySession: () => listeners.forEach((listener) => listener()),
  };
}

it("owns the Web ACP transport and sends correlated commands", async () => {
  const deps = dependencies();
  const runtime = new WebAcpSessionRuntime({
    agentLabel: "Codex",
    deps,
    paneId: "pane-1",
    podKey: "pod-1",
    title: "Release audit",
  });

  expect(runtime.sessionId).toBe("web-acp:pod-1");
  expect(runtime.getSnapshot(runtime.sessionId)).toBe(
    runtime.getSnapshot(runtime.sessionId),
  );
  await runtime.open(runtime.sessionId);
  await runtime.sendMessage(runtime.sessionId, "command-1", { text: "Run tests" });
  await runtime.sendSlashCommand?.(runtime.sessionId, "command-slash", {
    name: "compact",
    arguments: "",
  });
  await runtime.interrupt(runtime.sessionId, "command-2");
  await runtime.resolvePermission(
    runtime.sessionId,
    "command-3",
    "permission-1",
    {
      action: "accept",
      content: { answers: { format: ["PPTX", "MP4"] } },
    },
  );

  expect(deps.relay.subscribe).toHaveBeenCalledWith(
    "pod-1",
    "agent-workspace-pane-1-1",
    expect.any(Function),
  );
  expect(deps.relay.sendAcpCommand).toHaveBeenNthCalledWith(1, "pod-1", {
    type: "prompt",
    prompt: "Run tests",
    requestId: "command-1",
  });
  expect(deps.relay.sendAcpCommand).toHaveBeenNthCalledWith(2, "pod-1", {
    type: "prompt",
    prompt: "/compact",
    requestId: "command-slash",
  });
  expect(deps.relay.sendAcpCommand).toHaveBeenNthCalledWith(3, "pod-1", {
    type: "interrupt",
    requestId: "command-2",
  });
  expect(deps.relay.sendAcpCommand).toHaveBeenNthCalledWith(4, "pod-1", {
    type: "permission_response",
    requestId: "permission-1",
    approved: true,
    updatedInput: { answers: { format: ["PPTX", "MP4"] } },
  });
  expect(deps.removePermission).toHaveBeenCalledWith("pod-1", "permission-1");
  runtime.close(runtime.sessionId);
  expect(deps.relay.unsubscribe).toHaveBeenCalledWith(
    "pod-1",
    "agent-workspace-pane-1-1",
  );
});

it("projects actionable ACP configuration into the shared workspace", () => {
  const deps = dependencies();
  deps.readSession.mockReturnValue({
    ...EMPTY_SESSION,
    configuration: {
      permissionMode: "default",
      model: "gpt-5.6",
      supportedPermissionModes: ["default", "acceptEdits"],
    },
  });
  const runtime = new WebAcpSessionRuntime({
    agentLabel: "Codex",
    deps,
    paneId: "pane-1",
    podKey: "pod-1",
    title: "Release audit",
  });

  expect(runtime.getSnapshot(runtime.sessionId)).toMatchObject({
    commands: expect.arrayContaining([
      expect.objectContaining({ name: "compact", label: "/compact" }),
    ]),
    configuration: [
      expect.objectContaining({
        id: "permissionMode",
        value: "default",
        options: [
          expect.objectContaining({ value: "default" }),
          expect.objectContaining({ value: "acceptEdits" }),
        ],
      }),
      expect.objectContaining({
        id: "model",
        value: "gpt-5.6",
      }),
    ],
  });
});
it("rejects configuration commands when the agent declares no control capability", async () => {
  const deps = dependencies();
  const runtime = new WebAcpSessionRuntime({
    agentLabel: "Codex",
    deps,
    paneId: "pane-1",
    podKey: "pod-1",
    title: "Release audit",
  });

  await expect(
    runtime.updateConfiguration(runtime.sessionId, "command-config", {
      permissionMode: "default",
    }),
  ).rejects.toThrow("configuration control");
  expect(deps.relay.sendAcpCommand).not.toHaveBeenCalled();
});
it("reports a Relay subscription failure through the snapshot without rejecting open", async () => {
  const deps = dependencies();
  deps.relay.subscribe.mockRejectedValueOnce(new Error("Relay unavailable"));
  const runtime = new WebAcpSessionRuntime({
    agentLabel: "Codex",
    deps,
    paneId: "pane-1",
    podKey: "pod-1",
    title: "Release audit",
  });

  await expect(runtime.open(runtime.sessionId)).resolves.toBeUndefined();

  expect(runtime.getSnapshot(runtime.sessionId)).toMatchObject({
    connection: "disconnected",
    error: "Relay unavailable",
  });
});
it("starts a fresh Relay subscription after StrictMode closes a pending open", async () => {
  const deps = dependencies();
  let resolveFirst = () => undefined;
  deps.relay.subscribe
    .mockImplementationOnce(
      () => new Promise((resolve) => {
        resolveFirst = () => resolve(undefined);
      }),
    )
    .mockResolvedValueOnce(undefined);
  const runtime = new WebAcpSessionRuntime({
    agentLabel: "Codex",
    deps,
    paneId: "pane-1",
    podKey: "pod-1",
    title: "Release audit",
  });

  const first = runtime.open(runtime.sessionId);
  runtime.close(runtime.sessionId);
  const second = runtime.open(runtime.sessionId);
  resolveFirst();
  await Promise.all([first, second]);

  expect(deps.relay.subscribe).toHaveBeenNthCalledWith(
    1,
    "pod-1",
    "agent-workspace-pane-1-1",
    expect.any(Function),
  );
  expect(deps.relay.subscribe).toHaveBeenNthCalledWith(
    2,
    "pod-1",
    "agent-workspace-pane-1-2",
    expect.any(Function),
  );
  expect(deps.subscribeSession).toHaveBeenCalledTimes(2);
});
it("reports a disconnected snapshot after the runtime closes", async () => {
  const deps = dependencies();
  const runtime = new WebAcpSessionRuntime({
    agentLabel: "Codex",
    deps,
    paneId: "pane-1",
    podKey: "pod-1",
    title: "Release audit",
  });

  await runtime.open(runtime.sessionId);
  expect(runtime.getSnapshot(runtime.sessionId).connection).toBe("connected");

  runtime.close(runtime.sessionId);

  expect(runtime.getSnapshot(runtime.sessionId).connection).toBe(
    "disconnected",
  );
});
it("surfaces an immediate Relay command failure in the runtime snapshot", async () => {
  const deps = dependencies();
  deps.relay.sendAcpCommand.mockRejectedValueOnce(new Error("Relay write failed"));
  const runtime = new WebAcpSessionRuntime({
    agentLabel: "Codex",
    deps,
    paneId: "pane-1",
    podKey: "pod-1",
    title: "Release audit",
  });

  await expect(
    runtime.sendMessage(runtime.sessionId, "command-1", { text: "Run tests" }),
  ).rejects.toThrow("Relay write failed");

  expect(runtime.getSnapshot(runtime.sessionId).error).toBe("Relay write failed");
});

it("loads a projected workspace artifact through the authenticated Web adapter", async () => {
  const deps = dependencies();
  const runtime = new WebAcpSessionRuntime({
    agentLabel: "Codex",
    deps,
    paneId: "pane-1",
    podKey: "pod-1",
    title: "Release audit",
  });

  const blob = await runtime.loadArtifact?.(
    runtime.sessionId,
    "workspace:deliverables/preview.png",
  );

  expect(blob?.type).toBe("image/png");
  expect(deps.loadWorkspaceArtifact).toHaveBeenCalledWith(
    "pod-1",
    "deliverables/preview.png",
  );
});

it("discovers existing Worker artifacts when the ACP workspace opens", async () => {
  const deps = dependencies();
  deps.listWorkspaceArtifacts.mockResolvedValueOnce([
    {
      id: "workspace-discovery:artifact:0",
      kind: "artifact",
      artifactId: "workspace:output/demo.mp4",
      filename: "demo.mp4",
      mimeType: "video/mp4",
      status: "completed",
    },
  ]);
  const runtime = new WebAcpSessionRuntime({
    agentLabel: "Seedance Expert",
    deps,
    paneId: "pane-1",
    podKey: "pod-1",
    title: "Video creation",
  });

  await runtime.open(runtime.sessionId);

  expect(deps.listWorkspaceArtifacts).toHaveBeenCalledWith("pod-1");
  expect(runtime.getSnapshot(runtime.sessionId).items).toEqual(
    expect.arrayContaining([
      expect.objectContaining({
        kind: "artifact",
        artifactId: "workspace:output/demo.mp4",
      }),
    ]),
  );
});

it("discovers completed Worker artifacts without opening Relay", async () => {
  const deps = dependencies();
  deps.listWorkspaceArtifacts.mockResolvedValueOnce([
    {
      id: "workspace-discovery:artifact:0",
      kind: "artifact",
      artifactId: "workspace:deliverables/video/pattern-preview.mp4",
      filename: "pattern-preview.mp4",
      mimeType: "video/mp4",
      status: "completed",
    },
  ]);
  const runtime = new WebAcpSessionRuntime({
    agentLabel: "Pattern Designer",
    deps,
    live: false,
    paneId: "pane-1",
    podKey: "pod-1",
    title: "Pattern preview",
  });

  await runtime.open(runtime.sessionId);

  expect(runtime.getSnapshot(runtime.sessionId)).toMatchObject({
    connection: "disconnected",
    error: null,
  });
  expect(deps.relay.subscribe).not.toHaveBeenCalled();
  expect(runtime.getSnapshot(runtime.sessionId).items).toEqual(
    expect.arrayContaining([
      expect.objectContaining({
        artifactId: "workspace:deliverables/video/pattern-preview.mp4",
      }),
    ]),
  );
});

it("refreshes Worker artifacts after an ACP turn becomes idle", async () => {
  const deps = dependencies();
  deps.readSession.mockReturnValue({
    ...EMPTY_SESSION,
    state: "processing",
  });
  const runtime = new WebAcpSessionRuntime({
    agentLabel: "Seedance Expert",
    deps,
    paneId: "pane-1",
    podKey: "pod-1",
    title: "Video creation",
  });
  await runtime.open(runtime.sessionId);
  expect(deps.listWorkspaceArtifacts).toHaveBeenCalledTimes(1);

  deps.readSession.mockReturnValue({
    ...EMPTY_SESSION,
    state: "idle",
  });
  deps.notifySession();

  await vi.waitFor(() => {
    expect(deps.listWorkspaceArtifacts).toHaveBeenCalledTimes(2);
  });
});

it("shares one Relay event consumer across two panels for the same Pod", async () => {
  const deps = dependencies();
  const first = new WebAcpSessionRuntime({
    agentLabel: "Codex",
    deps,
    paneId: "pane-1",
    podKey: "pod-1",
    title: "Release audit",
  });
  const second = new WebAcpSessionRuntime({
    agentLabel: "Codex",
    deps,
    paneId: "pane-2",
    podKey: "pod-1",
    title: "Release audit",
  });

  await Promise.all([
    first.open(first.sessionId),
    second.open(second.sessionId),
  ]);

  expect(deps.relay.subscribe).toHaveBeenCalledTimes(1);
  expect(deps.relay.onAcpMessage).toHaveBeenCalledTimes(1);
  const relayListener = deps.relay.onAcpMessage.mock.calls[0]?.[1];
  relayListener?.(9, { type: "contentChunk", text: "once" });
  expect(deps.dispatchRelayEvent).toHaveBeenCalledTimes(1);

  first.close(first.sessionId);
  expect(deps.relay.unsubscribe).not.toHaveBeenCalled();
  second.close(second.sessionId);
  expect(deps.relay.unsubscribe).toHaveBeenCalledTimes(1);
});
