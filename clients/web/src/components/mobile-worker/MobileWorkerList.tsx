"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import Link from "next/link";
import { useParams } from "next/navigation";
import {
  AlertCircle,
  Bot,
  ChevronRight,
  Loader2,
  RefreshCw,
  SquareTerminal,
  WifiOff,
} from "lucide-react";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import { getPodDisplayName } from "@/lib/pod-display-name";
import { cn } from "@/lib/utils";
import type { Pod } from "@/stores/pod";
import {
  fetchMobileWorkerPods,
  useMobileWorkerPods,
} from "./mobileWorkerPodQuery";

const STATUS_TONE: Record<string, string> = {
  running: "bg-success",
  initializing: "bg-warning",
  paused: "bg-muted-foreground",
  disconnected: "bg-danger",
  orphaned: "bg-warning",
  completed: "bg-success",
};

function isAcpWorker(pod: Pod): boolean {
  return (pod as Pod & { interaction_mode?: string }).interaction_mode === "acp";
}

export function MobileWorkerList() {
  const t = useTranslations("mobile.workers");
  const statusT = useTranslations("pods.status");
  const params = useParams<{ org: string }>();
  const orgSlug = typeof params.org === "string" ? params.org : "";
  const pods = useMobileWorkerPods(orgSlug);
  const loadSequence = useRef(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [online, setOnline] = useState(
    () => typeof navigator === "undefined" || navigator.onLine,
  );

  const load = useCallback(async () => {
    const sequence = ++loadSequence.current;
    setLoading(true);
    setError(null);
    try {
      await fetchMobileWorkerPods(orgSlug);
    } catch (fetchError) {
      if (sequence === loadSequence.current) {
        setError(fetchError instanceof Error ? fetchError.message : "Failed to fetch pods");
      }
    } finally {
      if (sequence === loadSequence.current) setLoading(false);
    }
  }, [orgSlug]);

  useEffect(() => {
    void load();
    return () => {
      loadSequence.current += 1;
    };
  }, [load]);

  useEffect(() => {
    const sync = () => setOnline(navigator.onLine);
    window.addEventListener("online", sync);
    window.addEventListener("offline", sync);
    return () => {
      window.removeEventListener("online", sync);
      window.removeEventListener("offline", sync);
    };
  }, []);

  if (loading) {
    return (
      <div
        data-testid="mobile-workers-loading"
        className="flex h-full min-h-64 items-center justify-center"
      >
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex min-h-64 items-center justify-center p-6 text-center">
        <div className="space-y-3">
          <AlertCircle className="mx-auto h-8 w-8 text-danger" />
          <p className="text-sm font-medium">{t("errorTitle")}</p>
          <p className="text-xs text-muted-foreground">{error}</p>
          <Button className="h-11" variant="outline" onClick={() => void load()}>
            <RefreshCw className="h-4 w-4" />
            {t("retry")}
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="mx-auto w-full max-w-3xl">
      {!online && (
        <div className="flex min-h-11 items-center gap-2 border-b bg-warning-bg px-4 text-xs text-warning">
          <WifiOff className="h-4 w-4" />
          {t("offline")}
        </div>
      )}
      {pods.length === 0 ? (
        <div className="flex min-h-64 items-center justify-center p-6 text-center">
          <div className="space-y-2">
            <SquareTerminal className="mx-auto h-8 w-8 text-muted-foreground" />
            <p className="text-sm font-medium">{t("emptyTitle")}</p>
            <p className="text-xs text-muted-foreground">{t("emptyBody")}</p>
          </div>
        </div>
      ) : (
        <ul className="divide-y divide-border/70">
          {pods.map((pod) => {
            const acp = isAcpWorker(pod);
            const ModeIcon = acp ? Bot : SquareTerminal;
            return (
              <li key={pod.pod_key}>
                <Link
                  href={`/${encodeURIComponent(orgSlug)}/mobile/workers/${encodeURIComponent(pod.pod_key)}`}
                  className="flex min-h-16 items-center gap-3 px-4 py-3 active:bg-surface-muted"
                >
                  <ModeIcon className="h-5 w-5 shrink-0 text-muted-foreground" />
                  <div className="min-w-0 flex-1">
                    <p className="truncate font-mono text-sm font-medium">
                      {getPodDisplayName(pod, 48)}
                    </p>
                    <div className="mt-1 flex items-center gap-2 text-xs text-muted-foreground">
                      <span
                        className={cn(
                          "h-2 w-2 rounded-full",
                          STATUS_TONE[pod.status] ?? "bg-muted-foreground",
                        )}
                      />
                      <span>{statusT(pod.status)}</span>
                      <span aria-hidden="true">·</span>
                      <span>{acp ? t("acp") : t("terminal")}</span>
                    </div>
                  </div>
                  <ChevronRight className="h-5 w-5 shrink-0 text-muted-foreground" />
                </Link>
              </li>
            );
          })}
        </ul>
      )}
    </div>
  );
}
