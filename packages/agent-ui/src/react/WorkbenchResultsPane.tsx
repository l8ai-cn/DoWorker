import {
  File,
  FileCode2,
  FileImage,
  FileText,
  FileVideo,
  Presentation,
  Wrench,
} from "lucide-react";

import { artifactPresentation, type ArtifactKind } from "../artifactPresentation";
import { ArtifactCard } from "../ArtifactCard";
import { useAgentWorkspaceText } from "../AgentWorkspaceLocaleContext";
import type { AgentSessionRuntime } from "../contracts";
import type { ContentRendererRegistry } from "../registry/ContentRendererRegistry";
import type { AgentContentRendererRegistration } from "./contentRendererTypes";
import type { WorkbenchResult } from "./workbenchResults";

export interface WorkbenchResultsPaneProps {
  contentRenderers?: ContentRendererRegistry<AgentContentRendererRegistration>;
  onSelect: (id: string) => void;
  results: readonly WorkbenchResult[];
  runtime: AgentSessionRuntime;
  selected: WorkbenchResult | undefined;
  sessionId: string;
}

export function WorkbenchResultsPane({
  contentRenderers,
  onSelect,
  results,
  runtime,
  selected,
  sessionId,
}: WorkbenchResultsPaneProps) {
  const text = useAgentWorkspaceText();
  return (
    <div className="flex h-full min-h-0 flex-col bg-background">
      <header className="flex h-10 shrink-0 items-center border-b border-border px-3 text-xs font-medium">
        {text.results}
        <span className="ml-2 text-muted-foreground">{results.length}</span>
      </header>
      <div className="flex min-h-0 flex-1">
        <nav
          aria-label={text.results}
          className="flex w-14 shrink-0 flex-col items-center gap-1 overflow-y-auto border-r border-border py-2"
        >
          {results.map((result) => (
            <ResultTab
              active={result.id === selected?.id}
              key={result.id}
              onClick={() => onSelect(result.id)}
              result={result}
            />
          ))}
        </nav>
        <main className="min-h-0 min-w-0 flex-1 overflow-y-auto p-3">
          {selected?.kind === "artifact" ? (
            <ArtifactCard
              contentRenderers={contentRenderers}
              item={selected.item}
              runtime={runtime}
              sessionId={sessionId}
            />
          ) : selected?.kind === "tool" ? (
            <selected.Renderer
              item={selected.item}
              runtime={runtime}
              sessionId={sessionId}
            />
          ) : null}
        </main>
      </div>
    </div>
  );
}

function ResultTab({
  active,
  onClick,
  result,
}: {
  active: boolean;
  onClick: () => void;
  result: WorkbenchResult;
}) {
  const label =
    result.kind === "artifact" ? result.item.filename : result.item.title;
  const Icon =
    result.kind === "tool"
      ? Wrench
      : artifactIcon(
          artifactPresentation(
            result.item.mimeType,
            result.item.filename,
          ).kind,
        );
  return (
    <button
      aria-label={label}
      aria-pressed={active}
      className={`grid size-11 shrink-0 place-items-center rounded-md outline-none focus-visible:ring-2 focus-visible:ring-ring ${
        active
          ? "bg-muted text-foreground"
          : "text-muted-foreground hover:bg-muted/60"
      }`}
      onClick={onClick}
      title={label}
      type="button"
    >
      <Icon className="size-4" />
    </button>
  );
}

function artifactIcon(kind: ArtifactKind) {
  if (kind === "image") return FileImage;
  if (kind === "video") return FileVideo;
  if (kind === "presentation") return Presentation;
  if (kind === "code" || kind === "html") return FileCode2;
  if (kind === "pdf" || kind === "text") return FileText;
  return File;
}
