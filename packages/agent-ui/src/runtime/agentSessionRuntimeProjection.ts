import type {
  AgentConnectionStatus,
  AgentSessionSnapshot,
} from "../contracts";
import type {
  AgentSessionConnection,
  AgentSessionConnectionStatus,
} from "./AgentSessionConnection";
import { projectGeneratedSessionSnapshot } from "./projectGeneratedSessionSnapshot";

export interface AgentSessionProjectionContext {
  agentLabel: string;
  interactionMode: "acp" | "pty";
  sessionId: string;
  title: string;
}

export function projectConnectedSession(
  connection: AgentSessionConnection,
  context: AgentSessionProjectionContext,
): AgentSessionSnapshot {
  const state = connection.getStore().getState();
  const connectionStatus = projectConnectionStatus(connection.getStatus());
  const transportError = connection.getError()?.message ?? null;
  if (!state) {
    return emptySession(context, connectionStatus, transportError);
  }
  const projected = projectGeneratedSessionSnapshot(state.snapshot, {
    title: context.title,
    agentLabel: context.agentLabel,
    connection: connectionStatus,
    interactionMode: context.interactionMode,
    hasOlderItems: false,
  });
  return transportError && !projected.error
    ? { ...projected, error: transportError }
    : projected;
}

function emptySession(
  context: AgentSessionProjectionContext,
  connection: AgentConnectionStatus,
  error: string | null,
): AgentSessionSnapshot {
  return {
    sessionId: context.sessionId,
    title: context.title,
    agentLabel: context.agentLabel,
    status: "launching",
    connection,
    interactionMode: context.interactionMode,
    capabilities: {
      sendMessage: false,
      interrupt: false,
      resolvePermission: false,
      updateConfiguration: false,
      terminal: false,
    },
    items: [],
    plan: [],
    permissions: [],
    terminals: [],
    hasOlderItems: false,
    error,
  };
}

function projectConnectionStatus(
  status: AgentSessionConnectionStatus,
): AgentConnectionStatus {
  if (status === "connected") return "connected";
  if (status === "reconnecting") return "reconnecting";
  if (status === "failed" || status === "disconnected") return "disconnected";
  return "connecting";
}
