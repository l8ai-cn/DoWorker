import {
  AlertCircle,
  Download,
  ExternalLink,
  File,
  FileCode2,
  FileImage,
  FileText,
  FileVideo,
  Loader2,
  Presentation,
} from "lucide-react";

import type {
  AgentArtifactItem,
  AgentSessionRuntime,
} from "./contracts";
import {
  artifactPresentation,
  type ArtifactKind,
} from "./artifactPresentation";
import { useAgentWorkspaceText } from "./AgentWorkspaceLocaleContext";
import {
  STATIC_HTML_REFERRER_POLICY,
  STATIC_HTML_SANDBOX,
  staticHtmlDocument,
} from "./security/staticHtmlProfile";
import { useArtifactBlobUrl } from "./useArtifactBlobUrl";

export interface ArtifactCardProps {
  item: AgentArtifactItem;
  runtime: AgentSessionRuntime;
  sessionId: string;
}

export function ArtifactCard({
  item,
  runtime,
  sessionId,
}: ArtifactCardProps) {
  const state = useArtifactBlobUrl(item, runtime, sessionId);
  const text = useAgentWorkspaceText();
  const filename = item.filename.trim() || text.generatedArtifact;

  if (state.status === "loading") {
    return (
      <article
        className="flex items-center gap-2 rounded-md border border-border bg-muted/30 px-3 py-3 text-sm text-muted-foreground"
        role="status"
      >
        <Loader2 className="size-4 animate-spin" />
        {text.loadingArtifact(filename)}
      </article>
    );
  }
  if (state.status === "error") {
    return <ArtifactError filename={filename} message={state.message} />;
  }

  const type = artifactPresentation(state.mimeType, filename);
  return (
    <article className="overflow-hidden rounded-md border border-border bg-card">
      {type.kind === "image" && (
        <img
          alt={filename}
          className="max-h-96 w-full bg-muted object-contain"
          src={state.url}
        />
      )}
      {type.kind === "video" && (
        <video
          aria-label={text.videoPreview(filename)}
          className="max-h-96 w-full bg-muted object-contain"
          controls
          playsInline
          preload="metadata"
          src={state.url}
        />
      )}
      {type.kind === "html" && (
        <iframe
          className="aspect-[16/10] min-h-80 w-full border-b border-border bg-white"
          referrerPolicy={STATIC_HTML_REFERRER_POLICY}
          sandbox={STATIC_HTML_SANDBOX}
          srcDoc={staticHtmlDocument(state.text ?? "")}
          title={text.previewArtifact(filename)}
        />
      )}
      {(type.kind === "code" || type.kind === "text") && state.text !== null && (
        <pre
          aria-label={text.previewArtifact(filename)}
          className="max-h-80 overflow-auto border-b border-border bg-muted/40 p-4 text-xs leading-5"
        >
          <code>{state.text}</code>
        </pre>
      )}
      <div className="flex min-w-0 items-center gap-3 px-3 py-2.5">
        <ArtifactTypeIcon kind={type.kind} />
        <div className="min-w-0 flex-1">
          <div className="truncate text-sm font-medium">{filename}</div>
          <div className="truncate text-xs text-muted-foreground">
            {text.artifactType(type.label, type.label)}
          </div>
        </div>
        {type.kind !== "file" && type.kind !== "html" && (
          <a
            aria-label={text.openArtifact(filename)}
            className="inline-flex size-11 shrink-0 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
            href={state.url}
            rel="noreferrer"
            target="_blank"
          >
            <ExternalLink className="size-4" />
          </a>
        )}
        <a
          aria-label={text.downloadArtifact(filename)}
          className="inline-flex size-11 shrink-0 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          download={filename}
          href={state.url}
        >
          <Download className="size-4" />
        </a>
      </div>
    </article>
  );
}

function ArtifactError({
  filename,
  message,
}: {
  filename: string;
  message: string;
}) {
  return (
    <article
      className="flex items-start gap-2 rounded-md border border-destructive/30 bg-destructive/5 px-3 py-3 text-sm"
      role="alert"
    >
      <AlertCircle className="mt-0.5 size-4 shrink-0 text-destructive" />
      <div className="min-w-0">
        <div className="truncate font-medium">{filename}</div>
        <div className="text-xs text-destructive">{message}</div>
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
