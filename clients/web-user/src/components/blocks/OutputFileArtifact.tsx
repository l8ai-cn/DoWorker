import { useState } from "react";
import { AlertCircleIcon, DownloadIcon, FileTextIcon, VideoIcon } from "lucide-react";
import { useFileViewerConversationId } from "@/shell/FileViewerContext";

interface OutputFileArtifactProps {
  fileId: string;
  filename: string | null;
  contentType: string | null;
}

type VideoState = "loading" | "ready" | "load-error" | "playback-error";

export function OutputFileArtifact({
  fileId,
  filename,
  contentType,
}: OutputFileArtifactProps) {
  const sessionId = useFileViewerConversationId();
  const label = filename?.trim() || "Generated file";
  const path = sessionId ? sessionFileContentPath(sessionId, fileId) : null;

  if (!path) {
    return <ArtifactError title="File unavailable" detail="The session file cannot be loaded." />;
  }

  if (isMp4Artifact(filename, contentType)) {
    return <VideoArtifact path={path} filename={label} />;
  }

  return (
    <div className="flex min-w-0 items-center gap-3 rounded-md border border-border bg-muted/40 p-3">
      <FileTextIcon className="size-5 shrink-0 text-muted-foreground" aria-hidden="true" />
      <span className="min-w-0 flex-1 truncate text-sm font-medium">{label}</span>
      <a
        aria-label={`Download ${label}`}
        className="inline-flex size-11 shrink-0 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-muted hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
        download={filename || "generated-file"}
        href={path}
      >
        <DownloadIcon className="size-4" aria-hidden="true" />
      </a>
    </div>
  );
}

function VideoArtifact({ path, filename }: { path: string; filename: string }) {
  const [state, setState] = useState<VideoState>("loading");

  if (state === "load-error") {
    return (
      <ArtifactError title="Video could not be loaded" detail={filename} downloadPath={path} />
    );
  }
  if (state === "playback-error") {
    return <ArtifactError title="Video playback failed" detail={filename} downloadPath={path} />;
  }

  return (
    <figure className="min-w-0 overflow-hidden rounded-md border border-border bg-muted/40">
      <video
        aria-label={filename}
        className="aspect-video w-full bg-black object-contain"
        controls
        onError={() =>
          setState((current) => (current === "ready" ? "playback-error" : "load-error"))
        }
        onLoadedData={() => setState("ready")}
        playsInline
        preload="metadata"
        src={path}
      />
      <figcaption className="flex min-w-0 items-center gap-2 px-3 py-2">
        <VideoIcon className="size-4 shrink-0 text-muted-foreground" aria-hidden="true" />
        <span className="min-w-0 flex-1 truncate text-sm font-medium">{filename}</span>
        {state === "loading" && (
          <span className="shrink-0 text-xs text-muted-foreground" role="status">
            Loading preview
          </span>
        )}
        <a
          aria-label={`Download ${filename}`}
          className="inline-flex size-11 shrink-0 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-muted hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          download={filename}
          href={path}
        >
          <DownloadIcon className="size-4" aria-hidden="true" />
        </a>
      </figcaption>
    </figure>
  );
}

function ArtifactError({
  title,
  detail,
  downloadPath,
}: {
  title: string;
  detail: string;
  downloadPath?: string;
}) {
  return (
    <div
      className="flex min-w-0 items-center gap-3 rounded-md border border-destructive/30 bg-destructive/5 p-3 text-sm"
      role="alert"
    >
      <AlertCircleIcon className="size-5 shrink-0 text-destructive" aria-hidden="true" />
      <div className="min-w-0 flex-1">
        <p className="font-medium text-foreground">{title}</p>
        <p className="truncate text-xs text-muted-foreground">{detail}</p>
      </div>
      {downloadPath && (
        <a
          aria-label={`Download ${detail}`}
          className="inline-flex size-11 shrink-0 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-muted hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          download={detail}
          href={downloadPath}
        >
          <DownloadIcon className="size-4" aria-hidden="true" />
        </a>
      )}
    </div>
  );
}

function isMp4Artifact(filename: string | null, contentType: string | null): boolean {
  const mimeType = contentType?.split(";", 1)[0]?.trim().toLowerCase();
  return mimeType === "video/mp4" || !!filename && /\.mp4$/i.test(filename);
}

function sessionFileContentPath(sessionId: string, fileId: string): string {
  return `/v1/sessions/${encodeURIComponent(sessionId)}/resources/files/${encodeURIComponent(fileId)}/content`;
}
