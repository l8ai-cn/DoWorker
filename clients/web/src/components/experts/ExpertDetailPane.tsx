"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import { useTranslations } from "next-intl";
import { Bot, Pencil, Play, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { CenteredSpinner } from "@/components/ui/spinner";
import { useExpertStore, useCurrentExpert } from "@/stores/expert";
import { ExpertEditDrawer } from "./ExpertEditDrawer";
import { ExpertConfigList } from "./ExpertConfigList";
import { ExpertRevisionDialog } from "./ExpertRevisionDialog";
import { usePodStore } from "@/stores/pod";
import { getShortPodKey } from "@/lib/pod-display-name";
import { formatTimeAgo } from "@/lib/utils/time";

interface ExpertDetailPaneProps {
  slug: string;
  orgSlug: string;
}

export function ExpertDetailPane({ slug, orgSlug }: ExpertDetailPaneProps) {
  const t = useTranslations("experts");
  const tRoot = useTranslations();
  const router = useRouter();
  const expert = useCurrentExpert();
  const expertLoading = useExpertStore((s) => s.expertLoading);
  const error = useExpertStore((s) => s.error);
  const runExpert = useExpertStore((s) => s.runExpert);
  const deleteExpert = useExpertStore((s) => s.deleteExpert);
  const clearError = useExpertStore((s) => s.clearError);
  const fetchExpert = useExpertStore((s) => s.fetchExpert);

  const [running, setRunning] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [editOpen, setEditOpen] = useState(false);

  if (expertLoading && !expert) return <CenteredSpinner className="h-full" />;

  if (error && !expert) {
    return (
      <div className="flex h-full flex-col items-center justify-center gap-3 py-20">
        <p className="text-sm text-muted-foreground">{error}</p>
        <Button variant="outline" size="sm" onClick={() => { clearError(); fetchExpert(slug); }}>
          {t("retry")}
        </Button>
      </div>
    );
  }

  if (!expert) return null;

  const resourceManaged =
    expert.orchestration_resource_id != null ||
    expert.orchestration_resource_revision != null ||
    expert.worker_spec_snapshot_id != null;

  const handleRun = async () => {
    setRunning(true);
    try {
      const { pod, warning } = await runExpert(slug);
      usePodStore.getState().upsertPod(pod);
      toast.success(t("runSuccess"), {
        description: t("runSuccessDescription", { podKey: getShortPodKey(pod.pod_key) }),
      });
      if (warning) toast.warning(warning);
      router.push(`/${orgSlug}/workspace?pod=${encodeURIComponent(pod.pod_key)}`);
    } catch (e) {
      toast.error(e instanceof Error ? e.message : String(e));
    } finally {
      setRunning(false);
    }
  };

  const handleDelete = async () => {
    try {
      await deleteExpert(slug);
      router.push(`/${orgSlug}/experts`);
    } catch (e) {
      toast.error(e instanceof Error ? e.message : String(e));
    }
  };

  return (
    <div className="flex h-full flex-col overflow-hidden">
      <div className="border-b border-border bg-gradient-to-b from-muted/30 to-transparent px-4 pt-6 pb-5 sm:px-8">
        <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
          <div className="flex min-w-0 items-start gap-4">
            <div className="flex h-12 w-12 shrink-0 items-center justify-center rounded-xl bg-primary/10 ring-1 ring-primary/15">
              <Bot className="h-6 w-6 text-primary" />
            </div>
            <div className="min-w-0">
              <div className="flex items-center gap-2">
                <h1 className="text-xl font-semibold truncate">{expert.name}</h1>
                {expert.perpetual && (
                  <Badge variant="success" className="font-normal">{t("perpetual")}</Badge>
                )}
              </div>
              <p className="text-sm text-muted-foreground font-mono">{expert.slug}</p>
              {expert.description && (
                <p className="mt-1.5 text-sm text-muted-foreground max-w-2xl">{expert.description}</p>
              )}
              <div className="mt-2.5 flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
                <span className="inline-flex items-center gap-1 rounded-md bg-muted/60 px-2 py-0.5">
                  <Bot className="h-3 w-3" />
                  {expert.agent_slug}
                </span>
                <span className="inline-flex items-center gap-1 rounded-md bg-muted/60 px-2 py-0.5">
                  <Play className="h-3 w-3" />
                  {expert.run_count > 0 ? t("runCount", { count: expert.run_count }) : t("neverRun")}
                </span>
                {expert.last_run_at && (
                  <span className="inline-flex items-center gap-1 rounded-md bg-muted/60 px-2 py-0.5">
                    {t("lastRun", { time: formatTimeAgo(expert.last_run_at, tRoot) })}
                  </span>
                )}
              </div>
            </div>
          </div>
          <div className="flex w-full flex-wrap items-center gap-2 lg:w-auto lg:shrink-0">
            <Button size="sm" onClick={handleRun} disabled={running} className="flex-1 gap-1.5 sm:flex-none">
              <Play className="h-3.5 w-3.5" />
              {running ? t("running") : t("runExpert")}
            </Button>
            <Button size="sm" variant="outline" onClick={() => setEditOpen(true)} className="flex-1 gap-1.5 sm:flex-none">
              <Pencil className="h-3.5 w-3.5" />
              {t("edit.editExpert")}
            </Button>
            {!resourceManaged && (
              <Button size="sm" variant="outline" onClick={() => setDeleteOpen(true)} className="gap-1.5">
                <Trash2 className="h-3.5 w-3.5" />
                {t("deleteExpert")}
              </Button>
            )}
          </div>
        </div>
      </div>

      <ExpertMarketOperations
        key={slug}
        expertID={expert.id}
        expertSlug={slug}
        installedFromMarket={expert.source_market_application_id != null}
        submissionReady={Boolean(expert.worker_spec_snapshot_id)}
        onUpgraded={() => fetchExpert(slug)}
      />

      <ExpertConfigList expert={expert} />

      {resourceManaged ? (
        <ExpertRevisionDialog
          open={editOpen}
          orgSlug={orgSlug}
          expertSlug={slug}
          onOpenChange={setEditOpen}
          onApplied={() => {
            setEditOpen(false);
            void fetchExpert(slug);
          }}
        />
      ) : (
        <ExpertEditDrawer
          open={editOpen}
          onOpenChange={setEditOpen}
          expert={expert}
          onSaved={() => fetchExpert(slug)}
        />
      )}

      {!resourceManaged && (
        <ConfirmDialog
          open={deleteOpen}
          onOpenChange={setDeleteOpen}
          title={t("deleteConfirmTitle")}
          description={t("deleteConfirmDescription")}
          confirmText={t("deleteExpert")}
          variant="destructive"
          onConfirm={handleDelete}
        />
      )}
    </div>
  );
}
