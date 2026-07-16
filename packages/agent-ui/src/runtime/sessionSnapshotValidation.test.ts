import { create } from "@bufbuild/protobuf";
import { describe, expect, it } from "vitest";

import {
  ArtifactStatus,
  ArtifactDescriptorSchema,
} from "@do-worker/proto/agent_workbench/v2/artifact_pb";
import {
  CommandReceiptSchema,
  CommandReceiptState,
  PermissionDecision,
} from "@do-worker/proto/agent_workbench/v2/command_pb";
import {
  AuthorizationGrantSchema,
  PermissionRequestSchema,
  SessionResourceSchema,
  SessionSnapshotSchema,
  TimelineItemContentSchema,
  TimelineItemSchema,
} from "@do-worker/proto/agent_workbench/v2/session_pb";
import {
  PermissionRequestState,
  SessionResourceStatus,
  SessionStatus,
} from "@do-worker/proto/agent_workbench/v2/session_state_pb";
import {
  AgentSessionReductionError,
  applySessionSnapshot,
} from "./agentSessionReducer";

describe("session snapshot validation", () => {
  it.each([
    ["revision above sequence", { revision: 2n, latestSequence: 1n }],
    ["zero revision with nonzero sequence", { revision: 0n, latestSequence: 1n }],
    ["nonzero revision with zero sequence", { revision: 1n, latestSequence: 0n }],
  ])("rejects %s", (_name, cursor) => {
    expect(() =>
      applySessionSnapshot(create(SessionSnapshotSchema, {
        ...validSnapshot(),
        ...cursor,
      })),
    ).toThrowError(new AgentSessionReductionError("snapshot_position_invalid"));
  });

  it.each([
    ["empty command id", () => ({ commandReceipts: [receipt({ commandId: "" })] })],
    ["empty receipt digest", () => ({ commandReceipts: [receipt({ payloadDigest: "" })] })],
    [
      "duplicate command id",
      () => ({ commandReceipts: [receipt(), receipt()] }),
    ],
    [
      "cross-session grant",
      () => ({ grants: [grant({ sessionId: "session-2" })] }),
    ],
    ["empty permission id", () => ({ permissionRequests: [permission("")] })],
    [
      "duplicate permission id",
      () => ({ permissionRequests: [permission("permission-1"), permission("permission-1")] }),
    ],
    ["empty resource id", () => ({ resources: [resource("")] })],
    [
      "duplicate resource id",
      () => ({ resources: [resource("resource-1"), resource("resource-1")] }),
    ],
    ["empty artifact id", () => ({ artifacts: [artifact("")] })],
    [
      "duplicate artifact id",
      () => ({ artifacts: [artifact("artifact-1"), artifact("artifact-1")] }),
    ],
  ])("rejects %s", (_name, patch) => {
    expect(() =>
      applySessionSnapshot(create(SessionSnapshotSchema, {
        ...validSnapshot(),
        ...patch(),
      })),
    ).toThrow(AgentSessionReductionError);
  });

  it.each([
    ["history revision above snapshot", { revision: 5n, sequence: 1n }],
    ["history sequence above cursor", { revision: 4n, sequence: 5n }],
  ])("rejects %s", (_name, envelopePatch) => {
    const item = timelineItem("item-1", envelopePatch);

    expect(() =>
      applySessionSnapshot(create(SessionSnapshotSchema, {
        ...validSnapshot(),
        history: [item],
      })),
    ).toThrowError(new AgentSessionReductionError("snapshot_history_invalid"));
  });

  it("rejects duplicate history event sequences", () => {
    expect(() =>
      applySessionSnapshot(create(SessionSnapshotSchema, {
        ...validSnapshot(),
        latestSequence: 4n,
        history: [
          timelineItem("item-1", { sequence: 2n }),
          timelineItem("item-2", { sequence: 2n }),
        ],
      })),
    ).toThrowError(new AgentSessionReductionError("snapshot_history_invalid"));
  });

  it.each([
    [
      "resolved permission without resolution",
      permission("permission-1", PermissionRequestState.RESOLVED),
    ],
    [
      "resolution for a different permission",
      permission("permission-1", PermissionRequestState.RESOLVED, "permission-2"),
    ],
    [
      "pending permission with resolution",
      permission("permission-1", PermissionRequestState.PENDING, "permission-1"),
    ],
  ])("rejects %s", (_name, request) => {
    expect(() =>
      applySessionSnapshot(create(SessionSnapshotSchema, {
        ...validSnapshot(),
        permissionRequests: [request],
      })),
    ).toThrowError(new AgentSessionReductionError("snapshot_permission_invalid"));
  });

  it("rejects a timeline item without a typed content case", () => {
    const item = timelineItem("item-1");
    item.content!.content = { case: undefined };

    expect(() =>
      applySessionSnapshot(create(SessionSnapshotSchema, {
        ...validSnapshot(),
        history: [item],
      })),
    ).toThrowError(new AgentSessionReductionError("snapshot_history_invalid"));
  });

  it("rejects a pending permission without a request payload", () => {
    const request = create(PermissionRequestSchema, {
      permissionRequestId: "permission-1",
      state: PermissionRequestState.PENDING,
    });

    expect(() =>
      applySessionSnapshot(create(SessionSnapshotSchema, {
        ...validSnapshot(),
        permissionRequests: [request],
      })),
    ).toThrowError(new AgentSessionReductionError("snapshot_permission_invalid"));
  });
});

function validSnapshot() {
  return {
    sessionId: "session-1",
    streamEpoch: "epoch-1",
    revision: 4n,
    latestSequence: 4n,
    status: SessionStatus.RUNNING,
  };
}

function receipt(patch: { commandId?: string; payloadDigest?: string } = {}) {
  return create(CommandReceiptSchema, {
    sessionId: "session-1",
    commandId: patch.commandId ?? "command-1",
    payloadDigest: patch.payloadDigest ?? "sha256:command-1",
    state: CommandReceiptState.RUNNING,
  });
}

function grant(patch: { sessionId?: string } = {}) {
  return create(AuthorizationGrantSchema, {
    grantId: "grant-1",
    issuer: "backend",
    subject: "user-1",
    sessionId: patch.sessionId ?? "session-1",
    actions: ["artifact.read"],
    issuedAt: "2026-07-16T00:00:00Z",
  });
}

function permission(
  id: string,
  state = PermissionRequestState.PENDING,
  resolvedPermissionId?: string,
) {
  return create(PermissionRequestSchema, {
    permissionRequestId: id,
    state,
    request:
      state === PermissionRequestState.PENDING
        ? { case: "approval", value: { title: "Approve action" } }
        : { case: undefined },
    resolution: resolvedPermissionId
      ? {
          permissionRequestId: resolvedPermissionId,
          decision: PermissionDecision.ACCEPT,
        }
      : undefined,
  });
}

function resource(id: string) {
  return create(SessionResourceSchema, {
    resourceId: id,
    label: "Workspace",
    status: SessionResourceStatus.READY,
    resource: { case: "environment", value: {} },
  });
}

function artifact(id: string) {
  return create(ArtifactDescriptorSchema, {
    artifactId: id,
    revision: 1n,
    filename: "result.txt",
    mediaType: "text/plain",
    status: ArtifactStatus.READY,
  });
}

function timelineItem(
  itemId: string,
  patch: { revision?: bigint; sequence?: bigint } = {},
) {
  return create(TimelineItemSchema, {
    envelope: {
      sessionId: "session-1",
      streamEpoch: "epoch-1",
      revision: patch.revision ?? 4n,
      sequence: patch.sequence ?? 1n,
      itemId,
      createdAt: "2026-07-16T00:00:00Z",
    },
    content: create(TimelineItemContentSchema, {
      content: { case: "system", value: { content: [] } },
    }),
  });
}
