"use client";

import { useEffect, useState } from "react";
import { Code2, ExternalLink, Eye, Maximize2, Minimize2 } from "lucide-react";
import { useTranslations } from "next-intl";
import {
  STATIC_HTML_REFERRER_POLICY,
  STATIC_HTML_SANDBOX,
  openStaticHtmlInNewWindow,
  staticHtmlDocument,
} from "@do-worker/agent-ui";
import { cn } from "@/lib/utils";

type Tab = "preview" | "code";

interface HtmlPreviewCardProps {
  html: string;
  /**
   * While true (e.g. the assistant is still streaming the code block), the
   * card stays on the code tab so a half-written document doesn't reload the
   * iframe on every chunk. Once streaming ends the preview tab activates
   * automatically unless the user already picked a tab.
   */
  streaming?: boolean;
  className?: string;
}

export function HtmlPreviewCard({ html, streaming = false, className }: HtmlPreviewCardProps) {
  const t = useTranslations("media");
  const [tab, setTab] = useState<Tab>(streaming ? "code" : "preview");
  const [tall, setTall] = useState(false);
  const [prevStreaming, setPrevStreaming] = useState(streaming);
  const [userChose, setUserChose] = useState(false);
  const [openError, setOpenError] = useState(false);
  const [staticDocument, setStaticDocument] = useState<{ html: string; srcDoc: string }>();

  if (prevStreaming !== streaming) {
    setPrevStreaming(streaming);
    if (!streaming && !userChose) setTab("preview");
  }

  useEffect(() => {
    let active = true;
    queueMicrotask(() => {
      if (active) {
        setStaticDocument({ html, srcDoc: staticHtmlDocument(html) });
      }
    });
    return () => {
      active = false;
    };
  }, [html]);

  const selectTab = (next: Tab) => {
    setUserChose(true);
    setTab(next);
  };

  const openInNewTab = () => {
    const result = openStaticHtmlInNewWindow(html, t("htmlDocument"));
    setOpenError(!result.opened);
  };

  const showPreview = tab === "preview";
  const srcDoc = staticDocument?.html === html ? staticDocument.srcDoc : undefined;

  return (
    <div
      data-testid="html-preview-card"
      className={cn("my-2 overflow-hidden rounded-lg border border-border bg-card not-prose", className)}
    >
      <div className="flex items-center justify-between gap-2 border-b border-border/60 bg-muted/30 px-2 py-1.5">
        <div className="flex items-center gap-1">
          <TabButton
            active={tab === "code"}
            icon={Code2}
            label={t("code")}
            onClick={() => selectTab("code")}
          />
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

      {openError && (
        <p
          role="alert"
          className="border-b border-destructive/30 bg-destructive/5 px-3 py-2 text-xs text-destructive"
        >
          {t("popupBlocked")}
        </p>
      )}

      {showPreview ? (
        <iframe
          srcDoc={srcDoc}
          title={t("htmlDocument")}
          sandbox={STATIC_HTML_SANDBOX}
          referrerPolicy={STATIC_HTML_REFERRER_POLICY}
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
