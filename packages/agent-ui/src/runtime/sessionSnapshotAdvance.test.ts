import { create } from "@bufbuild/protobuf";
import { describe, expect, it } from "vitest";

import {
  ArtifactDescriptorSchema,
  ArtifactGrantSchema,
  ArtifactStatus,
} from "@do-worker/proto/agent_workbench/v2/artifact_pb";
import {
  AuthorizationGrantSchema,
  SessionSnapshotSchema,
} from "@do-worker/proto/agent_workbench/v2/session_pb";
import { SessionStatus } from "@do-worker/proto/agent_workbench/v2/session_state_pb";
import { validateSnapshotAdvance } from "./sessionSnapshotAdvance";

describe("validateSnapshotAdvance", () => {
  it("accepts refreshed session grants at the same cursor", () => {
    const current = snapshot();
    const next = snapshot();
    next.grants = [sessionGrant()];

    expect(validateSnapshotAdvance(current, next)).toBe("metadata");
  });

  it("accepts refreshed artifact grants at the same cursor", () => {
    const current = snapshot();
    current.artifacts = [artifact()];
    const next = snapshot();
    next.artifacts = [artifact()];
    next.artifacts[0].grants = [artifactGrant()];

    expect(validateSnapshotAdvance(current, next)).toBe("metadata");
  });

  it("rejects canonical content changes at the same cursor", () => {
    const current = snapshot();
    const next = snapshot();
    next.status = SessionStatus.RUNNING;

    expect(() => validateSnapshotAdvance(current, next)).toThrow(
      "snapshot_cursor_conflict",
    );
  });
});

function snapshot() {
  return create(SessionSnapshotSchema, {
    sessionId: "session-1",
    streamEpoch: "epoch-1",
    revision: 1n,
    latestSequence: 1n,
    status: SessionStatus.IDLE,
  });
}

function sessionGrant() {
  return create(AuthorizationGrantSchema, {
    grantId: "grant-1",
    issuer: "backend",
    subject: "user-1",
    sessionId: "session-1",
    actions: ["agent.prompt.send"],
    issuedAt: "2026-07-16T00:00:00Z",
  });
}

function artifact() {
  return create(ArtifactDescriptorSchema, {
    artifactId: "artifact-1",
    revision: 1n,
    filename: "result.txt",
    mediaType: "text/plain",
    status: ArtifactStatus.READY,
  });
}

function artifactGrant() {
  return create(ArtifactGrantSchema, {
    grantId: "artifact-grant-1",
    issuer: "backend",
    subject: "user-1",
    actions: ["artifact.content.read"],
    issuedAt: "2026-07-16T00:00:00Z",
  });
}
