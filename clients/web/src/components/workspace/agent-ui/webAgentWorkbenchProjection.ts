import { fromBinary } from "@bufbuild/protobuf";

import {
  projectGeneratedSessionSnapshot,
  type AgentConnectionStatus,
  type AgentSessionSnapshot,
} from "@agent-cloud/agent-ui";
import {
  SessionSnapshotSchema,
  type SessionSnapshot,
} from "@proto/agent_workbench/v2/session_pb";

export interface WebAgentWorkbenchProjectionContext {
  agentLabel: string;
  interactionMode: "acp" | "pty";
  sessionId: string;
  title: string;
}

export function decodeWebAgentWorkbenchSnapshot(
  bytes: Uint8Array | undefined,
  sessionId: string,
): SessionSnapshot | null {
  if (!bytes?.length) return null;
  const snapshot = fromBinary(SessionSnapshotSchema, bytes);
  if (snapshot.sessionId !== sessionId) {
    throw new Error("agent_workbench_state_session_mismatch");
  }
  return snapshot;
}

export function projectWebAgentWorkbenchSnapshot(
  raw: SessionSnapshot | null,
  context: WebAgentWorkbenchProjectionContext,
  connection: AgentConnectionStatus,
  transportError: string | null,
): AgentSessionSnapshot {
  if (!raw) return emptySnapshot(context, connection, transportError);
  const projected = projectGeneratedSessionSnapshot(raw, {
    agentLabel: context.agentLabel,
    connection,
    hasOlderItems: raw.capabilities?.history ?? false,
    interactionMode: context.interactionMode,
    title: context.title,
  });
  return transportError && !projected.error
    ? { ...projected, error: transportError }
    : projected;
}

function emptySnapshot(
  context: WebAgentWorkbenchProjectionContext,
  connection: AgentConnectionStatus,
  error: string | null,
): AgentSessionSnapshot {
  return {
    agentLabel: context.agentLabel,
    capabilities: {
      interrupt: false,
      resolvePermission: false,
      sendMessage: false,
      terminal: false,
      updateConfiguration: false,
    },
    connection,
    error,
    hasOlderItems: false,
    interactionMode: context.interactionMode,
    items: [],
    permissions: [],
    plan: [],
    sessionId: context.sessionId,
    status: "launching",
    terminals: [],
    title: context.title,
  };
}
