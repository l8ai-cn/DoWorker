const ENGINE_TO_AGENT: Record<string, string> = {
  codex: "codex-cli",
  "claude-code": "claude-code",
  "gemini-cli": "gemini-cli",
  cursor: "cursor-cli",
};

export function agentSlugForEngine(engineId: string): string {
  return ENGINE_TO_AGENT[engineId] ?? engineId;
}

export function displayAgentName(agentId: string): string {
  const labels: Record<string, string> = {
    "codex-cli": "Codex",
    "claude-code": "Claude Code",
    "gemini-cli": "Gemini CLI",
    "cursor-cli": "Cursor",
  };
  return labels[agentId] ?? agentId;
}
