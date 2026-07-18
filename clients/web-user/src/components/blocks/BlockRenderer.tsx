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
import { cn } from "@/lib/utils";
import {
  useFileViewer,
  useFileViewerConversationId,
  useIsChangedPath,
  useWorkspacePaths,
} from "@/shell/FileViewerContext";
import { toWorkspaceRelativePath, useWorkspaceFileExists } from "@/hooks/useWorkspaceChangedFiles";
import { ElicitationCard } from "./ApprovalCard";
import { OutputFileArtifact } from "./OutputFileArtifact";
import { ReasoningView } from "./ReasoningView";
import { SlashCommandCard } from "./SlashCommandCard";
import { SmartRoutingCard } from "./SmartRoutingCard";
import { TerminalCommandCard } from "./TerminalCommandCard";
import { ErrorBanner, PolicyDeniedBanner, RetryIndicator } from "./StatusBlocks";
import { ToolCard, ToolGroupSummary } from "./ToolCard";

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
  return renderItem(fragment.tool, runStart + fragment.index, false);
}

const _ADVISE_MODELS_NAMES = new Set(["sys_advise_models", "mcp__omnigent__sys_advise_models"]);
const _SESSION_SEND_NAMES = new Set(["sys_session_send", "mcp__omnigent__sys_session_send"]);

function isPersistentToolCard(item: RenderItem): boolean {
  return (
    item.kind === "tool" &&
    (_ADVISE_MODELS_NAMES.has(item.execution.name) || _SESSION_SEND_NAMES.has(item.execution.name))
  );
}

function isToolItem(item: RenderItem): boolean {
  return item.kind === "tool" || item.kind === "native_tool";
}

/**
 * If the transcript ends in a contiguous tool run, return its start
 * index — that run is the live activity and should keep its
 * streaming tail. Otherwise return -1: the agent has spoken (or
 * reasoned) after the most recent tools, so they're no longer
 * "current".
 */
function findStreamingRunStart(items: RenderItem[]): number {
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

function renderItem(
  item: RenderItem,
  index: number,
  isReasoningStreaming: boolean,
  followsText = false,
): ReactNode {
  const key = keyFor(item, index);
  switch (item.kind) {
    case "text":
      return (
        <div
          key={key}
          data-testid="assistant-text-section"
          className={cn("min-w-0", followsText && "mt-2")}
        >
          <FilePathAwareMessageResponse>{item.text}</FilePathAwareMessageResponse>
        </div>
      );
    case "reasoning":
      return (
        <ReasoningView
          key={key}
          text={item.text}
          isStreaming={isReasoningStreaming}
          duration={item.duration}
        />
      );
    case "tool":
      // Intelligent routing's fan-out sizing gets a structured plan card
      // instead of the generic name(json) row + raw-JSON expansion.
      if (_ADVISE_MODELS_NAMES.has(item.execution.name)) {
        return (
          <SmartRoutingCard
            key={key}
            arguments={item.execution.arguments}
            output={item.output}
            state={item.state}
          />
        );
      }
      return (
        <ToolCard
          key={key}
          name={item.execution.name}
          argsSummary={item.execution.argsSummary}
          arguments={item.execution.arguments}
          output={item.output}
          state={item.state}
          startedAt={item.startedAt}
          duration={item.duration}
        />
      );
    case "native_tool":
      // Reuse the same tool card. Native tools are server-side
      // (provider-managed) so they're always "completed" by the
      // time we see them; render the raw provider data as input.
      return (
        <ToolCard
          key={key}
          name={item.label}
          nativeToolType={item.toolType}
          arguments={item.data}
          output={null}
          state="output-available"
        />
      );
    case "slash_command":
      return (
        <SlashCommandCard
          key={key}
          kind={item.slashKind}
          name={item.name}
          arguments={item.arguments}
          output={item.output}
        />
      );
    case "terminal_command":
      return (
        <TerminalCommandCard
          key={key}
          kind={item.terminalKind}
          input={item.input}
          stdout={item.stdout}
          stderr={item.stderr}
        />
      );
    case "file":
      return <OutputFileArtifact key={key} fileId={item.fileId} filename={item.filename} />;
    case "error":
      return <ErrorBanner key={key} message={item.message} source={item.source} code={item.code} />;
    case "policy_denied":
      return <PolicyDeniedBanner key={key} reason={item.reason} phase={item.phase} />;
    case "retry":
      return (
        <RetryIndicator
          key={key}
          source={item.source}
          attempt={item.attempt}
          maxAttempts={item.maxAttempts}
          delaySeconds={item.delaySeconds}
        />
      );
    case "elicitation":
      return <ElicitationCard key={key} item={item} />;
  }
}

/**
 * Stable key for each render item. Prefer the server-assigned item id;
 * fall back to call_id for tools (unique within a response) or to
 * position for pre-finalization fragments that don't carry an item id
 * yet (text/reasoning chunks emitted before their `output_item.done`).
 */
function keyFor(item: RenderItem, index: number): string {
  if (item.itemId) return `${item.kind}:${item.itemId}`;
  if (item.kind === "tool") return `tool:${item.execution.callId}`;
  if (item.kind === "elicitation") return `elicitation:${item.elicitationId}`;
  return `${item.kind}:${index}`;
}
