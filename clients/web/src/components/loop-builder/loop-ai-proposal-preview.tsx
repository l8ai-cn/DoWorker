import type { LoopAIMessages } from "./loop-workbench-messages";

interface LoopAIProposalPreviewProps {
  currentSource: string;
  proposedSource: string;
  messages: LoopAIMessages;
}

export function LoopAIProposalPreview({
  currentSource,
  proposedSource,
  messages,
}: LoopAIProposalPreviewProps) {
  return (
    <div className="grid min-h-0 overflow-hidden rounded-md border border-border lg:grid-cols-2">
      <SourcePane label={messages.current} source={currentSource} />
      <SourcePane
        className="border-t border-border lg:border-l lg:border-t-0"
        label={messages.proposed}
        source={proposedSource}
      />
    </div>
  );
}

function SourcePane({
  className = "",
  label,
  source,
}: {
  className?: string;
  label: string;
  source: string;
}) {
  return (
    <section className={`min-w-0 ${className}`}>
      <h3 className="border-b border-border bg-surface-muted px-3 py-2 text-xs font-semibold">
        {label}
      </h3>
      <pre className="max-h-[52vh] overflow-auto whitespace-pre-wrap break-words bg-background p-3 font-mono text-xs leading-5">
        {source}
      </pre>
    </section>
  );
}
