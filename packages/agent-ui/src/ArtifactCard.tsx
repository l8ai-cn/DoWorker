import {
  Download,
  ExternalLink,
  File,
  FileCode2,
  FileImage,
  FileText,
  FileVideo,
  Presentation,
} from "lucide-react";
import { useState } from "react";

import type { AgentArtifactItem, AgentSessionRuntime } from "./contracts";
import {
  artifactPresentation,
  type ArtifactKind,
} from "./artifactPresentation";
import { ArtifactPreview } from "./ArtifactPreview";
import { useAgentWorkspaceText } from "./AgentWorkspaceLocaleContext";
import { useArtifactBlobUrl } from "./useArtifactBlobUrl";

export interface ArtifactCardProps {
  item: AgentArtifactItem;
  runtime: AgentSessionRuntime;
  sessionId: string;
}

export function ArtifactCard({ item, runtime, sessionId }: ArtifactCardProps) {
  const text = useAgentWorkspaceText();
  const filename = item.filename.trim() || text.artifact.generatedArtifact;
  const declaredType = artifactPresentation(item.mimeType, filename);
  const [requestedArtifactId, setRequestedArtifactId] = useState<string | null>(
    null,
  );
  const [attempt, setAttempt] = useState(0);
  const enabled =
    declaredType.kind !== "video" || requestedArtifactId === item.artifactId;
  const state = useArtifactBlobUrl(item, runtime, sessionId, enabled, attempt);
  const type =
    state.status === "ready"
      ? artifactPresentation(state.mimeType, filename)
      : declaredType;
  const load = () => {
    setRequestedArtifactId(item.artifactId);
    setAttempt((value) => value + 1);
  };
  return (
    <article className="overflow-hidden rounded-md border border-border bg-card">
      <ArtifactPreview
        filename={filename}
        kind={type.kind}
        onLoad={load}
        onRetry={load}
        state={state}
      />
      <div className="flex min-w-0 items-center gap-3 px-3 py-2.5">
        <ArtifactTypeIcon kind={type.kind} />
        <div className="min-w-0 flex-1">
          <div className="truncate text-sm font-medium">{filename}</div>
          <div className="truncate text-xs text-muted-foreground">
            {text.artifactType(type.label, type.label)}
          </div>
        </div>
        {state.status === "ready" &&
          type.kind !== "file" &&
          type.kind !== "html" && (
            <a
              aria-label={text.artifact.open(filename)}
              className="inline-flex size-11 shrink-0 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
              href={state.url}
              rel="noreferrer"
              target="_blank"
            >
              <ExternalLink className="size-4" />
            </a>
          )}
        {state.status === "ready" && (
          <a
            aria-label={text.artifact.download(filename)}
            className="inline-flex size-11 shrink-0 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
            download={filename}
            href={state.url}
          >
            <Download className="size-4" />
          </a>
        )}
      </div>
    </article>
  );
}

function ArtifactTypeIcon({ kind }: { kind: ArtifactKind }) {
  const Icon =
    kind === "image"
      ? FileImage
      : kind === "video"
        ? FileVideo
        : kind === "presentation"
          ? Presentation
          : kind === "code" || kind === "html"
            ? FileCode2
            : kind === "pdf" || kind === "text"
              ? FileText
              : File;
  return <Icon className="size-5 shrink-0 text-muted-foreground" />;
}
