import { AlertCircle, Loader2, Play, RotateCcw } from "lucide-react";
import { useEffect, useState } from "react";

import type { ArtifactKind } from "./artifactPresentation";
import { ArtifactAudioPreview } from "./viewers/audio/ArtifactAudioPreview";
import { ArtifactCsvPreview } from "./viewers/csv/ArtifactCsvPreview";
import { ArtifactMarkdownPreview } from "./viewers/markdown/ArtifactMarkdownPreview";
import { LazyArtifactPdfPreview } from "./viewers/pdf/LazyArtifactPdfPreview";
import { useAgentWorkspaceText } from "./AgentWorkspaceLocaleContext";
import {
  STATIC_HTML_REFERRER_POLICY,
  STATIC_HTML_SANDBOX,
  staticHtmlDocument,
} from "./security/staticHtmlProfile";
import type { ArtifactBlobState } from "./useArtifactBlobUrl";

interface ArtifactPreviewProps {
  filename: string;
  kind: ArtifactKind;
  state: ArtifactBlobState;
  onLoad: () => void;
  onRetry: () => void;
}

export function ArtifactPreview({
  filename,
  kind,
  state,
  onLoad,
  onRetry,
}: ArtifactPreviewProps) {
  const text = useAgentWorkspaceText().artifact;
  const [videoFailed, setVideoFailed] = useState(false);

  useEffect(() => {
    setVideoFailed(false);
  }, [state.status === "ready" ? state.url : state.status]);

  if (state.status === "idle") {
    return (
      <div className="flex min-h-36 items-center justify-center border-b border-border bg-muted/30 p-4">
        <button
          className="inline-flex min-h-11 items-center gap-2 rounded-md border border-border bg-background px-4 text-sm font-medium hover:bg-muted focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          onClick={onLoad}
          type="button"
        >
          <Play className="size-4" />
          {text.load(filename)}
        </button>
      </div>
    );
  }
  if (state.status === "loading") {
    return (
      <div
        className="flex min-h-36 items-center justify-center gap-2 border-b border-border bg-muted/30 p-4 text-sm text-muted-foreground"
        role="status"
      >
        <Loader2 className="size-4 animate-spin" />
        {text.loading(filename)}
      </div>
    );
  }
  if (state.status === "error") {
    const message =
      state.code === "generation_failed"
        ? text.generationFailed
        : state.code === "authorization_required"
          ? text.downloadNotAuthorized
        : state.code === "loading_unavailable"
          ? text.loadingUnavailable
          : text.loadFailed;
    return (
      <ArtifactPreviewError
        filename={filename}
        message={message}
        onRetry={state.retryable ? onRetry : undefined}
      />
    );
  }
  if (kind === "video" && videoFailed) {
    return (
      <ArtifactPreviewError
        filename={filename}
        message={text.videoPlaybackFailed(filename)}
        onRetry={onRetry}
      />
    );
  }
  if (kind === "audio") {
    return <ArtifactAudioPreview filename={filename} src={state.url} />;
  }
  if (kind === "image") {
    return (
      <img
        alt={filename}
        className="max-h-96 w-full bg-muted object-contain"
        src={state.url}
      />
    );
  }
  if (kind === "video") {
    return (
      <video
        aria-label={text.videoPreview(filename)}
        className="max-h-96 w-full bg-muted object-contain"
        controls
        onError={() => setVideoFailed(true)}
        playsInline
        preload="metadata"
        src={state.url}
      />
    );
  }
  if (kind === "html") {
    return (
      <iframe
        className="aspect-[16/10] min-h-80 w-full border-b border-border bg-white"
        referrerPolicy={STATIC_HTML_REFERRER_POLICY}
        sandbox={STATIC_HTML_SANDBOX}
        srcDoc={staticHtmlDocument(state.text ?? "")}
        title={text.preview(filename)}
      />
    );
  }
  if (kind === "markdown" && state.text !== null) {
    return (
      <ArtifactMarkdownPreview
        text={state.text}
        truncated={state.textTruncated}
      />
    );
  }
  if (kind === "csv" && state.text !== null) {
    return (
      <ArtifactCsvPreview
        filename={filename}
        text={state.text}
        truncated={state.textTruncated}
      />
    );
  }
  if (kind === "pdf") {
    return <LazyArtifactPdfPreview filename={filename} src={state.url} />;
  }
  if ((kind === "code" || kind === "text") && state.text !== null) {
    return (
      <pre
        aria-label={text.preview(filename)}
        className="max-h-80 overflow-auto border-b border-border bg-muted/40 p-4 text-xs leading-5"
      >
        <code>{state.text}</code>
      </pre>
    );
  }
  return null;
}

function ArtifactPreviewError({
  filename,
  message,
  onRetry,
}: {
  filename: string;
  message: string;
  onRetry?: () => void;
}) {
  const text = useAgentWorkspaceText().artifact;
  return (
    <div
      className="flex min-h-36 items-center justify-center gap-3 border-b border-destructive/30 bg-destructive/5 p-4 text-sm"
      role="alert"
    >
      <AlertCircle className="size-4 shrink-0 text-destructive" />
      <span className="text-destructive">{message}</span>
      {onRetry && (
        <button
          aria-label={text.retry(filename)}
          className="inline-flex size-11 shrink-0 items-center justify-center rounded-md hover:bg-destructive/10 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          onClick={onRetry}
          type="button"
        >
          <RotateCcw className="size-4" />
        </button>
      )}
    </div>
  );
}
