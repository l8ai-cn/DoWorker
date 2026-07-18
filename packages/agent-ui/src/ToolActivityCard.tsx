import {
  AlertTriangle,
  CheckCircle2,
  ChevronRight,
  Loader2,
} from "lucide-react";

import { useAgentWorkspaceText } from "./AgentWorkspaceLocaleContext";
import type { AgentToolActivityItem } from "./agentToolContracts";
import type { AgentToolRendererRegistration } from "./react/rendererTypes";
import type { ToolRendererRegistry } from "./registry/ToolRendererRegistry";
import {
  cleanToolEvidence,
  toolActivityRawEvidence,
} from "./toolActivityEvidence";
import {
  resolveToolActivityPresentation,
  toolActivityIdentity,
} from "./toolActivityPresentation";

export function ToolActivityCard({
  item,
  renderers,
}: {
  item: AgentToolActivityItem;
  renderers?: ToolRendererRegistry<AgentToolRendererRegistration>;
}) {
  const text = useAgentWorkspaceText();
  const renderer = renderers?.lookup(item.identity);
  const presentation = resolveToolActivityPresentation(item, renderer);
  const input = cleanToolEvidence(item.input);
  const output = cleanToolEvidence(item.output);
  const detail = cleanToolEvidence(item.detail);
  const rawEvidence = presentation.specialized
    ? undefined
    : toolActivityRawEvidence(item);
  const hasEvidence = Boolean(input || output || detail || rawEvidence);
  const Icon = presentation.icon;
  const RegisteredSummary = renderer?.summary;
  const RegisteredDetail = renderer?.detail;

  return (
    <article className="overflow-hidden rounded-md border border-border bg-card">
      <div className="flex min-h-10 items-center gap-2 px-3">
        <Icon className="size-4 shrink-0 text-muted-foreground" />
        <span className="min-w-0 flex-1 truncate text-sm font-medium">
          {text.toolText(presentation.label)}
        </span>
        <ActivityStatus
          label={text.activityStatus(item.status)}
          status={item.status}
        />
      </div>
      {!presentation.specialized && (
        <div
          className="border-t border-border bg-muted/15 px-3 py-2 text-xs text-muted-foreground"
          data-testid="unsupported-tool-preview"
        >
          <div>{text.unsupportedToolPreview}</div>
          <code className="mt-1 block break-all text-[11px] text-foreground">
            {toolActivityIdentity(item)}
          </code>
        </div>
      )}
      {RegisteredSummary ? (
        <div
          className="border-t border-border bg-muted/15 px-3 py-2"
          data-testid="registered-tool-summary"
        >
          <RegisteredSummary item={item} />
        </div>
      ) : null}
      {(hasEvidence || RegisteredDetail) && (
        <details
          className="group border-t border-border"
          open={item.status === "failed"}
        >
          <summary className="flex h-8 cursor-pointer list-none items-center gap-1.5 px-3 text-xs text-muted-foreground hover:bg-muted/40">
            <ChevronRight className="size-3.5 transition-transform group-open:rotate-90" />
            {text.details}
          </summary>
          <div className="space-y-3 border-t border-border bg-muted/20 px-3 py-3">
            {RegisteredDetail && <RegisteredDetail item={item} />}
            {input && (
              <Evidence
                label={text.toolText(presentation.inputLabel)}
                value={input}
              />
            )}
            {output && (
              <Evidence
                label={text.toolText(presentation.outputLabel)}
                value={output}
              />
            )}
            {detail && (
              <Evidence label={text.details} value={detail} />
            )}
            {rawEvidence && (
              <Evidence label={text.rawToolEvidence} value={rawEvidence} />
            )}
          </div>
        </details>
      )}
    </article>
  );
}

function ActivityStatus({
  label,
  status,
}: {
  label: string;
  status: AgentToolActivityItem["status"];
}) {
  const Icon =
    status === "running"
      ? Loader2
      : status === "failed"
        ? AlertTriangle
        : CheckCircle2;
  return (
    <span className="flex shrink-0 items-center gap-1 text-xs text-muted-foreground">
      <Icon
        className={`size-3.5 ${
          status === "running"
            ? "animate-spin text-primary"
            : status === "failed"
              ? "text-destructive"
              : "text-emerald-600"
        }`}
      />
      {label}
    </span>
  );
}

function Evidence({ label, value }: { label: string; value: string }) {
  return (
    <section>
      <div className="mb-1 text-[11px] font-medium uppercase text-muted-foreground">
        {label}
      </div>
      <pre className="overflow-x-auto whitespace-pre-wrap font-mono text-xs leading-5">
        {value}
      </pre>
    </section>
  );
}
