import { lightFetch } from "@/lib/light-auth/api-fetch";
import { readCurrentOrg } from "@/stores/auth";
import type { PodData } from "@/lib/api/facade/pod";

export interface ExpertKnowledgeMount {
  slug: string;
  mode?: string;
}

/**
 * Base64 avatar upload payload. Matches the backend `avatar` JSON field
 * (`avatarInput{ filename, content_base64 }` in
 * `backend/internal/api/rest/v1/expert_handler_types.go`). `content_base64`
 * is the raw base64 (no `data:` URL prefix); the backend sniffs the MIME type
 * and derives the stored path `assets/avatar.<ext>` — the filename is advisory.
 */
export interface ExpertAvatarUpload {
  filename: string;
  content_base64: string;
}

/**
 * Derived cache of `expert.json` extras persisted in the expert repo's
 * `metadata` jsonb. `avatar` is a repo-relative path (e.g. `assets/avatar.png`)
 * and `expertType` (类型) is a free-form category string.
 */
export interface ExpertMetadata {
  avatar?: string;
  expertType?: string;
  [key: string]: unknown;
}

export interface Expert {
  id: number;
  slug: string;
  name: string;
  description?: string | null;
  agent_slug: string;
  runner_id?: number | null;
  repository_id?: number | null;
  branch_name?: string | null;
  prompt?: string | null;
  interaction_mode: string;
  automation_level: string;
  perpetual: boolean;
  used_env_bundles: string[];
  skill_slugs: string[];
  knowledge_mounts: ExpertKnowledgeMount[] | string;
  config_overrides?: Record<string, unknown>;
  agentfile_layer?: string | null;
  source_pod_key?: string | null;
  worker_spec_snapshot_id?: number | null;
  orchestration_resource_id?: number | null;
  orchestration_resource_revision?: number | null;
  source_market_application_id?: number | null;
  source_market_release_id?: number | null;
  run_count: number;
  last_run_at?: string | null;
  created_at: string;
  updated_at: string;
  git_repo_path?: string | null;
  default_branch?: string;
  http_clone_url?: string | null;
  metadata?: ExpertMetadata;
}

export interface CreateExpertInput {
  name: string;
  slug: string;
  description?: string;
  agent_slug: string;
  runner_id?: number;
  repository_id?: number;
  branch_name?: string;
  prompt?: string;
  interaction_mode?: string;
  automation_level?: string;
  perpetual?: boolean;
  used_env_bundles?: string[];
  skill_slugs?: string[];
  knowledge_mounts?: ExpertKnowledgeMount[];
  config_overrides?: Record<string, unknown>;
  agentfile_layer?: string;
  avatar?: ExpertAvatarUpload;
  expert_type?: string;
}

export interface UpdateExpertInput {
  name?: string;
  description?: string;
  agent_slug?: string;
  runner_id?: number;
  repository_id?: number;
  branch_name?: string;
  prompt?: string;
  interaction_mode?: string;
  automation_level?: string;
  perpetual?: boolean;
  used_env_bundles?: string[];
  skill_slugs?: string[];
  knowledge_mounts?: ExpertKnowledgeMount[];
  config_overrides?: Record<string, unknown>;
  agentfile_layer?: string;
  avatar?: ExpertAvatarUpload;
  expert_type?: string;
}

export interface PublishExpertInput {
  name: string;
  slug: string;
  description?: string;
}

export interface RunExpertInput {
  alias?: string;
  prompt_override?: string;
  cols?: number;
  rows?: number;
}

function base(): string {
  const slug = readCurrentOrg()?.slug ?? "";
  return `/api/v1/orgs/${slug}/experts`;
}

export function parseExpertKnowledgeMounts(
  raw: Expert["knowledge_mounts"],
): ExpertKnowledgeMount[] {
  if (Array.isArray(raw)) return raw;
  if (typeof raw === "string" && raw.trim()) {
    try {
      const parsed = JSON.parse(raw) as ExpertKnowledgeMount[];
      return Array.isArray(parsed) ? parsed : [];
    } catch {
      return [];
    }
  }
  return [];
}

export const expertApi = {
  list: async (limit = 50, offset = 0): Promise<{ experts: Expert[]; total: number }> => {
    const r = await lightFetch<{ experts: Expert[]; total: number }>(base(), {
      authenticated: true,
      query: { limit, offset },
    });
    return { experts: r?.experts ?? [], total: r?.total ?? 0 };
  },

  get: async (expertSlug: string): Promise<Expert> => {
    const r = await lightFetch<{ expert: Expert }>(`${base()}/${expertSlug}`, {
      authenticated: true,
    });
    return r.expert;
  },

  create: async (data: CreateExpertInput): Promise<Expert> => {
    const r = await lightFetch<{ expert: Expert }>(base(), {
      method: "POST",
      body: data,
      authenticated: true,
    });
    return r.expert;
  },

  update: async (expertSlug: string, data: UpdateExpertInput): Promise<Expert> => {
    const r = await lightFetch<{ expert: Expert }>(`${base()}/${expertSlug}`, {
      method: "PATCH",
      body: data,
      authenticated: true,
    });
    return r.expert;
  },

  delete: async (expertSlug: string): Promise<void> => {
    await lightFetch(`${base()}/${expertSlug}`, { method: "DELETE", authenticated: true });
  },

  run: async (
    expertSlug: string,
    data: RunExpertInput = {},
  ): Promise<{ pod: PodData; warning?: string }> => {
    const r = await lightFetch<{ pod: PodData; warning?: string }>(
      `${base()}/${expertSlug}/run`,
      { method: "POST", body: data, authenticated: true },
    );
    return { pod: r.pod, warning: r.warning };
  },

  publishFromPod: async (podKey: string, data: PublishExpertInput): Promise<Expert> => {
    const orgSlug = readCurrentOrg()?.slug ?? "";
    const r = await lightFetch<{ expert: Expert }>(
      `/api/v1/orgs/${orgSlug}/pods/${podKey}/publish-expert`,
      { method: "POST", body: data, authenticated: true },
    );
    return r.expert;
  },
};
