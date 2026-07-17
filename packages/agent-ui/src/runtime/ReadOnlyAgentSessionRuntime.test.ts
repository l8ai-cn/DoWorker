import { describe, expect, it, vi } from "vitest";

import {
  agentWorkspaceRuntime,
  agentWorkspaceSnapshot,
} from "../AgentWorkspace.test-fixture";
import type { AgentSessionRuntime } from "../contracts";
import { ReadOnlyAgentSessionRuntime } from "./ReadOnlyAgentSessionRuntime";

describe("ReadOnlyAgentSessionRuntime", () => {
  it("preserves browsing while rejecting every mutating command", async () => {
    const snapshot = agentWorkspaceSnapshot();
    const { agentRuntime } = agentWorkspaceRuntime(snapshot);
    const runtime = new ReadOnlyAgentSessionRuntime(agentRuntime);

    expect(runtime.getSnapshot(snapshot.sessionId)).toMatchObject({
      capabilities: {
        interrupt: false,
        resolvePermission: false,
        sendMessage: false,
        terminal: false,
        updateConfiguration: false,
      },
      terminals: [{ writable: false }],
    });
    await expect(runtime.loadArtifact?.(
      snapshot.sessionId,
      "artifact-1",
    )).resolves.toBeInstanceOf(Blob);
    await expect(
      runtime.sendMessage(snapshot.sessionId, "command-1", { text: "mutate" }),
    ).rejects.toThrow("agent_session_read_only");
    await expect(
      runtime.interrupt(snapshot.sessionId, "command-2"),
    ).rejects.toThrow("agent_session_read_only");
    await expect(
      runtime.resolvePermission(
        snapshot.sessionId,
        "command-3",
        "permission-1",
        { action: "decline" },
      ),
    ).rejects.toThrow("agent_session_read_only");
    await expect(
      runtime.updateConfiguration(
        snapshot.sessionId,
        "command-4",
        { model: "gpt-5.6" },
      ),
    ).rejects.toThrow("agent_session_read_only");
    expect((runtime as AgentSessionRuntime).executeArtifactAction).toBeUndefined();
    expect(agentRuntime.sendMessage).not.toHaveBeenCalled();
    expect(agentRuntime.interrupt).not.toHaveBeenCalled();
    expect(agentRuntime.resolvePermission).not.toHaveBeenCalled();
    expect(agentRuntime.updateConfiguration).not.toHaveBeenCalled();
  });

  it("keeps snapshot identity stable until the source changes", () => {
    const snapshot = agentWorkspaceSnapshot();
    const { agentRuntime } = agentWorkspaceRuntime(snapshot);
    const runtime = new ReadOnlyAgentSessionRuntime(agentRuntime);

    expect(runtime.getSnapshot(snapshot.sessionId)).toBe(
      runtime.getSnapshot(snapshot.sessionId),
    );
  });
});
