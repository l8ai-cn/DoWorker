import { apiFetch } from "./api-fetch";
import { resolveDefaultModelResourceId } from "./model-resources-api";
import { PROJECT_LABEL_KEY } from "./project-label";

export type SessionStatus = "idle" | "launching" | "running" | "waiting" | "failed";
export type SessionInteractionMode = "acp" | "pty";

export interface LiveSessionSummary {
  id: string;
  title: string | null;
  agentId: string;
  agentName: string | null;
  podKey: string | null;
  status: SessionStatus;
  pendingApprovals: number;
  updatedAt: string;
  workspace: string | null;
  project: string | null;
  interactionMode: SessionInteractionMode | null;
}

export interface AvailableAgent {
  id: string;
  name: string;
  harness: string | null;
  supportedModes: SessionInteractionMode[];
  requiresModelResource: boolean;
}

export type SessionCreationAgent = Pick<AvailableAgent, "id" | "requiresModelResource">;

export interface LiveSessionDetail extends LiveSessionSummary {
  events: import("./live-session-reducer").LiveAgentEvent[];
}

interface SessionWire {
  id: string;
  agent_id: string;
  agent_name?: string | null;
  pod_key?: string | null;
  title?: string | null;
  status: SessionStatus;
  pending_elicitations_count?: number;
  workspace?: string | null;
  updated_at?: number;
  created_at?: number;
  labels?: Record<string, string>;
  interaction_mode?: SessionInteractionMode;
}

interface ListWire {
  data: Array<{
    id: string;
    title: string | null;
    agent_id?: string;
    agent_name?: string | null;
    pod_key?: string | null;
    status?: SessionStatus;
    pending_elicitations_count?: number;
    workspace?: string | null;
    updated_at?: number;
    labels?: Record<string, string>;
    interaction_mode?: SessionInteractionMode;
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
    podKey: w.pod_key ?? null,
    status: w.status ?? "idle",
    pendingApprovals: w.pending_elicitations_count ?? 0,
    updatedAt: formatRelative(w.updated_at),
    workspace: w.workspace ?? null,
    project,
    interactionMode: w.interaction_mode ?? null,
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

export async function getSessionByPodKey(podKey: string): Promise<LiveSessionSummary | null> {
  const res = await apiFetch(`/v1/sessions/by-pod/${encodeURIComponent(podKey)}`);
  if (res.status === 204) return null;
  const body = await readJson<SessionWire>(res);
  return summaryFromWire(body);
}

export async function createSession(
  agent: SessionCreationAgent,
  title?: string,
  initialText?: string,
  options?: { mode?: SessionInteractionMode },
): Promise<LiveSessionSummary> {
  const modelResourceId = agent.requiresModelResource
    ? await resolveDefaultModelResourceId()
    : undefined;
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
      agent_id: agent.id,
      initial_items,
      ...(modelResourceId !== undefined ? { model_resource_id: modelResourceId } : {}),
      ...(title ? { title } : {}),
      ...(options?.mode === "pty" ? { pty_only: true } : {}),
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
  const body = await readJson<{
    data: Array<{
      id: string;
      name: string;
      harness?: string;
      supported_modes?: string[];
      requires_model_resource?: boolean;
    }>;
  }>(res);
  return (body.data ?? []).map((a) => availableAgentFromWire(a));
}

function availableAgentFromWire(a: {
  id: string;
  name: string;
  harness?: string;
  supported_modes?: string[];
  requires_model_resource?: boolean;
}): AvailableAgent {
  if (typeof a.requires_model_resource !== "boolean") {
    throw new Error(`Worker ${a.id} 未声明模型资源要求`);
  }
  return {
    id: a.id,
    name: a.name,
    harness: a.harness ?? null,
    supportedModes: readSupportedModes(a.id, a.supported_modes),
    requiresModelResource: a.requires_model_resource,
  };
}

function readSupportedModes(
  agentID: string,
  modes: string[] | undefined,
): SessionInteractionMode[] {
  if (!modes || modes.length === 0) {
    throw new Error(`Worker ${agentID} 未声明支持的交互模式`);
  }
  const supportedModes = modes.filter(
    (mode): mode is SessionInteractionMode => mode === "acp" || mode === "pty",
  );
  if (
    supportedModes.length !== modes.length ||
    new Set(supportedModes).size !== supportedModes.length
  ) {
    throw new Error(`Worker ${agentID} 返回了无效的交互模式`);
  }
  return supportedModes;
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
