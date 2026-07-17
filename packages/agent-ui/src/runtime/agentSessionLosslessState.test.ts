import { create } from "@bufbuild/protobuf";
import { describe, expect, it } from "vitest";

import {
  AgentErrorSchema,
  PermissionDecision,
} from "@do-worker/proto/agent_workbench/v2/command_pb";
import {
  AgentEventSchema,
  PermissionRequestSchema,
  PermissionResolutionSchema,
  SessionDeltaBatchSchema,
  SessionSnapshotSchema,
} from "@do-worker/proto/agent_workbench/v2/session_pb";
import {
  PermissionRequestState,
  SessionStatus,
} from "@do-worker/proto/agent_workbench/v2/session_state_pb";
import {
  AgentSessionReductionError,
  applyDeltaBatch,
  applySessionSnapshot,
} from "./agentSessionReducer";

describe("lossless session state reduction", () => {
  it("stores the exact permission resolution in the request", () => {
    const resolution = create(PermissionResolutionSchema, {
      permissionRequestId: "permission-1",
      decision: PermissionDecision.ACCEPT,
      response: {
        mediaType: "application/json",
        data: Uint8Array.from(
          new TextEncoder().encode('{"scope":"once"}'),
        ),
      },
      resolvedAt: "2026-07-16T00:00:02Z",
    });
    const next = applyDeltaBatch(
      applySessionSnapshot(snapshot()),
      batch({
        case: "permissionResolved",
        value: { resolution },
      }),
    );

    resolution.resolvedAt = "mutated";

    expect(next.snapshot.permissionRequests[0]?.state).toBe(
      PermissionRequestState.RESOLVED,
    );
    expect(next.snapshot.permissionRequests[0]?.resolution?.resolvedAt).toBe(
      "2026-07-16T00:00:02Z",
    );
    expect(next.snapshot.permissionRequests[0]?.resolution?.response?.data).toEqual(
      Uint8Array.from(new TextEncoder().encode('{"scope":"once"}')),
    );
  });

  it("stores the exact error attached to a terminal status event", () => {
    const error = create(AgentErrorSchema, {
      code: "runner_failed",
      message: "Runner stopped unexpectedly.",
      retryable: true,
    });
    const next = applyDeltaBatch(
      applySessionSnapshot(snapshot()),
      batch({
        case: "sessionStatusChanged",
        value: { status: SessionStatus.FAILED, error },
      }),
    );

    error.message = "mutated";

    expect(next.snapshot.status).toBe(SessionStatus.FAILED);
    expect(next.snapshot.error?.message).toBe("Runner stopped unexpectedly.");
  });

  it("rejects an invalid permission resolution atomically", () => {
    const state = applySessionSnapshot(snapshot());

    expect(() =>
      applyDeltaBatch(
        state,
        batch({
          case: "permissionResolved",
          value: {
            resolution: create(PermissionResolutionSchema, {
              permissionRequestId: "permission-1",
              decision: PermissionDecision.UNSPECIFIED,
            }),
          },
        }),
      ),
    ).toThrowError(new AgentSessionReductionError("snapshot_permission_invalid"));
    expect(state.snapshot.permissionRequests[0]?.state).toBe(
      PermissionRequestState.PENDING,
    );
  });
});

function snapshot() {
  return create(SessionSnapshotSchema, {
    sessionId: "session-1",
    streamEpoch: "epoch-1",
    revision: 1n,
    latestSequence: 1n,
    status: SessionStatus.RUNNING,
    permissionRequests: [
      create(PermissionRequestSchema, {
        permissionRequestId: "permission-1",
        state: PermissionRequestState.PENDING,
        request: {
          case: "approval",
          value: { title: "Approve action" },
        },
      }),
    ],
  });
}

function batch(event: ReturnType<typeof eventPayload>) {
  return create(SessionDeltaBatchSchema, {
    sessionId: "session-1",
    streamEpoch: "epoch-1",
    baseRevision: 1n,
    revision: 2n,
    firstSequence: 2n,
    lastSequence: 2n,
    digest: `sha256:${event.case}`,
    events: [
      create(AgentEventSchema, {
        envelope: {
          sessionId: "session-1",
          streamEpoch: "epoch-1",
          revision: 2n,
          sequence: 2n,
          itemId: `event-${event.case}`,
          createdAt: "2026-07-16T00:00:02Z",
        },
        event,
      }),
    ],
  });
}

function eventPayload(
  event:
    | {
        case: "permissionResolved";
        value: {
          resolution: ReturnType<typeof create<typeof PermissionResolutionSchema>>;
        };
      }
    | {
        case: "sessionStatusChanged";
        value: {
          status: SessionStatus;
          error: ReturnType<typeof create<typeof AgentErrorSchema>>;
        };
      },
) {
  return event;
}
