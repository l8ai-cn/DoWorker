import {
  Download,
  ExternalLink,
  File,
  FileAudio,
  FileCode2,
  FileDown,
  FileImage,
  FileText,
  FileVideo,
  Loader2,
  Presentation,
} from "lucide-react";
import { useState } from "react";

import { artifactPresentation, type ArtifactKind } from "./artifactPresentation";
import { artifactActionAllowed } from "./artifactGrantActions";
import type { AgentArtifactItem, AgentSessionRuntime } from "./contracts";
import { ArtifactPreview } from "./ArtifactPreview";
import { useAgentWorkspaceText } from "./AgentWorkspaceLocaleContext";
import { useArtifactBlobUrl } from "./useArtifactBlobUrl";

export function GenericArtifactCard({
  filename,
  item,
  runtime,
  sessionId,
}: {
  filename: string;
  item: AgentArtifactItem;
  runtime: AgentSessionRuntime;
  sessionId: string;
}) {
  const text = useAgentWorkspaceText();
  const declaredType = artifactPresentation(item.mimeType, filename);
  const [requestedArtifactId, setRequestedArtifactId] = useState<string | null>(
    null,
  );
  const [attempt, setAttempt] = useState(0);
  const [sourceDownloadError, setSourceDownloadError] = useState<string | null>(
    null,
  );
  const [sourceDownloading, setSourceDownloading] = useState(false);
  const enabled =
    !["audio", "video"].includes(declaredType.kind) ||
    requestedArtifactId === item.artifactId;
  const state = useArtifactBlobUrl(item, runtime, sessionId, enabled, attempt);
  const selectedRepresentation = item.representations.find(
    (representation) =>
      representation.representationId === item.selectedRepresentationId,
  );
  const sourceRepresentation = item.representations.find(
    (representation) =>
      representation.representationId === "original" &&
      representation.status === "ready",
  );
  const loadedFilename = selectedRepresentation?.filename || filename;
  const sourceFilename = sourceRepresentation?.filename || filename;
  const canDownloadSource =
    Boolean(runtime.loadArtifact && sourceRepresentation) &&
    artifactActionAllowed(
      item,
      "artifact.download",
      sourceRepresentation?.representationId,
    ) &&
    sourceRepresentation?.representationId !== item.selectedRepresentationId;
  const type =
    state.status === "ready"
      ? artifactPresentation(state.mimeType, filename)
      : declaredType;
  const load = () => {
    setRequestedArtifactId(item.artifactId);
    setAttempt((value) => value + 1);
  };
  const downloadSource = async () => {
    if (!runtime.loadArtifact || !sourceRepresentation) return;
    setSourceDownloadError(null);
    setSourceDownloading(true);
    try {
      const blob = await runtime.loadArtifact(
        sessionId,
        item.artifactId,
        sourceRepresentation.representationId,
      );
      const url = URL.createObjectURL(blob);
      const link = document.createElement("a");
      link.download = sourceFilename;
      link.href = url;
      link.click();
      setTimeout(() => URL.revokeObjectURL(url), 0);
    } catch (cause) {
      console.error("Artifact source download failed", cause);
      setSourceDownloadError(
        cause instanceof Error ? cause.message : String(cause),
      );
    } finally {
      setSourceDownloading(false);
    }
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
        {canDownloadSource && (
          <button
            aria-label={text.artifact.download(sourceFilename)}
            className="inline-flex size-11 shrink-0 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:opacity-50"
            disabled={sourceDownloading}
            onClick={() => void downloadSource()}
            type="button"
          >
            {sourceDownloading ? (
              <Loader2 className="size-4 animate-spin" />
            ) : (
              <FileDown className="size-4" />
            )}
          </button>
        )}
        {state.status === "ready" && (
          <a
            aria-label={text.artifact.download(loadedFilename)}
            className="inline-flex size-11 shrink-0 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
            download={loadedFilename}
            href={state.url}
          >
            <Download className="size-4" />
          </a>
        )}
      </div>
      {sourceDownloadError && (
        <div className="border-t border-destructive/20 px-3 py-2 text-xs text-destructive" role="alert">
          {sourceDownloadError}
        </div>
      )}
    </article>
  );
}

export function ArtifactError({
  filename,
  message,
}: {
  filename: string;
  message: string;
}) {
  return (
    <article
      className="rounded-md border border-destructive/30 bg-destructive/5 px-3 py-3 text-sm"
      role="alert"
    >
      <div className="truncate font-medium">{filename}</div>
      <div className="text-xs text-destructive">{message}</div>
    </article>
  );
}

function ArtifactTypeIcon({ kind }: { kind: ArtifactKind }) {
  const Icon =
    kind === "audio"
      ? FileAudio
      : kind === "image"
        ? FileImage
        : kind === "video"
          ? FileVideo
          : kind === "presentation"
            ? Presentation
            : kind === "code" || kind === "csv" || kind === "html"
              ? FileCode2
              : kind === "markdown" || kind === "pdf" || kind === "text"
                ? FileText
                : File;
  return <Icon className="size-5 shrink-0 text-muted-foreground" />;
}
