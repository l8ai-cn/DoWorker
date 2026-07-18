import { useEffect, useState } from "react";
import {
  VideoArtifactViewer,
  artifactPresentation,
} from "@do-worker/agent-ui";
import { AlertCircleIcon, DownloadIcon, FileTextIcon, PlayIcon, VideoIcon } from "lucide-react";
import { useFileViewerConversationId } from "@/shell/FileViewerContext";
import { useSessionFileObjectUrl } from "./useSessionFileObjectUrl";

interface OutputFileArtifactProps {
  fileId: string;
  filename: string | null;
  contentType: string | null;
}

type VideoState = "loading" | "ready" | "load-error" | "playback-error";

export function OutputFileArtifact({ fileId, filename, contentType }: OutputFileArtifactProps) {
  const sessionId = useFileViewerConversationId();
  const label = filename?.trim() || "生成文件";
  const path = sessionId ? sessionFileContentPath(sessionId, fileId) : null;
  const file = useSessionFileObjectUrl(path);
  const [downloadRequested, setDownloadRequested] = useState(false);
  const knownVideo = isVideoArtifact(filename, contentType);
  const resolvedVideo =
    knownVideo || (file.status === "ready" && isVideoArtifact(filename, file.mimeType));

  useEffect(() => {
    if (!downloadRequested || file.status !== "ready") return;
    setDownloadRequested(false);
    if (!resolvedVideo) downloadObjectUrl(file.url, label);
  }, [downloadRequested, file, label, resolvedVideo]);

  if (!path) {
    return <ArtifactError title="文件不可用" detail="当前会话无法读取该文件。" />;
  }
  if (file.status === "error") {
    return (
      <ArtifactError
        title={knownVideo ? "视频加载失败" : "文件加载失败"}
        detail={label}
        onRetry={() => {
          setDownloadRequested(!knownVideo);
          file.load();
        }}
      />
    );
  }
  if (file.status === "ready" && resolvedVideo) {
    return (
      <VideoArtifact
        contentType={file.mimeType || contentType || "video/mp4"}
        path={file.url}
        filename={label}
      />
    );
  }

  const loading = file.status === "loading";
  const actionLabel = knownVideo ? `加载视频 ${label}` : `下载 ${label}`;
  const Icon = knownVideo ? VideoIcon : FileTextIcon;
  return (
    <div className="flex min-w-0 items-center gap-3 rounded-md border border-border bg-muted/40 p-3">
      <Icon className="size-5 shrink-0 text-muted-foreground" aria-hidden="true" />
      <span className="min-w-0 flex-1 truncate text-sm font-medium">{label}</span>
      {loading && (
        <span className="shrink-0 text-xs text-muted-foreground" role="status">
          加载中
        </span>
      )}
      <button
        aria-label={actionLabel}
        className="inline-flex size-11 shrink-0 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-muted hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:opacity-50"
        disabled={loading}
        onClick={() => {
          if (file.status === "ready") {
            downloadObjectUrl(file.url, label);
            return;
          }
          setDownloadRequested(!knownVideo);
          file.load();
        }}
        type="button"
      >
        {knownVideo ? (
          <PlayIcon className="size-4" aria-hidden="true" />
        ) : (
          <DownloadIcon className="size-4" aria-hidden="true" />
        )}
      </button>
    </div>
  );
}

function VideoArtifact({
  contentType,
  path,
  filename,
}: {
  contentType: string;
  path: string;
  filename: string;
}) {
  const [state, setState] = useState<VideoState>("loading");
  if (state === "load-error" || state === "playback-error") {
    return (
      <ArtifactError
        title={state === "load-error" ? "视频加载失败" : "视频播放失败"}
        detail={filename}
        downloadPath={path}
      />
    );
  }
  return (
    <div
      onErrorCapture={() =>
        setState((current) => (current === "ready" ? "playback-error" : "load-error"))
      }
      onLoadedDataCapture={() => setState("ready")}
    >
      <VideoArtifactViewer
        filename={filename}
        mimeType={contentType}
        onDownload={() => downloadObjectUrl(path, filename)}
        src={path}
        status="ready"
      />
    </div>
  );
}

function ArtifactError({
  title,
  detail,
  downloadPath,
  onRetry,
}: {
  title: string;
  detail: string;
  downloadPath?: string;
  onRetry?: () => void;
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
      {onRetry && (
        <button className="text-xs text-foreground underline" onClick={onRetry} type="button">
          重试
        </button>
      )}
      {downloadPath && (
        <a aria-label={`下载 ${detail}`} download={detail} href={downloadPath}>
          <DownloadIcon className="size-4" aria-hidden="true" />
        </a>
      )}
    </div>
  );
}

function isVideoArtifact(filename: string | null, contentType: string | null): boolean {
  return artifactPresentation(contentType, filename || "").kind === "video";
}

function downloadObjectUrl(url: string, filename: string): void {
  const link = document.createElement("a");
  link.href = url;
  link.download = filename;
  document.body.append(link);
  link.click();
  link.remove();
}

function sessionFileContentPath(sessionId: string, fileId: string): string {
  return `/v1/sessions/${encodeURIComponent(sessionId)}/resources/files/${encodeURIComponent(fileId)}/content`;
}
