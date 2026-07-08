/**
 * Utilities for generating AgentFile Layer source from form fields.
 * An AgentFile Layer is a DSL fragment that configures a Pod's environment.
 */

import { POD_MODE_PTY } from "@/lib/pod-modes";

/**
 * Escape a string for use in an AgentFile quoted value.
 * Must align with backend FormatStringLiteral (agentfile/format.go).
 */
function escapeAgentfileString(s: string): string {
  return s
    .replace(/\\/g, "\\\\")
    .replace(/"/g, '\\"')
    .replace(/\n/g, "\\n")
    .replace(/\t/g, "\\t");
}

/**
 * Escape and quote a string value for AgentFile syntax.
 * Must align with backend FormatStringLiteral (agentfile/format.go).
 */
function formatAgentfileValue(value: unknown): string {
  if (typeof value === "string") return `"${escapeAgentfileString(value)}"`;
  if (typeof value === "boolean") return value ? "true" : "false";
  if (typeof value === "number") return String(value);
  return `"${escapeAgentfileString(String(value))}"`;
}

/**
 * Build an AgentFile Layer source string from structured form parameters.
 * Each non-empty field is emitted as an AgentFile declaration line.
 */
export function buildAgentfileLayer(params: {
  configValues: Record<string, unknown>;
  repositorySlug?: string;
  branchName?: string;
  interactionMode?: string;
  /**
   * Credential bundle name (kind='credential') to attach. Emitted FIRST in
   * the USE_ENV_BUNDLE sequence so runtime preferences listed after can
   * override credential defaults on conflicting keys.
   * Empty string / undefined = no credential injection (Agent uses its own
   * default auth: OAuth, CLI login, etc.).
   */
  credentialBundleName?: string;
  /**
   * Runtime bundle names (kind='runtime') to attach. Emitted AFTER the
   * credential line, in array order. Later entries override earlier ones
   * on conflicting env keys (mirrors backend eval order).
   */
  runtimeBundleNames?: string[];
  /** Config JSON bundle names (kind='config') — emitted as USE_CONFIG_BUNDLE. */
  configBundleNames?: string[];
  /** Installed skill slugs to attach. Emitted as `SKILLS slug1, slug2`. */
  skillSlugs?: string[];
  /** Knowledge base mounts. Emitted as `KNOWLEDGE slug [rw], slug2`. */
  knowledgeMounts?: { slug: string; mode: "ro" | "rw" }[];
  /**
   * Optional per-Worker token budget cap (positive integer). Emitted as
   * `CONFIG token_budget = "N"` — the same directive the backend
   * orchestrator appends for model-pool bindings, so this stays within the
   * existing AgentFile contract.
   */
  tokenBudget?: number | null;
  prompt?: string;
}): string {
  const lines: string[] = [];

  // MODE declaration (if not default PTY)
  if (params.interactionMode && params.interactionMode !== POD_MODE_PTY) {
    lines.push(`MODE ${params.interactionMode}`);
  }

  // USE_ENV_BUNDLE declarations — credential first, then runtime bundles
  // in selection order. Backend's eval merges each bundle's KV into the
  // Pod's env in declaration order; later wins on conflicts.
  const bundleNames: string[] = [];
  if (params.credentialBundleName) {
    bundleNames.push(params.credentialBundleName);
  }
  if (params.runtimeBundleNames) {
    for (const name of params.runtimeBundleNames) {
      if (name) bundleNames.push(name);
    }
  }
  for (const name of bundleNames) {
    lines.push(`USE_ENV_BUNDLE "${escapeAgentfileString(name)}"`);
  }
  if (params.configBundleNames) {
    for (const name of params.configBundleNames) {
      if (name) {
        lines.push(`USE_CONFIG_BUNDLE "${escapeAgentfileString(name)}"`);
      }
    }
  }

  // PROMPT declaration (prompt content)
  if (params.prompt) {
    lines.push(`PROMPT "${escapeAgentfileString(params.prompt)}"`);
  }

  // CONFIG declarations
  for (const [key, value] of Object.entries(params.configValues)) {
    if (value !== undefined && value !== null && value !== "") {
      lines.push(`CONFIG ${key} = ${formatAgentfileValue(value)}`);
    }
  }

  // Token budget cap — quoted to match backend applyWorkerModel emission.
  if (
    params.tokenBudget != null &&
    Number.isFinite(params.tokenBudget) &&
    params.tokenBudget > 0
  ) {
    lines.push(`CONFIG token_budget = "${Math.floor(params.tokenBudget)}"`);
  }

  // Repository slug / branch
  if (params.repositorySlug) {
    lines.push(`REPO "${params.repositorySlug}"`);
  }
  if (params.branchName) {
    lines.push(`BRANCH "${params.branchName}"`);
  }

  if (params.skillSlugs && params.skillSlugs.length > 0) {
    lines.push(`SKILLS ${params.skillSlugs.join(", ")}`);
  }

  if (params.knowledgeMounts && params.knowledgeMounts.length > 0) {
    const refs = params.knowledgeMounts.map((m) =>
      m.mode === "rw" ? `${m.slug} [rw]` : m.slug,
    );
    lines.push(`KNOWLEDGE ${refs.join(", ")}`);
  }

  return lines.join("\n");
}
