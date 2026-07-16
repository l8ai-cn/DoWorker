import { useEffect, useState } from "react";
import type { FileContentResponse } from "@/hooks/useFileContent";
import { fileContentToBlob } from "@/hooks/useFileContent";

const VIDEO_EXTENSIONS = new Set(["mp4", "m4v", "webm", "ogg", "ogv"]);

export function isWorkspaceVideoFile(path: string, contentType?: string | null): boolean {
  if (contentType) return contentType.startsWith("video/");
  const extension = path.split(".").pop()?.toLowerCase() ?? "";
  return VIDEO_EXTENSIONS.has(extension);
}

export function WorkspaceVideoViewer({ data, path }: { data: FileContentResponse; path: string }) {
  const [url, setUrl] = useState<string | null>(null);
  const [playbackError, setPlaybackError] = useState(false);
  const filename = path.split("/").pop() ?? path;

  useEffect(() => {
    if (data.truncated) {
      setUrl(null);
      return;
    }
    setPlaybackError(false);
    const objectUrl = URL.createObjectURL(fileContentToBlob(data));
    setUrl(objectUrl);
    return () => URL.revokeObjectURL(objectUrl);
  }, [data]);

  if (data.truncated) {
    return (
      <div
        className="flex h-full items-center justify-center p-8 text-sm text-muted-foreground"
        role="alert"
      >
        Video is too large to preview because the workspace response was truncated.
      </div>
    );
  }
  if (playbackError) {
    return (
      <div
        className="flex h-full items-center justify-center p-8 text-sm text-destructive"
        role="alert"
      >
        Video playback failed.
      </div>
    );
  }

  return (
    <div className="flex h-full min-h-0 items-center justify-center overflow-auto bg-black p-4">
      {url && (
        <video
          aria-label={filename}
          className="max-h-full max-w-full"
          controls
          onError={() => setPlaybackError(true)}
          preload="metadata"
          src={url}
        />
      )}
    </div>
  );
}
