import { CircleAlert, Clock3, Download, LoaderCircle } from "lucide-react";
import { VideoPlaybackSurface } from "./VideoPlaybackSurface";

export type VideoArtifactStatus =
  | "queued"
  | "rendering"
  | "transcoding"
  | "ready"
  | "failed";
export interface VideoArtifactVersion {
  id: string;
  label?: string;
  src?: string;
  filename?: string;
  mimeType?: string;
  posterSrc?: string;
  durationSeconds?: number;
}
export interface VideoArtifactViewerProps {
  src: string;
  filename: string;
  mimeType: string;
  posterSrc?: string;
  durationSeconds?: number;
  status: VideoArtifactStatus;
  progress?: number;
  versions?: readonly VideoArtifactVersion[];
  selectedVersionId?: string;
  onSelectVersion?: (versionId: string) => void;
  onDownload?: () => void;
}
const STATUS_TEXT: Record<Exclude<VideoArtifactStatus, "ready">, string> = {
  queued: "视频已排队，等待生成",
  rendering: "正在渲染视频",
  transcoding: "正在转码视频",
  failed: "视频生成失败",
};

export function VideoArtifactViewer({
  src,
  filename,
  mimeType,
  posterSrc,
  durationSeconds,
  status,
  progress,
  versions = [],
  selectedVersionId,
  onSelectVersion,
  onDownload,
}: VideoArtifactViewerProps) {
  const selectedId = selectedVersionId ?? versions[0]?.id ?? "";
  const selectedVersion = versions.find((version) => version.id === selectedId);
  const activeFilename =
    selectedVersion?.filename?.trim() || filename.trim() || "生成视频";
  const activeSrc = selectedVersion?.src ?? src;
  const activeMimeType = selectedVersion?.mimeType ?? mimeType;
  const activePosterSrc = selectedVersion?.posterSrc ?? posterSrc;
  const activeDuration = selectedVersion?.durationSeconds ?? durationSeconds;
  const canDownload = status === "ready" && Boolean(onDownload);
  return (
    <article className="overflow-hidden rounded-md border border-border bg-card">
      {status === "ready" ? (
        <VideoPlaybackSurface
          filename={activeFilename}
          key={`${activeSrc}:${activeMimeType}`}
          posterSrc={activePosterSrc}
          src={activeSrc}
        />
      ) : (
        <VideoStatusPanel progress={progress} status={status} />
      )}
      <div className="flex min-w-0 flex-wrap items-center gap-3 border-t border-border px-3 py-2.5">
        <div className="min-w-40 flex-1">
          <div className="truncate text-sm font-medium" title={activeFilename}>
            {activeFilename}
          </div>
          <div className="flex flex-wrap gap-x-3 text-xs text-muted-foreground">
            <span>{activeMimeType}</span>
            {activeDuration !== undefined && (
              <span>时长 {formatDuration(activeDuration)}</span>
            )}
          </div>
        </div>

        {versions.length > 0 && (
          <label className="flex items-center gap-2 text-xs text-muted-foreground">
            <span>版本</span>
            <select
              aria-label="选择视频版本"
              className="h-11 max-w-40 rounded-md border border-input bg-background px-2 text-sm text-foreground outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
              disabled={!onSelectVersion || versions.length < 2}
              onChange={(event) => onSelectVersion?.(event.currentTarget.value)}
              value={selectedId}
            >
              {versions.map((version, index) => (
                <option key={version.id} value={version.id}>
                  {version.label ?? `版本 ${index + 1}`}
                </option>
              ))}
            </select>
          </label>
        )}
        <button
          aria-label={`下载视频：${activeFilename}`}
          className="inline-flex h-11 items-center gap-2 rounded-md border border-border px-3 text-sm font-medium text-foreground outline-none hover:bg-muted focus-visible:ring-2 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
          disabled={!canDownload}
          onClick={onDownload}
          type="button"
        >
          <Download aria-hidden="true" className="size-4" />
          下载
        </button>
      </div>
    </article>
  );
}

function VideoStatusPanel({
  progress,
  status,
}: {
  progress?: number;
  status: Exclude<VideoArtifactStatus, "ready">;
}) {
  const failed = status === "failed";
  const normalizedProgress =
    progress === undefined || !Number.isFinite(progress)
      ? undefined
      : Math.round(Math.min(100, Math.max(0, progress)));
  const StatusIcon =
    failed ? CircleAlert : status === "queued" ? Clock3 : LoaderCircle;
  return (
    <section
      aria-live={failed ? "assertive" : "polite"}
      className={`flex aspect-video min-h-48 flex-col items-center justify-center gap-4 px-6 text-center ${
        failed ? "bg-destructive/5 text-destructive" : "bg-muted/30"
      }`}
      role={failed ? "alert" : "status"}
    >
      <StatusIcon
        aria-hidden="true"
        className={`size-6 ${failed ? "" : status === "queued" ? "" : "animate-spin motion-reduce:animate-none"}`}
      />
      <div className="text-sm font-medium">{STATUS_TEXT[status]}</div>
      {!failed && (
        <div className="w-full max-w-72">
          <div
            aria-label="视频生成进度"
            aria-valuemax={100}
            aria-valuemin={0}
            aria-valuenow={normalizedProgress}
            aria-valuetext={
              normalizedProgress === undefined
                ? "进度未知"
                : `${normalizedProgress}%`
            }
            className="h-2 overflow-hidden rounded-full bg-muted"
            role="progressbar"
          >
            <div
              className={`h-full rounded-full bg-primary ${
                normalizedProgress === undefined
                  ? "w-1/3 animate-pulse motion-reduce:animate-none"
                  : ""
              }`}
              style={
                normalizedProgress === undefined
                  ? undefined
                  : { width: `${normalizedProgress}%` }
              }
            />
          </div>
          <div className="mt-2 text-xs text-muted-foreground">
            {normalizedProgress === undefined
              ? "正在等待进度更新"
              : `${normalizedProgress}%`}
          </div>
        </div>
      )}
    </section>
  );
}

function formatDuration(durationSeconds: number) {
  const totalSeconds = Math.max(0, Math.floor(durationSeconds));
  const hours = Math.floor(totalSeconds / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);
  const seconds = totalSeconds % 60;
  return hours > 0
    ? `${hours}:${String(minutes).padStart(2, "0")}:${String(seconds).padStart(2, "0")}`
    : `${minutes}:${String(seconds).padStart(2, "0")}`;
}
