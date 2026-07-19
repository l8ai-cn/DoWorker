import { useEffect, useState } from "react";
import { ImageIcon } from "lucide-react";
import { Spinner } from "@/components/ui/spinner";
import { authenticatedFetch } from "@/lib/identity";
import { cn } from "@/lib/utils";
import { ZoomableImage } from "@/components/ImageLightbox";

export interface SessionImageProps {
  path?: string;
  alt: string;
  className?: string;
}

export function SessionImage({ path, alt, className }: SessionImageProps) {
  const [blobUrl, setBlobUrl] = useState<string | null>(null);
  const [state, setState] = useState<"loading" | "loaded" | "error">("loading");

  useEffect(() => {
    if (!path) {
      setState("error");
      return;
    }
    setState("loading");
    setBlobUrl(null);
    let cancelled = false;
    let objectUrl: string | null = null;
    authenticatedFetch(path)
      .then((res) => (res.ok ? res.blob() : Promise.reject(new Error(`HTTP ${res.status}`))))
      .then((blob) => {
        if (cancelled) return;
        objectUrl = URL.createObjectURL(blob);
        setBlobUrl(objectUrl);
        setState("loaded");
      })
      .catch(() => {
        if (!cancelled) setState("error");
      });
    return () => {
      cancelled = true;
      if (objectUrl) URL.revokeObjectURL(objectUrl);
    };
  }, [path]);

  if (state === "error") {
    return (
      <div
        role="img"
        aria-label={alt}
        className={cn(
          "flex items-center gap-1.5 rounded-md border border-border bg-muted px-2 py-1.5 text-xs text-muted-foreground",
          className,
        )}
      >
        <ImageIcon className="size-3.5 shrink-0" />
        <span className="truncate">{alt}</span>
      </div>
    );
  }

  if (state === "loading" || !blobUrl) {
    return (
      <div
        role="status"
        aria-label="Loading image"
        className={cn(
          "flex size-24 items-center justify-center rounded-md border border-border bg-muted text-muted-foreground",
          className,
        )}
      >
        <Spinner />
      </div>
    );
  }

  return <ZoomableImage src={blobUrl} alt={alt} className={className} />;
}
