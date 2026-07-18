// Tool-call collapsing: within a contiguous run of tool / native_tool
// items, older tools fold into a single "See N steps" line (rendered
// by `ToolGroupSummary`). The trailing `STREAMING_TAIL` tools (any
// state) stay outside the group ONLY when (a) the session is still
// running and (b) the very last item in the transcript is a tool —
// meaning the agent hasn't produced any text/reasoning after this
// run yet, so these tools are the live activity. Once the agent
// emits anything else after a tool run (or once the session is
// idle), the run collapses except for still-in-progress spinners and
// durable routing/fan-out cards.

import type { ReactNode } from "react";
import type { RenderItem } from "@/lib/renderItems";
import type { SessionStatus } from "@/lib/types";
import { renderBlockItem } from "./BlockRenderItem";
import { findStreamingRunStart, isToolItem, renderBlockToolRun } from "./BlockToolRun";

export { FilePathAwareMessageResponse } from "./FilePathAwareMessageResponse";

interface BlockRendererProps {
  items: RenderItem[];
  sessionStatus: SessionStatus;
}

export function BlockRenderer({ items, sessionStatus }: BlockRendererProps) {
  const rendered: ReactNode[] = [];
  let previousRenderedItemWasText = false;
  const isAgentActive = sessionStatus === "running" || sessionStatus === "waiting";
  const streamingRunStart = isAgentActive ? findStreamingRunStart(items) : -1;
  // Reasoning is "currently streaming" iff the agent is live AND this
  // reasoning is the very last item in the bubble. Mirrors the
  // `streamingRunStart` rule for tool runs: the trailing live edge stays
  // expanded; once anything else lands after it, it collapses.
  const lastIdx = items.length - 1;
  const reasoningStreamingIdx =
    isAgentActive && lastIdx >= 0 && items[lastIdx]!.kind === "reasoning" ? lastIdx : -1;

  for (let i = 0; i < items.length; i += 1) {
    const item = items[i]!;

    if (isToolItem(item)) {
      // Consume contiguous run of tool / native_tool items.
      const runStart = i;
      while (i < items.length && isToolItem(items[i]!)) i += 1;
      const run = items.slice(runStart, i);
      i -= 1; // outer loop will i += 1

      // Only the run at `streamingRunStart` (when set) is treated as
      // "currently streaming". Earlier runs, and any run followed by
      // assistant text/reasoning, collapse the same way they would
      // when idle.
      const isStreamingRun = runStart === streamingRunStart;
      rendered.push(...renderBlockToolRun(run, runStart, isStreamingRun));
      previousRenderedItemWasText = false;
      continue;
    }

    const followsText = item.kind === "text" && previousRenderedItemWasText;
    rendered.push(renderBlockItem(item, i, i === reasoningStreamingIdx, followsText));
    previousRenderedItemWasText = item.kind === "text";
  }

  return <>{rendered}</>;
}
