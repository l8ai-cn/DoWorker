import { displayAgentName } from "./agent-slugs";

const VENDOR: Record<string, string> = {
  codex: "OpenAI",
  claude: "Anthropic",
  gemini: "Google",
  cursor: "Cursor",
  opencode: "OpenCode",
  aider: "Aider",
};

const AVATAR: Record<string, string> = {
  codex: "🅾️",
  claude: "🅰️",
  gemini: "✴️",
  cursor: "🅲",
  opencode: "◻️",
  aider: "🔧",
};

export interface AgentPickerOption {
  id: string;
  name: string;
  vendor: string;
  avatar: string;
  desc: string;
}

function harnessFamily(harness: string | null | undefined): string {
  const h = (harness ?? "").toLowerCase();
  if (h.includes("codex")) return "codex";
  if (h.includes("claude")) return "claude";
  if (h.includes("gemini")) return "gemini";
  if (h.includes("cursor")) return "cursor";
  if (h.includes("opencode")) return "opencode";
  if (h.includes("aider")) return "aider";
  return "codex";
}

export function agentPickerOption(
  id: string,
  name: string,
  harness?: string | null,
): AgentPickerOption {
  const family = harnessFamily(harness ?? name);
  const label = displayAgentName(name) || name;
  return {
    id,
    name: label,
    vendor: VENDOR[family] ?? "Agent",
    avatar: AVATAR[family] ?? "🤖",
    desc: `${label} · ${harness ?? id}`,
  };
}
