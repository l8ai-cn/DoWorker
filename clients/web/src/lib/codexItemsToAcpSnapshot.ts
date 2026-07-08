import { MsgType } from "@/stores/relayProtocol";

type ContentBlock = { type: string; text?: string };
type ConversationItem = Record<string, unknown> & { id: string; type: string };

export type AcpHistorySnapshot = {
  sessionId: string;
  state: "idle";
  messages: Array<{ text: string; role: string }>;
  toolCalls: Array<{
    toolCallId: string;
    toolName: string;
    status: string;
    argumentsJson: string;
    success?: boolean;
    resultText?: string;
    errorMessage?: string;
  }>;
};

function messageText(content: unknown): string {
  if (!Array.isArray(content)) return "";
  return (content as ContentBlock[])
    .map((block) => {
      if (block.type === "input_text" || block.type === "output_text") {
        return block.text ?? "";
      }
      if (block.type === "input_image" || block.type === "input_file") {
        return `[${block.type.replace("input_", "")}]`;
      }
      return "";
    })
    .filter((s) => s.length > 0)
    .join("\n");
}

/** Convert persisted conversation_items into an ACP snapshot for AcpActivityStream replay. */
export function codexItemsToAcpSnapshot(
  sessionId: string,
  items: ConversationItem[],
): AcpHistorySnapshot {
  const messages: AcpHistorySnapshot["messages"] = [];
  const toolCalls = new Map<string, AcpHistorySnapshot["toolCalls"][number]>();
  const outputs = new Map<string, string>();

  for (const item of items) {
    switch (item.type) {
      case "message": {
        if (item.is_meta === true) break;
        const role = item.role === "assistant" ? "assistant" : "user";
        const text = messageText(item.content);
        if (text.trim()) messages.push({ text, role });
        break;
      }
      case "function_call": {
        const callId = String(item.call_id ?? item.id);
        toolCalls.set(callId, {
          toolCallId: callId,
          toolName: String(item.name ?? "tool"),
          status: "completed",
          argumentsJson: String(item.arguments ?? "{}"),
        });
        break;
      }
      case "function_call_output": {
        outputs.set(String(item.call_id ?? ""), String(item.output ?? ""));
        break;
      }
      case "image_generation_call": {
        const callId = String(item.id);
        toolCalls.set(callId, {
          toolCallId: callId,
          toolName: "image_generation",
          status: "completed",
          argumentsJson: JSON.stringify({ prompt: item.prompt ?? "" }),
          success: true,
          resultText: "[image generated]",
        });
        break;
      }
      default:
        break;
    }
  }

  for (const [callId, output] of outputs) {
    const tc = toolCalls.get(callId);
    if (!tc) continue;
    tc.success = true;
    tc.resultText = output;
    tc.status = "completed";
  }

  return {
    sessionId,
    state: "idle",
    messages,
    toolCalls: [...toolCalls.values()],
  };
}

export const ACP_SNAPSHOT_MSG_TYPE = MsgType.AcpSnapshot;
