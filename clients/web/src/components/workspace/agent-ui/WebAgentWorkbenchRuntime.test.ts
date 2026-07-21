import { create, fromBinary, toBinary } from "@bufbuild/protobuf";
import { describe, expect, it, vi } from "vitest";

import {
  ArtifactDescriptorSchema,
  ArtifactManifestSchema,
  ArtifactRepresentationSchema,
  VideoManifestSchema,
  VideoStage,
} from "@proto/agent_workbench/v2/artifact_pb";
import { CommandEnvelopeSchema } from "@proto/agent_workbench/v2/command_pb";
import { SessionSnapshotSchema } from "@proto/agent_workbench/v2/session_pb";
import { SessionStatus } from "@proto/agent_workbench/v2/session_state_pb";
import { WebAgentWorkbenchRuntime } from "./WebAgentWorkbenchRuntime";
import type {
  WebAgentWorkbenchRuntimeDeps,
  WebAgentWorkbenchStream,
} from "./webAgentWorkbenchRuntimeTypes";

function snapshot(revision = 4n) {
  return create(SessionSnapshotSchema, {
    sessionId: "session-real-1",
    streamEpoch: "epoch-1",
    revision,
    latestSequence: revision,
    status: SessionStatus.IDLE,
    artifacts: [
      create(ArtifactDescriptorSchema, {
        artifactId: "video-1",
        filename: "demo.mp4",
        mediaType: "video/mp4",
        revision: 1n,
        representations: [
          create(ArtifactRepresentationSchema, {
            representationId: "playable",
            mediaType: "video/mp4",
            revision: 1n,
          }),
        ],
        manifest: create(ArtifactManifestSchema, {
          manifest: {
            case: "video",
            value: create(VideoManifestSchema, {
              playableRepresentationId: "playable",
              stage: VideoStage.READY,
            }),
          },
        }),
      }),
    ],
  });
}

function dependencies(
  sleep: WebAgentWorkbenchRuntimeDeps["sleep"] = vi.fn(async () => undefined),
) {
  let current = snapshot();
  let closeListener: ((detail: unknown) => void) | null = null;
  const stream: WebAgentWorkbenchStream = {
    close: vi.fn(),
    status: vi.fn(() => "open"),
    terminalError: vi.fn(() => undefined),
  };
  const service = {
    getSessionSnapshotConnect: vi.fn(async () =>
      toBinary(SessionSnapshotSchema, current),
    ),
    streamSessionDeltasConnect: vi.fn(async (
      _org: string,
      _token: string,
      _session: string,
      _limit: number,
      _onCommit: () => void,
      _onError: (error: string) => void,
      onClose: (detail: unknown) => void,
    ) => {
      closeListener = onClose;
      return stream;
    }),
    executeCommandConnect: vi.fn(async () => new Uint8Array()),
  };
  const deps: WebAgentWorkbenchRuntimeDeps = {
    getAccess: () => ({ bearerToken: "token-1", orgSlug: "acme" }),
    service,
    sleep,
    state: {
      projectionStatus: vi.fn(() => "ready"),
      resyncReason: vi.fn(() => undefined),
      revision: vi.fn(() => current.revision),
      snapshotBytes: vi.fn(() => toBinary(SessionSnapshotSchema, current)),
    },
  };
  return {
    closeRemote: () => closeListener?.({ status: "remote_closed", error: null }),
    deps,
    service,
    setSnapshot: (next: ReturnType<typeof snapshot>) => {
      current = next;
    },
    stream,
  };
}

describe("WebAgentWorkbenchRuntime", () => {
  it("uses the real session id, explicit access scope, and Rust snapshot projection", async () => {
    const { deps, service } = dependencies();
    const runtime = new WebAgentWorkbenchRuntime({
      agentLabel: "Codex",
      deps,
      interactionMode: "acp",
      sessionId: "session-real-1",
      title: "Video task",
    });

    expect(runtime.sessionId).toBe("session-real-1");
    await runtime.open(runtime.sessionId);
    await runtime.sendMessage(runtime.sessionId, "command-1", {
      text: "Render the video",
    });

    expect(service.getSessionSnapshotConnect).toHaveBeenCalledWith(
      "acme",
      "token-1",
      "session-real-1",
    );
    const commandBytes = service.executeCommandConnect.mock.calls[0]?.[2];
    const command = fromBinary(CommandEnvelopeSchema, commandBytes);
    expect(command).toMatchObject({
      commandId: "command-1",
      expectedRevision: 4n,
      sessionId: "session-real-1",
      streamEpoch: "epoch-1",
    });
    expect(runtime.getSnapshot(runtime.sessionId).items).toEqual(
      expect.arrayContaining([
        expect.objectContaining({
          artifactId: "video-1",
          kind: "artifact",
          manifest: expect.objectContaining({ kind: "video" }),
        }),
      ]),
    );
  });

  it.each([
    ["image", "brief.png", "image/png"],
    ["CSV", "sales.csv", "text/csv"],
    [
      "Word document",
      "brief.docx",
      "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
    ],
  ])("preserves a %s attachment in the command protocol", async (
    _kind,
    name,
    mediaType,
  ) => {
    const { deps, service } = dependencies();
    const runtime = new WebAgentWorkbenchRuntime({
      agentLabel: "Codex",
      deps,
      interactionMode: "acp",
      sessionId: "session-real-1",
      title: "Attachment task",
    });

    await runtime.sendMessage(runtime.sessionId, "command-attachment", {
      attachments: [{ bytes: 4, id: "file_1", mediaType, name }],
      text: "Review the attachment",
    });

    const command = fromBinary(
      CommandEnvelopeSchema,
      service.executeCommandConnect.mock.calls[0]?.[2],
    );
    expect(command.command).toMatchObject({
      case: "sendPrompt",
      value: {
        attachments: [
          {
            content: {
              case: "file",
              value: {
                artifactId: "file_1",
                filename: name,
                mediaType,
                representationId: "source",
                revision: 1n,
                role: "input",
              },
            },
            contentId: "file_1",
            identity: {
              namespace: "agentcloud.session-file",
              schemaVersion: "1",
              semanticKey: "attachment",
            },
          },
        ],
        text: "Review the attachment",
      },
    });
  });

  it("binds the current session and access scope when uploading an attachment", async () => {
    const { deps } = dependencies();
    const uploadAttachment = vi.fn(async () => ({
      bytes: 4,
      id: "file_1",
      mediaType: "text/csv",
      name: "sales.csv",
    }));
    deps.uploadAttachment = uploadAttachment;
    const runtime = new WebAgentWorkbenchRuntime({
      agentLabel: "Codex",
      deps,
      interactionMode: "acp",
      sessionId: "session-real-1",
      title: "Attachment task",
    });
    const file = new File(["data"], "sales.csv", { type: "text/csv" });

    await expect(runtime.uploadAttachment?.(runtime.sessionId, file)).resolves.toEqual({
      bytes: 4,
      id: "file_1",
      mediaType: "text/csv",
      name: "sales.csv",
    });
    expect(uploadAttachment).toHaveBeenCalledWith({
      access: { bearerToken: "token-1", orgSlug: "acme" },
      file,
      sessionId: "session-real-1",
    });
  });

  it("takes a fresh Rust snapshot after a remote stream close", async () => {
    const { closeRemote, deps, service, setSnapshot } = dependencies();
    const runtime = new WebAgentWorkbenchRuntime({
      agentLabel: "Codex",
      deps,
      interactionMode: "acp",
      sessionId: "session-real-1",
      title: "Reconnect task",
    });
    await runtime.open(runtime.sessionId);
    setSnapshot(snapshot(5n));

    closeRemote();

    await vi.waitFor(() => {
      expect(service.getSessionSnapshotConnect).toHaveBeenCalledTimes(2);
    });
    expect(runtime.getSnapshot(runtime.sessionId).connection).toBe("connected");
  });

  it("starts a new reconnect loop after close and reopen", async () => {
    let resumeFirstReconnect: (() => void) | undefined;
    const firstReconnect = new Promise<void>((resolve) => {
      resumeFirstReconnect = resolve;
    });
    const sleep = vi
      .fn<WebAgentWorkbenchRuntimeDeps["sleep"]>()
      .mockImplementationOnce(() => firstReconnect)
      .mockResolvedValue(undefined);
    const { closeRemote, deps, service } = dependencies(sleep);
    const runtime = new WebAgentWorkbenchRuntime({
      agentLabel: "Codex",
      deps,
      interactionMode: "acp",
      sessionId: "session-real-1",
      title: "Generation-safe reconnect",
    });

    await runtime.open(runtime.sessionId);
    closeRemote();
    await vi.waitFor(() => expect(sleep).toHaveBeenCalledTimes(1));

    runtime.close(runtime.sessionId);
    await runtime.open(runtime.sessionId);
    closeRemote();
    resumeFirstReconnect?.();

    await vi.waitFor(() => {
      expect(service.getSessionSnapshotConnect).toHaveBeenCalledTimes(3);
    });
  });

  it("opens a read-only completed session without a live delta stream", async () => {
    const { deps, service } = dependencies();
    const runtime = new WebAgentWorkbenchRuntime({
      agentLabel: "Codex",
      deps,
      interactionMode: "acp",
      live: false,
      sessionId: "session-real-1",
      title: "Completed task",
    });

    await runtime.open(runtime.sessionId);

    expect(service.getSessionSnapshotConnect).toHaveBeenCalledTimes(1);
    expect(service.streamSessionDeltasConnect).not.toHaveBeenCalled();
    expect(runtime.getSnapshot(runtime.sessionId).connection).toBe("connected");
    expect(runtime.getSnapshot(runtime.sessionId).items).toEqual(
      expect.arrayContaining([
        expect.objectContaining({ artifactId: "video-1" }),
      ]),
    );
  });
});
