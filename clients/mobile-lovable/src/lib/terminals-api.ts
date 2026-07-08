import { apiFetch } from "./api-fetch";
import { resolveWebSocketUrl } from "./ws-url";

export interface TerminalInfo {
  id: string;
  name: string;
  session: string;
  running: boolean;
}

const AGENT_TERMINAL_IDS = new Set([
  "terminal_tui_main",
  "terminal_claude_main",
  "terminal_codex_main",
  "terminal_opencode_main",
  "terminal_pi_main",
  "terminal_cursor_main",
  "terminal_kiro_main",
  "terminal_goose_main",
  "terminal_qwen_main",
  "terminal_antigravity_main",
  "terminal_kimi_main",
  "terminal_hermes_main",
]);

const SOFT_STATUSES = new Set([404, 409, 502, 503]);

function terminalFromResource(resource: Record<string, unknown>): TerminalInfo | null {
  const id = resource.id;
  if (typeof id !== "string" || !id) return null;
  const rawMeta = resource.metadata;
  const meta =
    rawMeta && typeof rawMeta === "object" && !Array.isArray(rawMeta)
      ? (rawMeta as Record<string, unknown>)
      : {};
  const terminalName = meta.terminal_name;
  const sessionKey = meta.session_key;
  const running = meta.running;
  const fallbackName = resource.name;
  return {
    id,
    name:
      typeof terminalName === "string" && terminalName
        ? terminalName
        : typeof fallbackName === "string"
          ? fallbackName
          : "",
    session: typeof sessionKey === "string" ? sessionKey : "",
    running: typeof running === "boolean" ? running : false,
  };
}

export async function listSessionTerminals(sessionId: string): Promise<TerminalInfo[]> {
  const res = await apiFetch(
    `/v1/sessions/${encodeURIComponent(sessionId)}/resources/terminals?order=asc&limit=100`,
  );
  if (SOFT_STATUSES.has(res.status)) return [];
  if (!res.ok) throw new Error(`terminals ${res.status}`);
  const json = (await res.json()) as { data?: unknown };
  const rows = Array.isArray(json.data) ? json.data : [];
  const out: TerminalInfo[] = [];
  for (const row of rows) {
    if (row && typeof row === "object") {
      const info = terminalFromResource(row as Record<string, unknown>);
      if (info) out.push(info);
    }
  }
  return out;
}

export function pickAgentTerminal(terminals: TerminalInfo[]): TerminalInfo | null {
  const agent = terminals.find((t) => AGENT_TERMINAL_IDS.has(t.id));
  if (agent) return agent;
  return terminals[0] ?? null;
}

export function buildTerminalAttachUrl(
  sessionId: string,
  terminalId: string,
  readOnly = false,
): string {
  const path =
    `/v1/sessions/${encodeURIComponent(sessionId)}` +
    `/resources/terminals/${encodeURIComponent(terminalId)}/attach` +
    (readOnly ? "?read_only=true" : "");
  return resolveWebSocketUrl(path);
}
