// REST client for the coordinator (auto-harness) surface. Unlike the rest of
// the dashboard, coordinator is a web-only admin feature with no Rust-core
// mirror, so it talks to the Gin REST API directly via lightFetch (Bearer from
// localStorage) instead of the wasm Connect bridge.

import { lightFetch } from "@/lib/light-auth/api-fetch";
import { readCurrentOrg } from "@/stores/auth";

export interface CoordinatorClaimPolicy {
  labels?: string[];
  states?: string[];
  titleKeywords?: string[];
  bodyKeywords?: string[];
  unassignedOnly?: boolean;
  maxActiveTasks?: number;
}

export interface CoordinatorProject {
  id: number;
  slug: string;
  name: string;
  repository_id: number;
  platform_type: string;
  source_type: string;
  label_filter: string[] | null;
  agent_slug: string;
  scan_interval_seconds: number;
  max_concurrent: number;
  enabled: boolean;
  created_at: string;
}

export interface CoordinatorExecution {
  id: number;
  project_id: number;
  ticket_id: number;
  status: string;
  stage: string;
  external_id: string;
  summary: string;
  feedback_status: string;
  error: string;
  started_at?: string;
  finished_at?: string;
  created_at: string;
}

export interface CoordinatorRunResult {
  project_id: number;
  scanned: number;
  candidates: number;
  claimed: number;
  dispatched: number;
  skipped: number;
  errors: string[];
}

export interface CreateProjectInput {
  repository_id: number;
  name: string;
  platform_type?: string;
  source_type?: string;
  label_filter?: string[];
  claim_policy?: CoordinatorClaimPolicy;
  agent_slug?: string;
  scan_interval_seconds?: number;
  max_concurrent?: number;
}

export interface UpdateProjectInput {
  name?: string;
  label_filter?: string[];
  agent_slug?: string;
  scan_interval_seconds?: number;
  max_concurrent?: number;
  enabled?: boolean;
}

function base(): string {
  const slug = readCurrentOrg()?.slug ?? "";
  return `/api/v1/orgs/${slug}/coordinator/projects`;
}

export const coordinatorApi = {
  listProjects: async (): Promise<CoordinatorProject[]> => {
    const r = await lightFetch<{ projects: CoordinatorProject[] }>(base(), { authenticated: true });
    return r?.projects ?? [];
  },
  createProject: async (data: CreateProjectInput): Promise<CoordinatorProject> => {
    const r = await lightFetch<{ project: CoordinatorProject }>(base(), {
      method: "POST",
      body: data,
      authenticated: true,
    });
    return r.project;
  },
  updateProject: async (id: number, data: UpdateProjectInput): Promise<CoordinatorProject> => {
    const r = await lightFetch<{ project: CoordinatorProject }>(`${base()}/${id}`, {
      method: "PATCH",
      body: data,
      authenticated: true,
    });
    return r.project;
  },
  deleteProject: async (id: number): Promise<void> => {
    await lightFetch(`${base()}/${id}`, { method: "DELETE", authenticated: true });
  },
  listExecutions: async (id: number, limit = 50): Promise<CoordinatorExecution[]> => {
    const r = await lightFetch<{ executions: CoordinatorExecution[] }>(`${base()}/${id}/executions`, {
      authenticated: true,
      query: { limit },
    });
    return r?.executions ?? [];
  },
  runNow: async (id: number): Promise<CoordinatorRunResult> => {
    const r = await lightFetch<{ result: CoordinatorRunResult }>(`${base()}/${id}/run`, {
      method: "POST",
      authenticated: true,
    });
    return r.result;
  },
};
