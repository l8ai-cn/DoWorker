"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import { useTranslations } from "next-intl";
import { Bot, Pencil, Play, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { CenteredSpinner } from "@/components/ui/spinner";
import { useExpertStore, useCurrentExpert } from "@/stores/expert";
import { ExpertEditDrawer } from "./ExpertEditDrawer";
import { ExpertConfigList } from "./ExpertConfigList";
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
      <div className="px-8 pt-6 pb-4 border-b border-border">
        <div className="flex items-start justify-between gap-4">
          <div className="flex items-start gap-3 min-w-0">
            <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-primary/10">
              <Bot className="h-5 w-5 text-primary" />
            </div>
            <div className="min-w-0">
              <h1 className="text-xl font-semibold truncate">{expert.name}</h1>
              <p className="text-sm text-muted-foreground">{expert.slug}</p>
              {expert.description && (
                <p className="mt-1 text-sm text-muted-foreground">{expert.description}</p>
              )}
              <p className="mt-1 text-xs text-muted-foreground">
                {expert.run_count > 0
                  ? t("runCount", { count: expert.run_count })
                  : t("neverRun")}
                {expert.last_run_at && ` · ${t("lastRun", { time: formatTimeAgo(expert.last_run_at, tRoot) })}`}
              </p>
            </div>
          </div>
          <div className="flex items-center gap-2 shrink-0">
            <Button size="sm" onClick={handleRun} disabled={running} className="gap-1.5">
              <Play className="h-3.5 w-3.5" />
              {running ? t("running") : t("runExpert")}
            </Button>
            <Button size="sm" variant="outline" onClick={() => setEditOpen(true)} className="gap-1.5">
              <Pencil className="h-3.5 w-3.5" />
              {t("edit.editExpert")}
            </Button>
            <Button size="sm" variant="outline" onClick={() => setDeleteOpen(true)} className="gap-1.5">
              <Trash2 className="h-3.5 w-3.5" />
              {t("deleteExpert")}
            </Button>
          </div>
        </div>
      </div>

      <ExpertConfigList expert={expert} />

      <ExpertEditDrawer
        open={editOpen}
        onOpenChange={setEditOpen}
        expert={expert}
        onSaved={() => fetchExpert(slug)}
      />

      <ConfirmDialog
        open={deleteOpen}
        onOpenChange={setDeleteOpen}
        title={t("deleteConfirmTitle")}
        description={t("deleteConfirmDescription")}
        confirmLabel={t("deleteExpert")}
        variant="destructive"
        onConfirm={handleDelete}
      />
    </div>
  );
}
