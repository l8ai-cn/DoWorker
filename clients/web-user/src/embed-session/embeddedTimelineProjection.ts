import type { AgentTimelineItem } from "@do-worker/agent-ui";

import type { RenderItem } from "@/lib/renderItems";
import type { ActiveResponse } from "@/store/types";

export function projectEmbeddedTimelineItem(
  item: Exclude<RenderItem, { kind: "elicitation" }>,
  id: string,
  lifecycle: ActiveResponse["state"],
): AgentTimelineItem {
  if (item.kind === "text") {
    return {
      id,
      kind: "message",
      role: "assistant",
      status: lifecycle === "failed" ? "failed" : item.final ? "completed" : "streaming",
      text: item.text,
    };
  }
  if (item.kind === "reasoning") {
    return activity(id, "reasoning", "Reasoning", item.text, item.text ? "running" : "pending");
  }
  if (item.kind === "file") {
    return {
      id,
      kind: "artifact",
      artifactId: item.fileId,
      filename: item.filename || "Generated artifact",
      mimeType: item.contentType,
      status: "completed",
    };
  }
  if (item.kind === "tool") {
    return activity(
      id,
      "tool",
      item.execution.name,
      undefined,
      item.state === "input-available"
        ? "running"
        : item.state === "output-error"
          ? "failed"
          : "completed",
      toolInput(item),
      item.output ?? undefined,
    );
  }
  if (item.kind === "error" || item.kind === "policy_denied") {
    return activity(
      id,
      "error",
      item.kind === "error" ? item.message || item.code : "Policy denied",
      item.kind === "error" ? item.source : item.reason,
      "failed",
    );
  }
  return activity(id, "system", activityTitle(item), activityDetail(item), "completed");
}

function activity(
  id: string,
  kind: "reasoning" | "tool" | "error" | "system",
  title: string,
  detail: string | undefined,
  status: "pending" | "running" | "completed" | "failed",
  input?: string,
  output?: string,
): AgentTimelineItem {
  return { id, kind, title, detail, input, output, status };
}

function toolInput(item: Extract<RenderItem, { kind: "tool" }>): string | undefined {
  const input = JSON.stringify(item.execution.arguments, null, 2);
  return input === "{}" ? undefined : input;
}

type OtherItem = Exclude<
  RenderItem,
  {
    kind:
      | "elicitation"
      | "text"
      | "reasoning"
      | "tool"
      | "error"
      | "policy_denied"
      | "file";
  }
>;

function activityTitle(item: OtherItem): string {
  if (item.kind === "native_tool") return item.label;
  if (item.kind === "slash_command") return `/${item.name}`;
  if (item.kind === "terminal_command") return "Terminal command";
  return `Retrying ${item.source}`;
}

function activityDetail(item: OtherItem): string | undefined {
  if (item.kind === "native_tool") return JSON.stringify(item.data, null, 2);
  if (item.kind === "slash_command") return item.output ?? item.arguments;
  if (item.kind === "terminal_command") {
    return item.input ?? [item.stdout, item.stderr].filter(Boolean).join("\n");
  }
  return `Attempt ${item.attempt} of ${item.maxAttempts}`;
}
