import type { ConversationItem } from "@/lib/conversationItems";
import type { SessionStatus } from "@/lib/types";
import type { TerminalResource } from "@do-worker/agent-ui";
import type { EmbeddedItemsPage, EmbeddedSession, EmbedRelayConnection } from "./embed-session-api";

interface SessionWire {
  agent_name?: unknown;
  id?: unknown;
  interaction_mode?: unknown;
  pod_key?: unknown;
  runner_id?: unknown;
  title?: unknown;
  total_cost_usd?: unknown;
  status?: unknown;
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
    typeof body.id !== "string" ||
    (body.agent_name !== undefined && typeof body.agent_name !== "string") ||
    (body.interaction_mode !== undefined &&
      body.interaction_mode !== "acp" &&
      body.interaction_mode !== "pty") ||
    (body.pod_key !== undefined && body.pod_key !== null && typeof body.pod_key !== "string") ||
    (body.runner_id !== undefined &&
      body.runner_id !== null &&
      typeof body.runner_id !== "string") ||
    (typeof body.title !== "string" && body.title !== null && body.title !== undefined) ||
    (body.total_cost_usd !== undefined &&
      body.total_cost_usd !== null &&
      typeof body.total_cost_usd !== "number") ||
    !isSessionStatus(body.status)
  ) {
    throw new Error("Embedded session response is invalid");
  }
  return {
    agentLabel: body.agent_name ?? "Agent",
    id: body.id,
    interactionMode: body.interaction_mode ?? "acp",
    podKey: body.pod_key ?? null,
    runnerId: body.runner_id ?? null,
    title: body.title ?? null,
    totalCostUsd: body.total_cost_usd ?? null,
    status: body.status,
  };
}

export function parseEmbeddedItems(value: unknown): EmbeddedItemsPage {
  const body = value as { data?: unknown; has_more?: unknown };
  if (!Array.isArray(body.data) || typeof body.has_more !== "boolean") {
    throw new Error("Embedded session items response is invalid");
  }
  return {
    items: [...body.data].reverse() as ConversationItem[],
    hasMore: body.has_more,
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

function isSessionStatus(value: unknown): value is SessionStatus {
  return (
    value === "idle" ||
    value === "launching" ||
    value === "running" ||
    value === "waiting" ||
    value === "failed"
  );
}
