"use client";

import { useState } from "react";
import { Code2, ExternalLink, Eye, Maximize2, Minimize2 } from "lucide-react";
import { useTranslations } from "next-intl";
import { cn } from "@/lib/utils";
import { isSafeRenderableSrc } from "@/lib/media/url";

type Tab = "preview" | "code";

interface HtmlPreviewCardProps {
  /** Inline HTML document rendered via iframe srcDoc. */
  html?: string;
  /** Remote HTML document URL (must be http/https). Mutually exclusive with `html`. */
  src?: string;
  /**
   * While true (e.g. the assistant is still streaming the code block), the
   * card stays on the code tab so a half-written document doesn't reload the
   * iframe on every chunk. Once streaming ends the preview tab activates
   * automatically unless the user already picked a tab.
   */
  streaming?: boolean;
  className?: string;
}

// Sandboxed preview for AI-generated web pages. The iframe deliberately has
// `allow-scripts` but NOT `allow-same-origin`, so embedded scripts run in an
// opaque origin and cannot touch the host app's storage or cookies.
export function HtmlPreviewCard({ html, src, streaming = false, className }: HtmlPreviewCardProps) {
  const t = useTranslations("media");
  const hasCode = typeof html === "string";
  const [tab, setTab] = useState<Tab>(hasCode && streaming ? "code" : "preview");
  const [tall, setTall] = useState(false);
  const [prevStreaming, setPrevStreaming] = useState(streaming);
  const [userChose, setUserChose] = useState(false);

  // Adjust-state-during-render: when streaming finishes, flip to the preview
  // tab unless the user already picked a tab manually.
  if (prevStreaming !== streaming) {
    setPrevStreaming(streaming);
    if (!streaming && !userChose && hasCode) setTab("preview");
  }

  const safeSrc = src && isSafeRenderableSrc(src) ? src : undefined;
  if (!hasCode && !safeSrc) return null;

  const selectTab = (next: Tab) => {
    setUserChose(true);
    setTab(next);
  };

  const openInNewTab = () => {
    if (safeSrc) {
      window.open(safeSrc, "_blank", "noopener,noreferrer");
      return;
    }
    if (hasCode) {
      const blob = new Blob([html ?? ""], { type: "text/html" });
      const url = URL.createObjectURL(blob);
      window.open(url, "_blank", "noopener,noreferrer");
      // Give the new tab time to load the document before revoking.
      setTimeout(() => URL.revokeObjectURL(url), 30_000);
    }
  };

  const showPreview = tab === "preview";

  return (
    <div
      data-testid="html-preview-card"
      className={cn("my-2 overflow-hidden rounded-lg border border-border bg-card not-prose", className)}
    >
      <div className="flex items-center justify-between gap-2 border-b border-border/60 bg-muted/30 px-2 py-1.5">
        <div className="flex items-center gap-1">
          {hasCode && (
            <TabButton
              active={tab === "code"}
              icon={Code2}
              label={t("code")}
              onClick={() => selectTab("code")}
            />
          )}
          <TabButton
            active={tab === "preview"}
            icon={Eye}
            label={t("preview")}
            onClick={() => selectTab("preview")}
          />
        </div>
        <div className="flex items-center gap-1">
          <button
            type="button"
            onClick={() => setTall((v) => !v)}
            aria-label={tall ? t("collapse") : t("expand")}
            className="flex h-6 w-6 items-center justify-center rounded text-muted-foreground hover:bg-muted hover:text-foreground"
          >
            {tall ? <Minimize2 className="h-3.5 w-3.5" /> : <Maximize2 className="h-3.5 w-3.5" />}
          </button>
          <button
            type="button"
            onClick={openInNewTab}
            aria-label={t("openInNewTab")}
            className="flex h-6 w-6 items-center justify-center rounded text-muted-foreground hover:bg-muted hover:text-foreground"
          >
            <ExternalLink className="h-3.5 w-3.5" />
          </button>
        </div>
      </div>

      {showPreview ? (
        <iframe
          {...(safeSrc ? { src: safeSrc } : { srcDoc: html })}
          title={t("htmlDocument")}
          sandbox="allow-scripts"
          referrerPolicy="no-referrer"
          className={cn("w-full border-0 bg-white", tall ? "h-[32rem]" : "h-80")}
        />
      ) : (
        <pre
          className={cn(
            "m-0 overflow-auto bg-muted p-3 text-xs leading-relaxed",
            tall ? "max-h-[32rem]" : "max-h-80",
          )}
        >
          <code>{html}</code>
        </pre>
      )}
    </div>
  );
}

function TabButton({
  active,
  icon: Icon,
  label,
  onClick,
}: {
  active: boolean;
  icon: React.ComponentType<{ className?: string }>;
  label: string;
  onClick: () => void;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      aria-pressed={active}
      className={cn(
        "flex items-center gap-1 rounded px-2 py-0.5 text-xs font-medium transition-colors",
        active
          ? "bg-background text-foreground shadow-xs"
          : "text-muted-foreground hover:text-foreground",
      )}
    >
      <Icon className="h-3 w-3" />
      {label}
    </button>
  );
}
