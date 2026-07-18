import type { AgentActivityItem, AgentTimelineItem } from "./contracts";

export type AgentToolActivityItem = AgentActivityItem & { kind: "tool" };

export interface ToolActivityRun {
  id: string;
  kind: "tool-run";
  tools: AgentToolActivityItem[];
}

export type GroupedTimelineItem = AgentTimelineItem | ToolActivityRun;

export function groupToolActivity(
  items: AgentTimelineItem[],
): GroupedTimelineItem[] {
  const grouped: GroupedTimelineItem[] = [];
  let tools: AgentToolActivityItem[] = [];

  const flushTools = () => {
    if (tools.length === 0) return;
    grouped.push({
      id: `tool-run:${tools[0].id}`,
      kind: "tool-run",
      tools,
    });
    tools = [];
  };

  for (const item of items) {
    if (isToolActivity(item)) {
      tools.push(item);
      continue;
    }
    flushTools();
    grouped.push(item);
  }
  flushTools();

  return grouped;
}

function isToolActivity(
  item: AgentTimelineItem,
): item is AgentToolActivityItem {
  return item.kind === "tool";
}
