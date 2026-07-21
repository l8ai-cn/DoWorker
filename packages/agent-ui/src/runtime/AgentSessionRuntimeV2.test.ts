import { create } from "@bufbuild/protobuf";
import { describe, expect, it, vi } from "vitest";

import {
  CommandReceiptSchema,
  CommandReceiptState,
  type CommandEnvelope,
} from "@agent-cloud/proto/agent_workbench/v2/command_pb";
import { SessionSnapshotSchema } from "@agent-cloud/proto/agent_workbench/v2/session_pb";
import { SessionStatus } from "@agent-cloud/proto/agent_workbench/v2/session_state_pb";
import type { AgentWorkbenchSessionTransport } from "./AgentWorkbenchConnectTransport";
import { AgentSessionConnection } from "./AgentSessionConnection";
import { AgentSessionRuntimeV2 } from "./AgentSessionRuntimeV2";

describe("AgentSessionRuntimeV2", () => {
  it("projects connection state and sends protocol commands without optimistic state", async () => {
    const commands: CommandEnvelope[] = [];
    const transport = fakeTransport(commands);
    const connection = new AgentSessionConnection(transport);
    const runtime = new AgentSessionRuntimeV2({
      agentLabel: "OpenAI Codex",
      connection,
      interactionMode: "acp",
      now: () => "2026-07-16T10:00:00Z",
      sessionId: "session-1",
      title: "真实任务",
    });

    expect(runtime.getSnapshot("session-1")).toMatchObject({
      connection: "connecting",
      items: [],
      status: "launching",
    });

    await runtime.open("session-1");
    await runtime.sendMessage("session-1", "command-send", { text: "生成视频" });
    await runtime.interrupt("session-1", "command-stop");
    await runtime.updateConfiguration("session-1", "command-config", {
      model: "gpt-5",
      permission_mode: "default",
    });
    await runtime.executeArtifactAction("session-1", {
      actionSchemaVersion: "1",
      actionType: "image.edit",
      artifactId: "image-1",
      baseRevision: 3n,
      commandId: "command-image",
      payload: { instruction: "移除背景" },
      representationId: "source",
    });

    expect(commands.map((command) => command.command.case)).toEqual([
      "sendPrompt",
      "interrupt",
      "changeConfiguration",
      "artifactAction",
    ]);
    expect(commands[0]).toMatchObject({
      expectedRevision: 1n,
      issuedAt: "2026-07-16T10:00:00Z",
      command: { case: "sendPrompt", value: { text: "生成视频" } },
    });
    expect(commands[2]?.command.value).toMatchObject({
      values: [
        {
          key: "model",
          value: { mediaType: "application/json" },
        },
        {
          key: "permission_mode",
          value: { mediaType: "application/json" },
        },
      ],
    });
    expect(commands[3]?.command.value).toMatchObject({
      actionSchemaVersion: "1",
      actionType: "image.edit",
      artifactId: "image-1",
      baseRevision: 3n,
      clientActionId: "command-image",
      representationId: "source",
    });
    expect(runtime.getSnapshot("session-1").items).toEqual([]);
    runtime.close("session-1");
  });

  it("delegates artifact bytes to the host resource adapter", async () => {
    const loadArtifact = vi.fn(async () => new Blob(["video"]));
    const runtime = new AgentSessionRuntimeV2({
      agentLabel: "Agent",
      connection: new AgentSessionConnection(fakeTransport([])),
      interactionMode: "acp",
      loadArtifact,
      sessionId: "session-1",
      title: "Task",
    });

    const blob = await runtime.loadArtifact?.(
      "session-1",
      "video-1",
      "playable",
    );

    expect(await blob?.text()).toBe("video");
    expect(loadArtifact).toHaveBeenCalledWith({
      artifactId: "video-1",
      representationId: "playable",
      sessionId: "session-1",
    });
  });
});

function fakeTransport(
  commands: CommandEnvelope[],
): AgentWorkbenchSessionTransport {
  return {
    async execute(command) {
      commands.push(command);
      return create(CommandReceiptSchema, {
        sessionId: command.sessionId,
        commandId: command.commandId,
        payloadDigest: command.payloadDigest,
        state: CommandReceiptState.ACCEPTED,
      });
    },
    async getSnapshot() {
      return create(SessionSnapshotSchema, {
        sessionId: "session-1",
        streamEpoch: "epoch-1",
        revision: 1n,
        latestSequence: 1n,
        status: SessionStatus.IDLE,
      });
    },
    streamDeltas(_cursor, signal) {
      return waitForAbort(signal);
    },
  };
}

async function* waitForAbort(signal: AbortSignal) {
  await new Promise<void>((resolve) => {
    if (signal.aborted) resolve();
    else signal.addEventListener("abort", () => resolve(), { once: true });
  });
}
