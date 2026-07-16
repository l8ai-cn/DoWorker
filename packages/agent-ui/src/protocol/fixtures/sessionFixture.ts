import { create } from "@bufbuild/protobuf";

import {
  CommandEnvelopeSchema,
  type CommandEnvelope,
} from "@do-worker/proto/agent_workbench/v2/command_pb";
import {
  SessionSnapshotSchema,
  type SessionSnapshot,
} from "@do-worker/proto/agent_workbench/v2/session_pb";
import { SessionStatus } from "@do-worker/proto/agent_workbench/v2/session_state_pb";
import { createFixtureHistory } from "./sessionFixtureTimeline";
import {
  createFixtureArtifacts,
  createFixtureCommandReceipts,
  createFixturePermissionRequests,
  createFixtureResources,
} from "./sessionFixtureState";

export interface LosslessSessionFixture {
  snapshot: SessionSnapshot;
  terminalCommand: CommandEnvelope;
}

function createTerminalCommand(): CommandEnvelope {
  return create(CommandEnvelopeSchema, {
    sessionId: "session-lossless-1",
    streamEpoch: "epoch-lossless-1",
    commandId: "terminal-command-1",
    payloadDigest: "sha256:terminal-command",
    expectedRevision: 9_007_199_254_740_992n,
    issuedAt: "2026-07-16T00:00:02Z",
    command: {
      case: "terminalOperation",
      value: {
        resourceId: "terminal-1",
        leaseId: "terminal-lease-1",
        fencingEpoch: 4_294_967_297n,
        operation: {
          case: "input",
          value: { data: new TextEncoder().encode("pnpm test\r") },
        },
      },
    },
  });
}

export function createLosslessSessionFixture(): LosslessSessionFixture {
  return {
    snapshot: create(SessionSnapshotSchema, {
      sessionId: "session-lossless-1",
      streamEpoch: "epoch-lossless-1",
      revision: 9_007_199_254_740_993n,
      latestSequence: 9_007_199_254_740_995n,
      activeTurnId: "turn-1",
      history: createFixtureHistory(),
      commandReceipts: createFixtureCommandReceipts(),
      permissionRequests: createFixturePermissionRequests(),
      resources: createFixtureResources(),
      artifacts: createFixtureArtifacts(),
      status: SessionStatus.FAILED,
      digest: "sha256:session-lossless-1",
      error: {
        code: "fixture_runner_failed",
        message: "Fixture runner stopped after producing results.",
        retryable: true,
      },
    }),
    terminalCommand: createTerminalCommand(),
  };
}
