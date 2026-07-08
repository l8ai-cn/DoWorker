import { apiFetch } from "./api-fetch";
import { orgScopedApiPath } from "./org-api-path";

export interface LiveExpert {
  slug: string;
  name: string;
  description: string | null;
  agent_slug: string;
  interaction_mode: "pty" | "acp" | string;
  run_count: number;
  last_run_at: string | null;
  prompt: string | null;
}

interface ListWire {
  experts: LiveExpert[];
  total: number;
}

interface ExpertWire {
  expert: LiveExpert;
}

interface PodWire {
  pod_key: string;
  session_id?: string | null;
  external_session_id?: string | null;
  interaction_mode?: string;
  status?: string;
}

async function readJson<T>(res: Response): Promise<T> {
  if (!res.ok) {
    const text = await res.text();
    throw new Error(text || `HTTP ${res.status}`);
  }
  return res.json() as Promise<T>;
}

export async function listLiveExperts(limit = 50, offset = 0): Promise<LiveExpert[]> {
  const res = await apiFetch(
    orgScopedApiPath(`/experts?limit=${limit}&offset=${offset}`),
  );
  const body = await readJson<ListWire>(res);
  return body.experts ?? [];
}

export async function getLiveExpert(slug: string): Promise<LiveExpert | null> {
  const res = await apiFetch(orgScopedApiPath(`/experts/${encodeURIComponent(slug)}`));
  if (res.status === 404) return null;
  const body = await readJson<ExpertWire>(res);
  return body.expert ?? null;
}

export async function runLiveExpert(
  slug: string,
  promptOverride?: string,
): Promise<{ pod: PodWire; warning?: string }> {
  const res = await apiFetch(orgScopedApiPath(`/experts/${encodeURIComponent(slug)}/run`), {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(
      promptOverride?.trim() ? { prompt_override: promptOverride.trim() } : {},
    ),
  });
  const body = await readJson<{ pod: PodWire; warning?: string }>(res);
  return body;
}

export function sessionIdFromPod(pod: PodWire): string | null {
  const ext = pod.external_session_id?.trim();
  if (ext) return ext;
  const sid = pod.session_id?.trim();
  return sid || null;
}
