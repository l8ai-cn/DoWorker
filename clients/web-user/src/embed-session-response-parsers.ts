import type { TerminalResource } from "@agent-cloud/agent-ui";
import type { EmbeddedSession, EmbedRelayConnection } from "./embed-session-api";

interface SessionWire {
  agent_name?: unknown;
  interaction_mode?: unknown;
  title?: unknown;
}

export async function readEmbeddedJson(response: Response): Promise<unknown> {
  if (!response.ok) {
    throw new Error(`Embedded session request failed (${response.status})`);
  }
  return response.json();
}

export function parseEmbeddedSession(value: unknown): EmbeddedSession {
  const body = value as SessionWire;
  if (
    (body.agent_name !== undefined && typeof body.agent_name !== "string") ||
    (body.interaction_mode !== undefined &&
      body.interaction_mode !== "acp" &&
      body.interaction_mode !== "pty") ||
    (typeof body.title !== "string" && body.title !== null && body.title !== undefined)
  ) {
    throw new Error("Embedded session response is invalid");
  }
  return {
    agentLabel: body.agent_name ?? "Agent",
    interactionMode: body.interaction_mode ?? "acp",
    title: body.title ?? "",
  };
}

export function parseEmbeddedTerminals(value: unknown, writable: boolean): TerminalResource[] {
  const body = value as { data?: unknown };
  if (!Array.isArray(body.data)) {
    throw new Error("Embedded terminal response is invalid");
  }
  return body.data.map((item) => {
    const row = item as {
      id?: unknown;
      name?: unknown;
      metadata?: { running?: unknown };
    };
    if (typeof row.id !== "string" || typeof row.name !== "string") {
      throw new Error("Embedded terminal response is invalid");
    }
    return {
      id: row.id,
      label: row.name,
      status: row.metadata?.running === false ? "disconnected" : "connected",
      writable,
    };
  });
}

export function parseEmbeddedRelayConnection(value: unknown): EmbedRelayConnection {
  const body = value as {
    relay_url?: unknown;
    token?: unknown;
    pod_key?: unknown;
  };
  if (
    typeof body.relay_url !== "string" ||
    typeof body.token !== "string" ||
    typeof body.pod_key !== "string"
  ) {
    throw new Error("Embedded Relay connection response is invalid");
  }
  return {
    relayUrl: body.relay_url,
    token: body.token,
    podKey: body.pod_key,
  };
}
