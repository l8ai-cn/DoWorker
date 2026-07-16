import type { ReactNode } from "react";
import type { RenderItem } from "@/lib/renderItems";
import { renderBlockItem } from "./BlockRenderItem";
import { ToolGroupSummary } from "./ToolCard";
import { isPersistentToolCard } from "./ToolCardClassification";

const STREAMING_TAIL = 3;

type ToolRunFragment =
  | {
      kind: "group";
      tools: RenderItem[];
      count?: number;
    }
  | {
      kind: "standalone";
      tool: RenderItem;
      index: number;
    };

export function renderBlockToolRun(
  run: RenderItem[],
  runStart: number,
  isStreamingRun: boolean,
): ReactNode[] {
  const fragments = partitionToolRun(run, isStreamingRun);

  if (isStreamingRun && fragments[0]?.kind === "group") {
    const [group, ...tail] = fragments;
    // Wrap (group + trailing tail) in a single MessageContent child
    // so the message column's `gap-2` only applies AROUND this
    // pair, not BETWEEN them — the tail's `peer-data-[state=open]:mt-0`
    // can then truly bring the two bordered blocks flush when the
    // group is expanded.
    return [
      <div key={`tool-group-with-tail:${runStart}`}>
        <ToolGroupSummary tools={group.tools} count={group.count} />
        {tail.length > 0 && (
          <div className="mt-1 ml-2 space-y-1 border-l pl-3 py-1 peer-data-[state=open]:mt-0">
            {tail.map((fragment, idx) => renderToolRunFragment(fragment, runStart, idx))}
          </div>
        )}
      </div>,
    ];
  }

  return fragments.map((fragment, idx) => renderToolRunFragment(fragment, runStart, idx));
}

/**
 * Split a contiguous tool run into the part that folds into the
 * "See N steps" group versus the part rendered individually.
 *
 * For the live-streaming run, the trailing `STREAMING_TAIL` tools
 * (regardless of state) stay outside the group so the user can watch
 * the most recent activity. For any other run — older runs in the
 * transcript, or any run once the loop is idle — only still-in-progress
 * tools and persistent routing plan cards stay outside; everything else
 * folds.
 */
function partitionToolRun(run: RenderItem[], isStreamingRun: boolean): ToolRunFragment[] {
  if (isStreamingRun) {
    const tailStart = Math.max(0, run.length - STREAMING_TAIL);
    const fragments: ToolRunFragment[] = [];
    let group: RenderItem[] = [];
    // The "See N steps" count reflects the WHOLE run (folded head + visible
    // tail), so a folded head doesn't read "See 2 steps" with more visible.
    const flushGroup = () => {
      if (group.length === 0) return;
      fragments.push({ kind: "group", tools: group, count: run.length });
      group = [];
    };
    for (let index = 0; index < run.length; index += 1) {
      const item = run[index]!;
      // The trailing tail is the live edge; in-progress spinners and durable
      // routing/fan-out cards never fold (mirrors the idle branch below), so a
      // routing judgement isn't swallowed mid-fan-out when later spawns push
      // it past the tail window.
      if (index >= tailStart || isInProgressTool(item) || isPersistentToolCard(item)) {
        flushGroup();
        fragments.push({ kind: "standalone", tool: item, index });
      } else {
        group.push(item);
      }
    }
    flushGroup();
    return fragments;
  }

  const fragments: ToolRunFragment[] = [];
  let group: RenderItem[] = [];
  const flushGroup = () => {
    if (group.length === 0) return;
    fragments.push({ kind: "group", tools: group });
    group = [];
  };

  for (let index = 0; index < run.length; index += 1) {
    const item = run[index]!;
    if (isInProgressTool(item) || isPersistentToolCard(item)) {
      flushGroup();
      fragments.push({ kind: "standalone", tool: item, index });
    } else {
      group.push(item);
    }
  }
  flushGroup();
  return fragments;
}

function renderToolRunFragment(
  fragment: ToolRunFragment,
  runStart: number,
  fragmentIndex: number,
): ReactNode {
  if (fragment.kind === "group") {
    return (
      <ToolGroupSummary
        key={`tool-group:${runStart}:${fragmentIndex}`}
        tools={fragment.tools}
        count={fragment.count}
      />
    );
  }
  return renderBlockItem(fragment.tool, runStart + fragment.index, false);
}

export function isToolItem(item: RenderItem): boolean {
  return item.kind === "tool" || item.kind === "native_tool";
}

/**
 * If the transcript ends in a contiguous tool run, return its start
 * index — that run is the live activity and should keep its
 * streaming tail. Otherwise return -1: the agent has spoken (or
 * reasoned) after the most recent tools, so they're no longer
 * "current".
 */
export function findStreamingRunStart(items: RenderItem[]): number {
  if (items.length === 0) return -1;
  if (!isToolItem(items[items.length - 1]!)) return -1;
  let i = items.length - 1;
  while (i > 0 && isToolItem(items[i - 1]!)) i -= 1;
  return i;
}

/**
 * A tool item is in-progress only when it's a `tool` (not a
 * `native_tool` — those are provider-managed and always arrive
 * completed) and its derived UI state is `input-available`.
 */
function isInProgressTool(item: RenderItem): boolean {
  return item.kind === "tool" && item.state === "input-available";
}
