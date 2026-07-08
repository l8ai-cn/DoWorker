import type { LiveAgentEvent } from "./live-session-reducer";
import { dedupeRepeatedText } from "./text-normalizer";

interface ItemWire {
  id?: string;
  type?: string;
  role?: string;
  name?: string;
  arguments?: string;
  content?: Array<{ type: string; text?: string }>;
  status?: string;
}

function textFromContent(content: ItemWire["content"], prefer: string): string {
  if (!content) return "";
  const preferred = content
    .filter((c) => c.type === prefer && c.text)
    .map((c) => (c.text as string).trim())
    .filter(Boolean);
  if (preferred.length > 0) {
    const unique = [...new Set(preferred)];
    return dedupeRepeatedText(unique.join("\n\n"));
  }
  return content.find((c) => c.text)?.text ?? "";
}

function toolKindFor(name: string): LiveAgentEvent["toolKind"] {
  const n = name.toLowerCase();
  if (n.includes("bash") || n.includes("shell")) return "shell";
  if (n.includes("read")) return "read";
  if (n.includes("write")) return "write";
  if (n.includes("edit")) return "edit";
  if (n.includes("grep") || n.includes("glob")) return "search";
  if (n.includes("fetch") || n.includes("web")) return "fetch";
  return "other";
}

export function itemsToLiveEvents(items: ItemWire[]): LiveAgentEvent[] {
  const out: LiveAgentEvent[] = [];
  let seq = 0;
  const ts = () =>
    new Date().toLocaleTimeString("zh-CN", { hour: "2-digit", minute: "2-digit" });
  const id = (p: string) => {
    seq += 1;
    return `${p}-${seq}`;
  };

  for (const item of items) {
    if (item.type === "message") {
      if (item.role === "user") {
        const text = textFromContent(item.content, "input_text");
        if (text) {
          out.push({
            id: item.id ?? id("user"),
            type: "user_message",
            ts: ts(),
            title: "你",
            markdown: text,
          });
        }
      } else if (item.role === "assistant") {
        const text =
          textFromContent(item.content, "output_text") ||
          textFromContent(item.content, "text");
        if (text) {
          out.push({
            id: item.id ?? id("msg"),
            type: "agent_message",
            ts: ts(),
            title: "Agent",
            markdown: text,
            status: "completed",
          });
        }
      }
      continue;
    }
    if (item.type === "function_call") {
      out.push({
        id: item.id ?? id("tool"),
        type: "tool_call",
        ts: ts(),
        title: String(item.name ?? "tool"),
        tool: String(item.name ?? "tool"),
        toolKind: toolKindFor(String(item.name ?? "")),
        detail: String(item.arguments ?? ""),
        status: item.status === "completed" ? "completed" : "in_progress",
      });
    }
  }
  return out;
}
