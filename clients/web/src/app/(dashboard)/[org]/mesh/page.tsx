"use client";

import { useEffect, useMemo } from "react";
import { CenteredSpinner } from "@/components/ui/spinner";
import { MeshTopology } from "@/components/mesh";
import { useMeshStore, useTopology, type MeshNode } from "@/stores/mesh";
import { useCurrentOrg } from "@/stores/auth";
import { useTranslations } from "next-intl";
import { cn } from "@/lib/utils";
import { RefreshCw } from "lucide-react";

export default function MeshPage() {
  const t = useTranslations();
  const currentOrg = useCurrentOrg();
  const topology = useTopology();
  const loading = useMeshStore((s) => s.loading);
  const error = useMeshStore((s) => s.error);
  const fetchTopology = useMeshStore((s) => s.fetchTopology);
  const clearError = useMeshStore((s) => s.clearError);

  useEffect(() => {
    if (currentOrg) fetchTopology();
  }, [currentOrg, fetchTopology]);

  const activePodCount = useMemo(
    () =>
      topology?.nodes.filter((n: MeshNode) => n.status === "running" || n.status === "initializing").length || 0,
    [topology?.nodes],
  );

  const runnerCounts = useMemo(() => {
    const runners = topology?.runners || [];
    const online = runners.filter((r) => r.status === "online").length;
    return { online, total: runners.length };
  }, [topology?.runners]);

  return (
    <div className="flex h-full w-full min-w-0 flex-col overflow-hidden">
      <header className="flex items-center justify-between panel-lift bg-surface-muted/40 px-6 py-3.5">
        <h1 className="text-[18px] font-semibold text-foreground">{t("mesh.page.title")}</h1>

        <div className="flex items-center gap-2">
          <span className="inline-flex items-center gap-1.5 rounded-full bg-surface-muted px-3 py-1 text-xs font-medium text-foreground shadow-[var(--shadow-soft)]">
            <span className="h-2 w-2 rounded-full bg-success" />
            {t("mesh.page.activePods", { count: activePodCount })}
          </span>

          <span className="inline-flex items-center gap-1.5 rounded-full bg-surface-muted px-3 py-1 text-xs font-medium text-foreground shadow-[var(--shadow-soft)]">
            <span aria-hidden="true">🖥</span>
            {t("mesh.page.runnersOnline", {
              online: runnerCounts.online,
              total: runnerCounts.total,
            })}
          </span>

          <button
            type="button"
            onClick={() => fetchTopology()}
            className="motion-interactive pressable inline-flex h-7 items-center gap-1.5 rounded-md bg-surface-raised px-2.5 text-xs font-medium text-foreground shadow-[var(--shadow-soft)] hover:bg-surface-muted"
          >
            <RefreshCw className={cn("h-3.5 w-3.5", loading && "animate-spin")} />
            {t("mesh.page.refresh")}
          </button>
        </div>
      </header>

      {error && (
        <div className="mx-6 mt-4 flex items-center justify-between rounded-md border border-destructive/30 bg-destructive/10 px-4 py-2 text-sm text-destructive">
          <span>{error}</span>
          <button type="button" onClick={clearError} className="text-xs font-medium hover:underline">
            {t("mesh.page.dismiss")}
          </button>
        </div>
      )}

      <div className="relative flex-1 bg-subtle">
        {loading && !topology && (
          <div className="absolute inset-0 z-10 flex items-center justify-center bg-background/50">
            <CenteredSpinner />
          </div>
        )}
        <MeshTopology />

        {loading && topology && (
          <div className="absolute right-4 top-4 inline-flex items-center gap-1.5 rounded-full bg-surface-raised/90 px-3 py-1 text-xs text-muted-foreground shadow-[var(--shadow-soft)] backdrop-blur-sm">
            <span className="h-2 w-2 animate-pulse rounded-full bg-primary" />
            {t("mesh.page.updating")}
          </div>
        )}
      </div>
    </div>
  );
}
