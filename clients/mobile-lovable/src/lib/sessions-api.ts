import { apiFetch } from "./api-fetch";
import { PROJECT_LABEL_KEY } from "./project-label";

export type SessionStatus = "idle" | "launching" | "running" | "waiting" | "failed";

export interface LiveSessionSummary {
  id: string;
  title: string | null;
  agentId: string;
  agentName: string | null;
  status: SessionStatus;
  pendingApprovals: number;
  updatedAt: string;
  workspace: string | null;
  project: string | null;
}

export interface AvailableAgent {
  id: string;
  name: string;
  harness: string | null;
}

export interface LiveSessionDetail extends LiveSessionSummary {
  events: import("./live-session-reducer").LiveAgentEvent[];
}

interface SessionWire {
  id: string;
  agent_id: string;
  agent_name?: string | null;
  title?: string | null;
  status: SessionStatus;
  pending_elicitations_count?: number;
  workspace?: string | null;
  updated_at?: number;
  created_at?: number;
  labels?: Record<string, string>;
}

interface ListWire {
  data: Array<{
    id: string;
    title: string | null;
    agent_id?: string;
    agent_name?: string | null;
    status?: SessionStatus;
    pending_elicitations_count?: number;
    workspace?: string | null;
    updated_at?: number;
    labels?: Record<string, string>;
  }>;
}

async function readJson<T>(res: Response): Promise<T> {
  if (!res.ok) {
    const text = await res.text();
    throw new Error(text || `HTTP ${res.status}`);
  }
  return res.json() as Promise<T>;
}

function formatRelative(ts?: number): string {
  if (!ts) return "刚刚";
  const diff = Date.now() - ts * 1000;
  const mins = Math.floor(diff / 60_000);
  if (mins < 1) return "刚刚";
  if (mins < 60) return `${mins} 分钟前`;
  const hrs = Math.floor(mins / 60);
  if (hrs < 24) return `${hrs} 小时前`;
  return `${Math.floor(hrs / 24)} 天前`;
}

function summaryFromWire(w: SessionWire | ListWire["data"][number]): LiveSessionSummary {
  const labels = "labels" in w ? w.labels : undefined;
  const project = labels?.[PROJECT_LABEL_KEY]?.trim() || null;
  return {
    id: w.id,
    title: w.title ?? null,
    agentId: w.agent_id ?? "agent",
    agentName: w.agent_name ?? null,
    status: w.status ?? "idle",
    pendingApprovals: w.pending_elicitations_count ?? 0,
    updatedAt: formatRelative(w.updated_at),
    workspace: w.workspace ?? null,
    project,
  };
}

export async function listSessions(limit = 50): Promise<LiveSessionSummary[]> {
  const res = await apiFetch(`/v1/sessions?limit=${limit}`);
  const body = await readJson<ListWire>(res);
  return (body.data ?? []).map(summaryFromWire);
}

export async function getSession(sessionId: string): Promise<LiveSessionSummary> {
  const res = await apiFetch(`/v1/sessions/${encodeURIComponent(sessionId)}`);
  const body = await readJson<SessionWire>(res);
  return summaryFromWire(body);
}

export interface ModelConfig {
  id: number;
  name: string;
  provider_type: string;
  model: string;
  is_default: boolean;
}

async function resolveModelConfigId(agentId: string): Promise<number | undefined> {
  const res = await apiFetch("/v1/model-configs");
  if (!res.ok) return undefined;
  const body = (await res.json()) as { data?: ModelConfig[] };
  const models = body.data ?? [];
  const preferred = agentId === "codex-cli" ? "openai" : agentId === "do-agent" ? undefined : undefined;
  if (preferred) {
    const match = models.find((m) => m.provider_type === preferred && m.is_default)
      ?? models.find((m) => m.provider_type === preferred);
    if (match) return match.id;
  }
  const fallback = models.find((m) => m.is_default);
  return fallback?.id;
}

export async function createSession(
  agentId: string,
  title?: string,
  initialText?: string,
): Promise<LiveSessionSummary> {
  const modelConfigId = await resolveModelConfigId(agentId);
  const initial_items = initialText?.trim()
    ? [
        {
          type: "message",
          data: {
            role: "user",
            content: [{ type: "input_text", text: initialText.trim() }],
          },
        },
      ]
    : [];
  const res = await apiFetch("/v1/sessions", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      agent_id: agentId,
      initial_items,
      ...(title ? { title } : {}),
      ...(modelConfigId != null ? { model_config_id: modelConfigId } : {}),
    }),
  });
  const body = await readJson<SessionWire>(res);
  return summaryFromWire(body);
}

export type MessageContentBlock =
  | { type: "input_text"; text: string }
  | { type: "input_image"; file_id: string; filename?: string }
  | { type: "input_file"; file_id: string; filename: string };

export async function postMessageContent(
  sessionId: string,
  content: MessageContentBlock[],
): Promise<void> {
  const res = await apiFetch(`/v1/sessions/${encodeURIComponent(sessionId)}/events`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      type: "message",
      data: { role: "user", content },
    }),
  });
  if (!res.ok) throw new Error(await res.text());
}

export async function postMessage(sessionId: string, text: string): Promise<void> {
  await postMessageContent(sessionId, [{ type: "input_text", text }]);
}

export async function stopSession(sessionId: string): Promise<void> {
  const res = await apiFetch(`/v1/sessions/${encodeURIComponent(sessionId)}/events`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ type: "stop_session", data: {} }),
  });
  if (!res.ok) throw new Error(await res.text());
}

export async function resolveElicitation(
  sessionId: string,
  elicitationId: string,
  accept: boolean,
): Promise<void> {
  const res = await apiFetch(
    `/v1/sessions/${encodeURIComponent(sessionId)}/elicitations/${encodeURIComponent(elicitationId)}/resolve`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ action: accept ? "accept" : "decline" }),
    },
  );
  if (!res.ok) throw new Error(await res.text());
}

export function openSessionStream(sessionId: string, signal: AbortSignal): Promise<Response> {
  return apiFetch(`/v1/sessions/${encodeURIComponent(sessionId)}/stream`, {
    headers: { Accept: "text/event-stream" },
    signal,
  });
}

interface ItemsPageWire {
  data: Array<Record<string, unknown>>;
  has_more: boolean;
}

export async function fetchSessionItems(
  sessionId: string,
  limit = 50,
): Promise<Array<Record<string, unknown>>> {
  const res = await apiFetch(
    `/v1/sessions/${encodeURIComponent(sessionId)}/items?limit=${limit}&order=desc`,
  );
  const page = await readJson<ItemsPageWire>(res);
  return [...(page.data ?? [])].reverse();
}

export async function listAgents(): Promise<AvailableAgent[]> {
  const res = await apiFetch("/v1/agents");
  const body = await readJson<{ data: Array<{ id: string; name: string; harness?: string }> }>(res);
  return (body.data ?? []).map((a) => ({
    id: a.id,
    name: a.name,
    harness: a.harness ?? null,
  }));
}

export async function listProjects(): Promise<string[]> {
  const res = await apiFetch("/v1/sessions/projects");
  return readJson<string[]>(res);
}

export async function assignSessionProject(sessionId: string, project: string): Promise<void> {
  const res = await apiFetch(`/v1/sessions/${encodeURIComponent(sessionId)}`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ labels: { [PROJECT_LABEL_KEY]: project.trim() } }),
  });
  if (!res.ok) throw new Error(await res.text());
}
