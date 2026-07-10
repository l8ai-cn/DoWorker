/**
 * Utilities for generating AgentFile Layer source from form fields.
 * An AgentFile Layer is a DSL fragment that configures a Pod's environment.
 */

import { POD_MODE_PTY } from "@/lib/pod-modes";

/** POSIX-style env var name: uppercase letters, digits, underscores. */
const ENV_NAME_PATTERN = /^[A-Z_][A-Z0-9_]*$/;

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
  /** Runtime bundle names (kind='runtime') to attach in explicit selection order. */
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
  /**
   * Per-Worker custom environment variables. Emitted as `ENV KEY = "value"`
   * lines AFTER USE_ENV_BUNDLE so custom values win over bundle values on
   * conflicting keys (matches backend eval order). Entries with an empty or
   * malformed key are skipped — the form validates keys before submit.
   */
  customEnv?: { key: string; value: string }[];
}): string {
  const lines: string[] = [];

  // MODE declaration (if not default PTY)
  if (params.interactionMode && params.interactionMode !== POD_MODE_PTY) {
    lines.push(`MODE ${params.interactionMode}`);
  }

  // USE_ENV_BUNDLE declarations for explicit runtime preferences.
  if (params.runtimeBundleNames) {
    for (const name of params.runtimeBundleNames) {
      if (name) {
        lines.push(`USE_ENV_BUNDLE "${escapeAgentfileString(name)}"`);
      }
    }
  }
  if (params.configBundleNames) {
    for (const name of params.configBundleNames) {
      if (name) {
        lines.push(`USE_CONFIG_BUNDLE "${escapeAgentfileString(name)}"`);
      }
    }
  }

  // Custom ENV declarations — after bundles so they win on key conflicts.
  if (params.customEnv) {
    for (const { key, value } of params.customEnv) {
      const trimmedKey = key.trim();
      if (trimmedKey && ENV_NAME_PATTERN.test(trimmedKey)) {
        lines.push(`ENV ${trimmedKey} = ${formatAgentfileValue(value)}`);
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
