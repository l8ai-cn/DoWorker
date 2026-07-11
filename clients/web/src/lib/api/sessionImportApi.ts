import { getApiBaseUrl } from "@/lib/env";
import { getAuthManager } from "@/lib/wasm-core";
import { readCurrentOrg } from "@/stores/auth";

export type ImportCodexResult = {
  sessionId: string;
  podKey: string;
  agentId: string;
  title: string | null;
  sourceKind: string;
  sourceId: string;
  itemCount: number;
};

type ConversationItemWire = Record<string, unknown> & { id: string; type: string };

function orgHeaders(): HeadersInit | null {
  const token = getAuthManager().get_token();
  const org = readCurrentOrg()?.slug;
  if (!token || !org) return null;
  return {
    Authorization: `Bearer ${token}`,
    "X-Organization-Slug": org,
    "Content-Type": "application/json",
  };
}

async function sessionFetch(path: string, init?: RequestInit): Promise<Response> {
  const h = orgHeaders();
  if (!h) throw new Error("not authenticated");
  const base = getApiBaseUrl().replace(/\/$/, "");
  const res = await fetch(`${base}/v1${path}`, { ...init, headers: h });
  if (!res.ok) {
    const body = await res.text();
    throw new Error(body || `request failed: ${res.status}`);
  }
  return res;
}

async function sessionReq<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await sessionFetch(path, init);
  return res.json() as Promise<T>;
}

/** Migrate a local Codex rollout or output_* directory into a new Worker session. */
export async function importCodexSession(
  sourcePath: string,
  agentId: string,
  options: { title?: string; hostId?: string } = {},
): Promise<ImportCodexResult> {
  const body: Record<string, string> = {
    source_path: sourcePath,
    agent_id: agentId,
  };
  if (options.title) body.title = options.title;
  if (options.hostId) body.host_id = options.hostId;

  const wire = await sessionReq<{
    session: { id: string; agent_id: string; title?: string | null };
    pod_key: string;
    source_kind: string;
    source_id: string;
    item_count: number;
  }>("/sessions/import", {
    method: "POST",
    body: JSON.stringify(body),
  });

  return {
    sessionId: wire.session.id,
    podKey: wire.pod_key,
    agentId: wire.session.agent_id,
    title: wire.session.title ?? null,
    sourceKind: wire.source_kind,
    sourceId: wire.source_id,
    itemCount: wire.item_count,
  };
}

export type ImportedSessionSummary = {
  id: string;
  title: string | null;
  agentId: string;
  podKey: string | null;
  status: string;
  updatedAt: number;
};

/** List migrated / created Worker sessions for the sidebar. */
export async function listImportedSessions(): Promise<ImportedSessionSummary[]> {
  const wire = await sessionReq<{
    data: Array<{
      id: string;
      title?: string | null;
      agent_id: string;
      pod_key?: string | null;
      status: string;
      updated_at: number;
    }>;
  }>("/sessions?limit=50");

  return (wire.data ?? []).map((row) => ({
    id: row.id,
    title: row.title ?? null,
    agentId: row.agent_id,
    podKey: row.pod_key ?? null,
    status: row.status,
    updatedAt: row.updated_at,
  }));
}

/** Fetch session metadata linked to a Worker pod_key. */
export async function fetchSessionByPodKey(podKey: string): Promise<{ id: string; title: string | null } | null> {
  const res = await sessionFetch(`/sessions/by-pod/${encodeURIComponent(podKey)}`);
  if (res.status === 204) return null;
  const wire = await res.json() as { id: string; title?: string | null };
  return { id: wire.id, title: wire.title ?? null };
}

export async function fetchAllSessionItems(sessionId: string): Promise<ConversationItemWire[]> {
  const items: ConversationItemWire[] = [];
  let after = "";
  const limit = 200;

  for (;;) {
    const qs = new URLSearchParams({ limit: String(limit) });
    if (after) qs.set("after", after);
    const page = await sessionReq<{
      data: ConversationItemWire[];
      has_more: boolean;
    }>(`/sessions/${encodeURIComponent(sessionId)}/items?${qs.toString()}`);

    if (page.data.length === 0) break;
    items.push(...page.data);
    if (!page.has_more) break;
    after = page.data[page.data.length - 1]!.id;
  }

  return items;
}
