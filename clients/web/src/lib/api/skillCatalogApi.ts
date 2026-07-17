import { lightFetch } from "@/lib/light-auth/api-fetch";
import { readCurrentOrg } from "@/stores/auth";
import type { CatalogSkill, SkillImportAuthType } from "@/lib/viewModels/extension";

export interface ImportSkillsInput {
  url: string;
  branch?: string;
  subdir?: string;
  agent_filter?: string[];
  auth_type?: SkillImportAuthType;
  auth_credential?: string;
}

export interface UpdateCatalogSkillInput {
  name?: string;
  description?: string;
  license?: string;
  instructions?: string;
  tags?: string[];
}

export interface ImportSkillsResult {
  skills: CatalogSkill[];
  imported: number;
  partial_errors?: string;
}

function base(): string {
  const slug = readCurrentOrg()?.slug ?? "";
  return `/api/v1/orgs/${slug}/authored-skills`;
}

export const skillCatalogApi = {
  list: async (limit = 50, offset = 0): Promise<{ skills: CatalogSkill[]; total: number }> => {
    const r = await lightFetch<{ skills: CatalogSkill[]; total: number }>(base(), {
      authenticated: true,
      query: { limit, offset },
    });
    return { skills: r?.skills ?? [], total: r?.total ?? 0 };
  },

  listAll: async (): Promise<{ skills: CatalogSkill[]; total: number }> => {
    const r = await lightFetch<{ skills: CatalogSkill[]; total: number }>(base(), {
      authenticated: true,
      query: { all: true },
    });
    return { skills: r?.skills ?? [], total: r?.total ?? 0 };
  },

  get: async (skillSlug: string): Promise<CatalogSkill> => {
    const r = await lightFetch<{ skill: CatalogSkill }>(`${base()}/${skillSlug}`, {
      authenticated: true,
    });
    return r.skill;
  },

  import: async (data: ImportSkillsInput): Promise<ImportSkillsResult> => {
    const r = await lightFetch<ImportSkillsResult>(`${base()}/import`, {
      method: "POST",
      body: data,
      authenticated: true,
    });
    return { skills: r?.skills ?? [], imported: r?.imported ?? 0, partial_errors: r?.partial_errors };
  },

  update: async (skillSlug: string, data: UpdateCatalogSkillInput): Promise<CatalogSkill> => {
    const r = await lightFetch<{ skill: CatalogSkill }>(`${base()}/${skillSlug}`, {
      method: "PATCH",
      body: data,
      authenticated: true,
    });
    return r.skill;
  },

  delete: async (skillSlug: string): Promise<void> => {
    await lightFetch(`${base()}/${skillSlug}`, { method: "DELETE", authenticated: true });
  },

  syncUpstream: async (skillSlug: string): Promise<CatalogSkill> => {
    const r = await lightFetch<{ skill: CatalogSkill }>(`${base()}/${skillSlug}/sync-upstream`, {
      method: "POST",
      authenticated: true,
    });
    return r.skill;
  },
};
