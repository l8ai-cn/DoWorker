"use client";

import { ExternalLink } from "lucide-react";
import { cn } from "@/lib/utils";
import { buildEmbedURL, isSafeRenderableSrc, isSafeURL, type MediaKind } from "@/lib/media/url";

interface VideoEmbedProps {
  url: string;
  kind: MediaKind;
  className?: string;
}

// Inline playback for chat messages: direct file URLs get a native player,
// whitelisted providers (YouTube / Vimeo / Loom / Figma / CodeSandbox) get a
// sandboxed iframe. Anything else renders a plain external link.
export function VideoEmbed({ url, kind, className }: VideoEmbedProps) {
  if (!isSafeRenderableSrc(url)) return null;

  if (kind === "video") {
    return (
      <div className={cn("max-w-xl overflow-hidden rounded-md border border-border bg-black", className)}>
        <video controls src={url} className="w-full" preload="metadata" />
      </div>
    );
  }

  if (kind === "audio") {
    return (
      <div className={cn("max-w-xl", className)}>
        <audio controls src={url} className="w-full" preload="metadata" />
      </div>
    );
  }

  const embed = buildEmbedURL(url, kind);
  if (embed && isSafeURL(embed)) {
    return (
      <div className={cn("max-w-xl overflow-hidden rounded-md border border-border bg-black", className)}>
        <iframe
          src={embed}
          title={url}
          className="aspect-video w-full"
          allow="accelerometer; autoplay; encrypted-media; gyroscope; picture-in-picture; fullscreen"
          sandbox="allow-scripts allow-same-origin allow-popups allow-popups-to-escape-sandbox allow-presentation"
          referrerPolicy="no-referrer-when-downgrade"
          allowFullScreen
        />
      </div>
    );
  }

  return (
    <a
      href={url}
      target="_blank"
      rel="noopener noreferrer"
      className={cn(
        "inline-flex items-center gap-1.5 rounded-md border border-border bg-muted/40 px-2 py-1 text-sm text-primary hover:bg-muted",
        className,
      )}
    >
      <ExternalLink className="h-3.5 w-3.5" />
      <span className="max-w-[320px] truncate">{url}</span>
    </a>
  );
}
