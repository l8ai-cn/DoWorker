"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { AlertCircle, ExternalLink, Loader2, RefreshCw, Video } from "lucide-react";
import { Button } from "@/components/ui/button";
import { getPodPreviewSession } from "@/lib/api/podPreview";

interface VideoPreviewDeliveryProps {
  orgSlug: string;
  podKey: string;
  t: (key: string, params?: Record<string, string | number>) => string;
}

export function VideoPreviewDelivery({
  orgSlug,
  podKey,
  t,
}: VideoPreviewDeliveryProps) {
  const [sessionURL, setSessionURL] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const requestRef = useRef(0);

  const loadPreview = useCallback(async () => {
    const request = ++requestRef.current;
    setLoading(true);
    setError("");
    try {
      const session = await getPodPreviewSession(orgSlug, podKey);
      if (request !== requestRef.current) return;
      setSessionURL(session.session_url);
    } catch {
      if (request !== requestRef.current) return;
      setSessionURL("");
      setError(t("videoWorker.previewUnavailable"));
    } finally {
      if (request === requestRef.current) setLoading(false);
    }
  }, [orgSlug, podKey, t]);

  useEffect(() => {
    void loadPreview();
    return () => {
      requestRef.current += 1;
    };
  }, [loadPreview]);

  if (loading) {
    return (
      <div className="flex h-full items-center justify-center gap-2 text-xs text-muted-foreground">
        <Loader2 className="h-4 w-4 animate-spin" />
        {t("videoWorker.previewLoading")}
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex h-full flex-col items-center justify-center gap-2 text-center">
        <AlertCircle className="h-6 w-6 text-warning" />
        <p className="text-xs font-medium">
          {t("videoWorker.previewUnavailable")}
        </p>
        <p className="max-w-md text-[11px] text-muted-foreground">{error}</p>
        <Button size="sm" variant="outline" onClick={loadPreview}>
          <RefreshCw className="h-3.5 w-3.5" />
          {t("common.refresh")}
        </Button>
      </div>
    );
  }

  return (
    <div className="flex h-full min-h-48 flex-col overflow-hidden rounded-md border border-border/60 bg-black">
      <div className="flex h-8 shrink-0 items-center gap-2 border-b border-white/10 bg-black px-2 text-white">
        <Video className="h-3.5 w-3.5" />
        <span className="text-xs font-medium">
          {t("videoWorker.videoPreview")}
        </span>
        <div className="flex-1" />
        <Button
          size="sm"
          variant="ghost"
          className="h-6 w-6 p-0 text-white hover:bg-white/10 hover:text-white"
          onClick={loadPreview}
          title={t("common.refresh")}
        >
          <RefreshCw className="h-3.5 w-3.5" />
        </Button>
        <a
          href={sessionURL}
          target="_blank"
          rel="noreferrer"
          className="inline-flex h-6 w-6 items-center justify-center rounded text-white hover:bg-white/10"
          title={t("videoWorker.openPreview")}
        >
          <ExternalLink className="h-3.5 w-3.5" />
        </a>
      </div>
      <iframe
        key={sessionURL}
        src={sessionURL}
        title={t("videoWorker.videoPreview")}
        className="min-h-0 flex-1 border-0 bg-black"
        allow="autoplay; fullscreen"
      />
    </div>
  );
}
