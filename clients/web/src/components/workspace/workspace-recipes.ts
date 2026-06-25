export interface WorkspaceRecipe {
  id: string;
  emoji: string;
  agentSlug: string;
  agentLabel: string;
}

export interface WorkspaceRecipeSelection {
  agentSlug: string;
  prompt: string;
}

// Workspace empty-state starter cards. `id` keys into the
// `workspace.recipes.<id>` i18n namespace (title/description/duration/prompt);
// `agentSlug` is the preferred agent — pre-selected only when the active
// runner exposes it, otherwise the form falls back to manual selection.
export const WORKSPACE_RECIPES: WorkspaceRecipe[] = [
  { id: "explain", emoji: "🔍", agentSlug: "claude-code", agentLabel: "claude-code" },
  { id: "tests", emoji: "🧪", agentSlug: "claude-code", agentLabel: "claude-code" },
  { id: "bug", emoji: "🐛", agentSlug: "codex", agentLabel: "codex · acp" },
];
