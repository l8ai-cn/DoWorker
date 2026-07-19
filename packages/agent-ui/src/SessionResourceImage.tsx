import { ImageIcon, LoaderCircle } from "lucide-react";
import { useEffect, useState } from "react";

import { ZoomableImage } from "./AgentImageLightbox";

export interface SessionResourceImageProps {
  alt: string;
  className?: string;
  loadBlob: (path: string) => Promise<Blob>;
  path?: string;
}

export function SessionResourceImage({
  alt,
  className,
  loadBlob,
  path,
}: SessionResourceImageProps) {
  const [blobUrl, setBlobUrl] = useState<string | null>(null);
  const [state, setState] = useState<"loading" | "loaded" | "error">(
    "loading",
  );

  useEffect(() => {
    if (!path) {
      setState("error");
      return;
    }
    setBlobUrl(null);
    setState("loading");
    let active = true;
    let objectUrl: string | null = null;
    void loadBlob(path)
      .then((blob) => {
        if (!active) return;
        objectUrl = URL.createObjectURL(blob);
        setBlobUrl(objectUrl);
        setState("loaded");
      })
      .catch(() => {
        if (active) setState("error");
      });
    return () => {
      active = false;
      if (objectUrl) URL.revokeObjectURL(objectUrl);
    };
  }, [loadBlob, path]);

  if (state === "error") {
    return (
      <div
        aria-label={alt}
        className={joinClassNames(
          "flex items-center gap-1.5 rounded-md border border-border bg-muted px-2 py-1.5 text-xs text-muted-foreground",
          className,
        )}
        role="img"
      >
        <ImageIcon className="size-3.5 shrink-0" />
        <span className="truncate">{alt}</span>
      </div>
    );
  }

  if (state === "loading" || !blobUrl) {
    return (
      <div
        aria-label="Loading image"
        className={joinClassNames(
          "flex size-24 items-center justify-center rounded-md border border-border bg-muted text-muted-foreground",
          className,
        )}
        role="status"
      >
        <LoaderCircle className="size-4 animate-spin" />
      </div>
    );
  }

  return <ZoomableImage alt={alt} className={className} src={blobUrl} />;
}

function joinClassNames(...values: Array<string | undefined>) {
  return values.filter(Boolean).join(" ");
}
