/**
 * Extension ViewModels — UI-side projection of `proto.extension.v1` types.
 *
 * The proto wire encodes args/http_headers/env_vars as JSON strings (backend
 * keeps them verbatim because the schema is user-defined) and never carries
 * the denormalized `market_item` join — these UI shapes are not direct proto
 * mirrors but business projections. The adapters in `lib/api/*Extension.ts`
 * parse the JSON strings and surface these ViewModels.
 */
/** Git auth mode when importing a skill from an external repository. */
export type SkillImportAuthType = "none" | "github_pat" | "gitlab_pat" | "ssh_key";

/**
 * Unified skill-catalog row (mirrors the REST `skilldom.Skill` JSON at
 * `/orgs/:slug/authored-skills`). Every skill — platform-authored or
 * imported from an external git repo — is one row backed by its own
 * internal git repo. `install_source` distinguishes the two; the
 * `upstream_*` fields carry provenance for imports.
 */
export interface CatalogSkill {
  id: number;
  organization_id: number | null;
  slug: string;
  display_name: string;
  description: string;
  license: string;
  category?: string;
  compatibility?: string;
  allowed_tools?: string;
  agent_filter?: string[];
  is_active: boolean;
  git_repo_path: string;
  default_branch: string;
  http_clone_url?: string;
  upstream_url?: string;
  upstream_subdir?: string;
  upstream_commit_sha?: string;
  install_source: "gitops" | "import";
  content_sha: string;
  storage_key: string;
  package_size: number;
  version: number;
  created_by_id?: number;
  created_at: string;
  updated_at: string;
}

export interface SkillMarketItem {
  id: number;
  slug: string;
  display_name: string;
  description: string;
  license: string;
  category: string;
  content_sha: string;
  version: number;
  is_active: boolean;
}

export interface McpMarketItem {
  id: number;
  slug: string;
  name: string;
  description: string;
  icon: string;
  transport_type: string;
  command: string;
  default_args?: string[] | null;
  default_http_url?: string;
  default_http_headers?: McpHeaderSchemaEntry[] | null;
  env_var_schema?: EnvVarSchemaEntry[] | null;
  category: string;
  source?: string;
  registry_name?: string;
  version?: string;
  repository_url?: string;
}

export interface McpHeaderSchemaEntry {
  name: string;
  description?: string;
  value?: string;
  required: boolean;
  sensitive: boolean;
}

export interface EnvVarSchemaEntry {
  name: string;
  label: string;
  required: boolean;
  sensitive: boolean;
  placeholder?: string;
}

export interface InstalledSkill {
  id: number;
  organization_id: number;
  repository_id: number;
  market_item_id: number | null;
  scope: "org" | "user";
  installed_by: number;
  slug: string;
  install_source: "market" | "github" | "upload";
  source_url: string;
  content_sha: string;
  package_size: number;
  pinned_version: number | null;
  is_enabled: boolean;
  market_item?: SkillMarketItem;
}

export interface InstalledMcpServer {
  id: number;
  organization_id: number;
  repository_id: number;
  market_item_id: number | null;
  scope: "org" | "user";
  installed_by: number;
  name: string;
  slug: string;
  transport_type: string;
  command: string;
  args?: string[] | null;
  http_url: string;
  http_headers?: Record<string, string> | null;
  env_vars: Record<string, string>;
  is_enabled: boolean;
  market_item?: McpMarketItem;
}
