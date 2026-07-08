import { dedupeRepeatedText } from "./text-normalizer";
import { reconcileLiveEvents } from "./live-session-merge";

export type LiveEventType =
  | "user_message"
  | "agent_message"
  | "agent_thought"
  | "tool_call"
  | "permission_request"
  | "error";

export interface LiveAgentEvent {
  id: string;
  type: LiveEventType;
  ts: string;
  title: string;
  detail?: string;
  markdown?: string;
  tool?: string;
  toolKind?: "read" | "write" | "edit" | "shell" | "search" | "fetch" | "other";
  status?: "pending" | "in_progress" | "completed" | "failed";
  elicitationId?: string;
}

export class LiveSessionReducer {
  private events: LiveAgentEvent[] = [];
  private textBuf = "";
  private textEventId: string | null = null;
  private thoughtBuf = "";
  private thoughtId: string | null = null;
  private seq = 0;

  get snapshot(): LiveAgentEvent[] {
    return this.events.slice();
  }

  seed(events: LiveAgentEvent[]): void {
    this.events = events.slice();
    this.seq = events.length;
  }

  reconcile(persisted: LiveAgentEvent[]): void {
    const ephemeral = this.events.filter((e) => !e.id.startsWith("item_"));
    this.events = reconcileLiveEvents(persisted, ephemeral);
    this.seq = Math.max(this.seq, this.events.length);
  }

  appendUserMessage(text: string): void {
    if (this.events.some((e) => e.type === "user_message" && e.markdown === text)) return;
    this.push({ type: "user_message", title: "你", markdown: text });
  }

  apply(eventType: string, data: Record<string, unknown>): LiveAgentEvent[] {
    switch (eventType) {
      case "session.input.consumed":
        this.onInputConsumed(data);
        break;
      case "turn.text.delta":
      case "response.output_text.delta":
        this.onTextDelta(data);
        break;
      case "turn.item.done":
      case "response.output_item.done":
        this.onOutputItemDone(data);
        break;
      case "turn.reasoning.delta":
      case "response.reasoning_text.delta":
        this.onReasoningDelta(data);
        break;
      case "turn.elicitation.request":
      case "response.elicitation_request":
        this.onElicitationRequest(data);
        break;
      case "turn.elicitation.resolved":
      case "response.elicitation_resolved":
        this.onElicitationResolved(data);
        break;
      case "response.error":
        this.onError(data);
        break;
      default:
        break;
    }
    return this.snapshot;
  }

  private onInputConsumed(data: Record<string, unknown>) {
    const inner = data.data as Record<string, unknown> | undefined;
    const payload = (inner?.data ?? {}) as Record<string, unknown>;
    const content = payload.content as Array<{ type: string; text?: string }> | undefined;
    const text = content?.find((c) => c.type === "input_text")?.text ?? "";
    if (text) this.appendUserMessage(text);
  }

  private onTextDelta(data: Record<string, unknown>) {
    const delta = data.delta;
    if (typeof delta !== "string" || !delta) return;
    if (!this.textEventId) {
      this.textEventId = this.nextId("msg");
      this.textBuf = delta;
      this.events.push({
        id: this.textEventId,
        type: "agent_message",
        ts: nowTs(),
        title: "Agent",
        markdown: this.textBuf,
        status: "in_progress",
      });
    } else {
      this.textBuf += delta;
      this.patch(this.textEventId, { markdown: this.textBuf });
    }
  }

  private onOutputItemDone(data: Record<string, unknown>) {
    const item = data.item as Record<string, unknown> | undefined;
    if (!item) return;
    if (item.type === "message" && item.role === "assistant") this.finalizeText();
    if (item.type === "function_call") {
      this.finalizeText();
      this.push({
        type: "tool_call",
        title: String(item.name ?? "tool"),
        tool: String(item.name ?? "tool"),
        toolKind: toolKindFor(String(item.name ?? "")),
        status: "in_progress",
        detail: String(item.arguments ?? ""),
      });
    }
  }

  private onReasoningDelta(data: Record<string, unknown>) {
    const delta = data.delta;
    if (typeof delta !== "string" || !delta) return;
    if (!this.thoughtId) {
      this.thoughtId = this.nextId("think");
      this.thoughtBuf = delta;
      this.events.push({
        id: this.thoughtId,
        type: "agent_thought",
        ts: nowTs(),
        title: "思考中",
        markdown: this.thoughtBuf,
      });
    } else {
      this.thoughtBuf += delta;
      this.patch(this.thoughtId, { markdown: this.thoughtBuf });
    }
  }

  private onElicitationRequest(data: Record<string, unknown>) {
    this.finalizeText();
    const eid = String(data.elicitation_id ?? "");
    const params = (data.params ?? {}) as Record<string, unknown>;
    this.push({
      type: "permission_request",
      title: String(params.message ?? "需要审批"),
      detail: String(params.content_preview ?? ""),
      status: "pending",
      elicitationId: eid,
    });
  }

  private onElicitationResolved(data: Record<string, unknown>) {
    const eid = String(data.elicitation_id ?? "");
    for (const ev of this.events) {
      if (ev.elicitationId === eid && ev.type === "permission_request") {
        ev.status = "completed";
      }
    }
  }

  private onError(data: Record<string, unknown>) {
    const err = data.error as Record<string, unknown> | undefined;
    this.push({
      type: "error",
      title: "执行出错",
      detail: String(err?.message ?? "未知错误"),
      status: "failed",
    });
  }

  private finalizeText() {
    if (this.textEventId) {
      const ev = this.events.find((e) => e.id === this.textEventId);
      const markdown = ev?.markdown ? dedupeRepeatedText(ev.markdown) : undefined;
      this.patch(this.textEventId, { status: "completed", ...(markdown ? { markdown } : {}) });
    }
    this.textEventId = null;
    this.textBuf = "";
    this.thoughtId = null;
    this.thoughtBuf = "";
  }

  private push(partial: Omit<LiveAgentEvent, "id" | "ts">) {
    this.events.push({ id: this.nextId("ev"), ts: nowTs(), ...partial });
  }

  private patch(id: string, patch: Partial<LiveAgentEvent>) {
    const idx = this.events.findIndex((e) => e.id === id);
    if (idx >= 0) this.events[idx] = { ...this.events[idx], ...patch };
  }

  private nextId(prefix: string): string {
    this.seq += 1;
    return `${prefix}-${this.seq}`;
  }
}

function nowTs(): string {
  return new Date().toLocaleTimeString("zh-CN", { hour: "2-digit", minute: "2-digit" });
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

export async function* parseSseStream(body: ReadableStream<Uint8Array>): AsyncGenerator<{
  event: string;
  data: Record<string, unknown>;
}> {
  const decoder = new TextDecoder();
  let buf = "";
  let currentEvent: string | null = null;
  const reader = body.getReader();
  try {
    while (true) {
      const { done, value } = await reader.read();
      if (done) break;
      buf += decoder.decode(value, { stream: true });
      while (buf.includes("\n")) {
        const idx = buf.indexOf("\n");
        let line = buf.slice(0, idx);
        buf = buf.slice(idx + 1);
        if (line.endsWith("\r")) line = line.slice(0, -1);
        if (line.startsWith("event: ")) {
          currentEvent = line.slice(7);
        } else if (line.startsWith("data: ")) {
          const dataStr = line.slice(6);
          if (dataStr.trim() === "[DONE]") return;
          if (currentEvent) {
            try {
              yield { event: currentEvent, data: JSON.parse(dataStr) as Record<string, unknown> };
            } catch {
              /* skip malformed */
            }
            currentEvent = null;
          }
        } else if (line === "") {
          currentEvent = null;
        }
      }
    }
  } finally {
    reader.cancel().catch(() => {});
  }
}
