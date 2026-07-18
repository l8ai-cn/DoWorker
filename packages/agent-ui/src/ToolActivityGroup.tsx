import {
  AlertTriangle,
  ChevronRight,
  Loader2,
} from "lucide-react";

import { useAgentWorkspaceText } from "./AgentWorkspaceLocaleContext";
import { ToolActivityCard } from "./ToolActivityCard";
import type { AgentToolRendererRegistration } from "./react/rendererTypes";
import type { ToolRendererRegistry } from "./registry/ToolRendererRegistry";
import type { AgentToolActivityItem } from "./toolActivityGrouping";
import type { ToolActivityCount } from "./toolActivityGroupText";
import { resolveToolActivityPresentation } from "./toolActivityPresentation";

export function ToolActivityGroup({
  renderers,
  tools,
}: {
  renderers?: ToolRendererRegistry<AgentToolRendererRegistration>;
  tools: AgentToolActivityItem[];
}) {
  const text = useAgentWorkspaceText();
  const counts = countToolKinds(tools, renderers);
  const summary = text.toolActivityGroupSummary(counts);
  const status = groupStatus(tools);
  const Icon = resolveToolActivityPresentation(
    tools[0]!,
    renderers?.lookup(tools[0]!.identity),
  ).icon;

  return (
    <details className="group/tool-run">
      <summary className="flex min-h-8 cursor-pointer list-none items-center gap-2 rounded-sm px-1 text-xs text-muted-foreground hover:bg-muted/40">
        <ChevronRight className="size-3.5 shrink-0 transition-transform group-open/tool-run:rotate-90" />
        <Icon className="size-3.5 shrink-0" />
        <span className="min-w-0 flex-1">{summary}</span>
        <GroupStatus status={status} />
      </summary>
      <div className="ml-2 mt-1 space-y-2 border-l border-border pl-3">
        {tools.map((tool) => (
          <ToolActivityCard
            item={tool}
            key={tool.id}
            renderers={renderers}
          />
        ))}
      </div>
    </details>
  );
}

function GroupStatus({
  status,
}: {
  status: ToolGroupStatus;
}) {
  if (!status.failed && !status.running) return null;
  return (
    <span className="flex shrink-0 items-center gap-2">
      {status.failed && <StatusLabel status="failed" />}
      {status.running && <StatusLabel status="running" />}
    </span>
  );
}

function StatusLabel({ status }: { status: "running" | "failed" }) {
  const text = useAgentWorkspaceText();
  const Icon = status === "running" ? Loader2 : AlertTriangle;
  return (
    <span className="flex items-center gap-1">
      <Icon
        className={
          status === "running"
            ? "size-3.5 animate-spin text-primary"
            : "size-3.5 text-destructive"
        }
      />
      {text.activityStatus(status)}
    </span>
  );
}

function countToolKinds(
  tools: AgentToolActivityItem[],
  renderers?: ToolRendererRegistry<AgentToolRendererRegistration>,
) {
  const counts = new Map<string, ToolActivityCount>();
  for (const tool of tools) {
    const label = resolveToolActivityPresentation(
      tool,
      renderers?.lookup(tool.identity),
    ).label;
    const count = counts.get(label);
    counts.set(label, {
      label,
      count: (count?.count ?? 0) + 1,
    });
  }
  return [...counts.values()];
}

function groupStatus(tools: AgentToolActivityItem[]) {
  return {
    failed: tools.some((tool) => tool.status === "failed"),
    running: tools.some(
      (tool) => tool.status === "running" || tool.status === "pending",
    ),
  };
}

type ToolGroupStatus = ReturnType<typeof groupStatus>;
