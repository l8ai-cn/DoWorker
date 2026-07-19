import { apiFetch } from "./api-fetch";
import { listMobileAgents } from "./mobile-session-catalog";
import { createMobileWorkerSession } from "./mobile-session-creation";
import { PROJECT_LABEL_KEY } from "./project-label";
export {
  fetchSessionItems,
  openSessionStream,
  postMessage,
  postMessageContent,
  resolveElicitation,
  stopSession,
  type MessageContentBlock,
} from "./mobile-session-events";

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
  workerTypeSlug?: string;
  name: string;
  harness: string | null;
  supportedModes: SessionInteractionMode[];
  requiresModelResource: boolean;
}

export type SessionCreationAgent = Pick<
  AvailableAgent,
  "id" | "workerTypeSlug" | "supportedModes" | "requiresModelResource"
>;

export interface LiveSessionDetail extends LiveSessionSummary {
  events: import("./live-session-reducer").LiveAgentEvent[];
}

export interface SessionWire {
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
  const body = await createMobileWorkerSession(agent, title, initialText, options?.mode ?? "acp");
  return summaryFromWire(body);
}

export async function listAgents(): Promise<AvailableAgent[]> {
  return listMobileAgents();
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
