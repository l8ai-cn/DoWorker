import type { KnowledgeMountSelection } from "@/lib/api/facade/knowledgeBaseApi";
import {
  parseExpertKnowledgeMounts,
  type Expert,
  type ExpertAvatarUpload,
} from "@/lib/api/expertApi";

/**
 * A locally-selected avatar (形象) upload before it is committed. `contentBase64`
 * is the raw base64 (no `data:` prefix) sent as the backend `content_base64`
 * field; `previewUrl` is a full data URL used only for the on-page preview.
 */
export interface ExpertAvatarDraft {
  filename: string;
  contentBase64: string;
  previewUrl: string;
}

export interface ExpertFormState {
  name: string;
  slug: string;
  description: string;
  agentSlug: string;
  prompt: string;
  interactionMode: string;
  perpetual: boolean;
  skillSlugs: string[];
  runnerId: number | null;
  repositoryId: number | null;
  branchName: string;
  usedEnvBundles: string[];
  knowledgeMounts: KnowledgeMountSelection[];
  configOverrides: string;
  agentfileLayer: string;
  avatar: ExpertAvatarDraft | null;
  expertType: string;
}

export const EMPTY_EXPERT_FORM: ExpertFormState = {
  name: "",
  slug: "",
  description: "",
  agentSlug: "",
  prompt: "",
  interactionMode: "pty",
  perpetual: false,
  skillSlugs: [],
  runnerId: null,
  repositoryId: null,
  branchName: "",
  usedEnvBundles: [],
  knowledgeMounts: [],
  configOverrides: "",
  agentfileLayer: "",
  avatar: null,
  expertType: "",
};

export function slugifyExpert(name: string): string {
  return name
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "")
    .slice(0, 100);
}

export function expertToForm(e: Expert): ExpertFormState {
  return {
    name: e.name ?? "",
    slug: e.slug,
    description: e.description ?? "",
    agentSlug: e.agent_slug,
    prompt: e.prompt ?? "",
    interactionMode: e.interaction_mode || "pty",
    perpetual: e.perpetual,
    skillSlugs: e.skill_slugs ?? [],
    runnerId: e.runner_id ?? null,
    repositoryId: e.repository_id ?? null,
    branchName: e.branch_name ?? "",
    usedEnvBundles: e.used_env_bundles ?? [],
    knowledgeMounts: parseExpertKnowledgeMounts(e.knowledge_mounts).map((m) => ({
      slug: m.slug,
      mode: m.mode === "rw" ? "rw" : "ro",
    })),
    configOverrides:
      e.config_overrides && Object.keys(e.config_overrides).length > 0
        ? JSON.stringify(e.config_overrides, null, 2)
        : "",
    agentfileLayer: e.agentfile_layer ?? "",
    // Avatar bytes are not re-hydrated into the form; the existing repo-relative
    // path lives in metadata. A draft is only set when the user picks a new file.
    avatar: null,
    expertType: e.metadata?.expertType ?? "",
  };
}

export function isValidConfigOverrides(raw: string): boolean {
  const trimmed = raw.trim();
  if (!trimmed) return true;
  try {
    const parsed = JSON.parse(trimmed);
    return parsed != null && typeof parsed === "object" && !Array.isArray(parsed);
  } catch {
    return false;
  }
}

export class ExpertConfigJsonError extends Error {}

export interface ExpertConfigPayload {
  name: string;
  description: string;
  prompt: string;
  interaction_mode: string;
  perpetual: boolean;
  skill_slugs: string[];
  runner_id?: number;
  repository_id?: number;
  branch_name?: string;
  used_env_bundles: string[];
  knowledge_mounts: { slug: string; mode: string }[];
  config_overrides?: Record<string, unknown>;
  agentfile_layer?: string;
  avatar?: ExpertAvatarUpload;
  expert_type?: string;
}

export function buildExpertConfig(form: ExpertFormState): ExpertConfigPayload {
  let configOverrides: Record<string, unknown> | undefined;
  const raw = form.configOverrides.trim();
  if (raw) {
    if (!isValidConfigOverrides(raw)) throw new ExpertConfigJsonError("invalid config_overrides");
    configOverrides = JSON.parse(raw) as Record<string, unknown>;
  }
  return {
    name: form.name.trim(),
    description: form.description.trim(),
    prompt: form.prompt,
    interaction_mode: form.interactionMode,
    perpetual: form.perpetual,
    skill_slugs: form.skillSlugs,
    runner_id: form.runnerId ?? undefined,
    repository_id: form.repositoryId ?? undefined,
    branch_name: form.branchName.trim() || undefined,
    used_env_bundles: form.usedEnvBundles,
    knowledge_mounts: form.knowledgeMounts.map((m) => ({ slug: m.slug, mode: m.mode })),
    config_overrides: configOverrides,
    agentfile_layer: form.agentfileLayer.trim() || undefined,
    avatar: form.avatar
      ? { filename: form.avatar.filename, content_base64: form.avatar.contentBase64 }
      : undefined,
    expert_type: form.expertType.trim() || undefined,
  };
}
