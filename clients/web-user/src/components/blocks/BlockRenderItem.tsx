import type { ReactNode } from "react";
import type { RenderItem } from "@/lib/renderItems";
import { cn } from "@/lib/utils";
import { ElicitationCard } from "./ApprovalCard";
import { FilePathAwareMessageResponse } from "./FilePathAwareMessageResponse";
import { OutputFileArtifact } from "./OutputFileArtifact";
import { ReasoningView } from "./ReasoningView";
import { SlashCommandCard } from "./SlashCommandCard";
import { SmartRoutingCard } from "./SmartRoutingCard";
import { ErrorBanner, PolicyDeniedBanner, RetryIndicator } from "./StatusBlocks";
import { TerminalCommandCard } from "./TerminalCommandCard";
import { ToolCard } from "./ToolCard";
import { isSmartRoutingTool } from "./ToolCardClassification";

export function renderBlockItem(
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
      if (isSmartRoutingTool(item)) {
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
      return <OutputFileArtifact key={key} {...item} />;
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
