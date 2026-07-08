"use client";

import { useState } from "react";
import { File, Download, Globe, ChevronDown, ChevronRight } from "lucide-react";
import { useTranslations } from "next-intl";
import { cn } from "@/lib/utils";
import { classifyMediaUrl } from "@/lib/media/url";
import { LightboxImage } from "@/components/media/MediaLightbox";
import { VideoEmbed } from "@/components/media/VideoEmbed";
import { HtmlPreviewCard } from "@/components/media/HtmlPreviewCard";

interface AttachmentCardProps {
  url: string;
  className?: string;
}

function fileNameOf(url: string): string {
  try {
    const parsed = new URL(url, "http://x");
    const path = parsed.pathname;
    const segments = path.split("/").filter(Boolean);
    return segments[segments.length - 1] || url;
  } catch {
    return url;
  }
}

export function AttachmentCard({ url, className }: AttachmentCardProps) {
  const t = useTranslations("channels.attachment");
  const kind = classifyMediaUrl(url);
  const [errored, setErrored] = useState(false);
  const [htmlExpanded, setHtmlExpanded] = useState(false);

  if (kind === "image" && !errored) {
    return (
      <div data-testid="message-attachment" className={cn("mt-1.5", className)}>
        <LightboxImage
          src={url}
          className="max-w-[320px]"
          imgClassName="max-h-[240px] w-full object-cover"
          onError={() => setErrored(true)}
        />
      </div>
    );
  }

  if (kind === "video" || kind === "audio") {
    return (
      <div data-testid="message-attachment" className={cn("mt-1.5", className)}>
        <VideoEmbed url={url} kind={kind} className="max-w-[400px]" />
      </div>
    );
  }

  if (kind === "html") {
    const name = fileNameOf(url);
    return (
      <div data-testid="message-attachment" className={cn("mt-1.5", className)}>
        <button
          type="button"
          onClick={() => setHtmlExpanded((v) => !v)}
          aria-expanded={htmlExpanded}
          className="inline-flex items-center gap-1.5 rounded-md border border-border bg-muted/40 px-2 py-1 text-xs text-foreground hover:bg-muted"
        >
          {htmlExpanded ? (
            <ChevronDown className="h-3 w-3 text-muted-foreground" />
          ) : (
            <ChevronRight className="h-3 w-3 text-muted-foreground" />
          )}
          <Globe className="h-3.5 w-3.5 text-muted-foreground" />
          <span className="max-w-[220px] truncate">{name}</span>
        </button>
        {htmlExpanded && <HtmlPreviewCard src={url} className="mt-1.5 max-w-xl" />}
      </div>
    );
  }

  const name = fileNameOf(url);
  return (
    <a
      href={url}
      target="_blank"
      rel="noreferrer"
      download
      data-testid="message-attachment"
      className={cn(
        "mt-1.5 inline-flex items-center gap-1.5 rounded-md border border-border bg-muted/40 px-2 py-1 text-xs text-foreground hover:bg-muted",
        className,
      )}
      aria-label={t("download")}
    >
      <File className="h-3.5 w-3.5 text-muted-foreground" />
      <span className="max-w-[220px] truncate">{name}</span>
      <Download className="h-3 w-3 text-muted-foreground" />
    </a>
  );
}

export default AttachmentCard;
