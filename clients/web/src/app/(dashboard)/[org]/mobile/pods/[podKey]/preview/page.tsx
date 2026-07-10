"use client";

import { useEffect, useRef, useState } from "react";
import Link from "next/link";
import { useParams } from "next/navigation";
import { AlertCircle } from "lucide-react";
import { useTranslations } from "next-intl";
import { CenteredSpinner } from "@/components/ui/spinner";
import { getPodPreviewSession } from "@/lib/api/podPreview";
import { buildPodMobileConsoleUrl } from "@/lib/pod-mobile-access";

export default function MobilePodPreviewPage() {
  const params = useParams<{ org: string; podKey: string }>();
  const t = useTranslations();
  const started = useRef(false);
  const [error, setError] = useState<string | null>(null);
  const orgSlug = typeof params.org === "string" ? params.org : "";
  const podKey = typeof params.podKey === "string" ? params.podKey : "";

  useEffect(() => {
    if (!orgSlug || !podKey || started.current) return;
    started.current = true;
    getPodPreviewSession(orgSlug, podKey)
      .then((session) => window.location.replace(session.session_url))
      .catch((err) => {
        setError(err instanceof Error ? err.message : t("mobile.preview.error"));
      });
  }, [orgSlug, podKey, t]);

  if (!error) {
    return (
      <div data-testid="mobile-preview-loading" className="h-full bg-background">
        <CenteredSpinner className="h-full" />
      </div>
    );
  }

  return (
    <div className="flex h-full items-center justify-center bg-background p-6 text-center">
      <div className="max-w-sm space-y-4">
        <AlertCircle className="mx-auto h-10 w-10 text-destructive" />
        <div className="space-y-1">
          <p className="text-sm font-medium text-foreground">{t("mobile.preview.unavailable")}</p>
          <p className="text-xs text-muted-foreground">{error}</p>
        </div>
        <Link
          href={buildPodMobileConsoleUrl(orgSlug, podKey)}
          className="motion-interactive inline-flex h-9 items-center justify-center rounded-md bg-surface-raised px-4 py-2 text-sm font-medium text-foreground shadow-xs ring-1 ring-border/30 hover:bg-accent"
        >
          {t("mobile.preview.backToConsole")}
        </Link>
      </div>
    </div>
  );
}
