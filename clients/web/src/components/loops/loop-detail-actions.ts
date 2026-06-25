"use client";

import { useCallback, useState } from "react";
import { useRouter } from "next/navigation";
import { useTranslations } from "next-intl";
import { toast } from "sonner";
import { useLoopStore } from "@/stores/loop";
import { useConfirmDialog } from "@/components/ui/confirm-dialog";

export function useLoopDetailActions(slug: string, orgSlug: string, embedded?: boolean) {
  const t = useTranslations();
  const router = useRouter();

  const triggerLoop = useLoopStore((s) => s.triggerLoop);
  const cancelRun = useLoopStore((s) => s.cancelRun);
  const enableLoop = useLoopStore((s) => s.enableLoop);
  const disableLoop = useLoopStore((s) => s.disableLoop);
  const deleteLoop = useLoopStore((s) => s.deleteLoop);
  const loadMoreRuns = useLoopStore((s) => s.loadMoreRuns);

  const [editOpen, setEditOpen] = useState(false);
  const [triggering, setTriggering] = useState(false);

  const deleteDialog = useConfirmDialog({
    title: t("loops.deleteConfirm"),
    confirmText: t("common.delete"),
    variant: "destructive",
  });

  const handleTrigger = useCallback(async () => {
    setTriggering(true);
    try {
      const result = await triggerLoop(slug);
      if (result.skipped) toast.info(t("loops.triggerSkipped"), { description: result.reason });
      else if (result.run) toast.success(t("loops.triggered"), { description: `Run #${result.run.run_number}` });
    } catch {
      toast.error(t("loops.triggerFailed"));
    } finally {
      setTriggering(false);
    }
  }, [slug, triggerLoop, t]);

  const handleLoadMore = useCallback(() => loadMoreRuns(slug), [slug, loadMoreRuns]);

  const handleCancelRun = useCallback(
    async (runId: number) => {
      try {
        await cancelRun(slug, runId);
        toast.success(t("loops.runCancelled"));
      } catch {
        toast.error(t("loops.cancelFailed"));
      }
    },
    [slug, cancelRun, t],
  );

  const handleOpenRun = useCallback(
    (run: { pod_key?: string }) => {
      if (run.pod_key) router.push(`/${orgSlug}/workspace?pod=${run.pod_key}`);
    },
    [router, orgSlug],
  );

  const handleEnable = useCallback(async () => {
    try {
      await enableLoop(slug);
      toast.success(t("loops.enabled"));
    } catch {
      toast.error(t("loops.enableFailed"));
    }
  }, [slug, enableLoop, t]);

  const handleDisable = useCallback(async () => {
    try {
      await disableLoop(slug);
      toast.success(t("loops.disabled"));
    } catch {
      toast.error(t("loops.disableFailed"));
    }
  }, [slug, disableLoop, t]);

  const handleDelete = useCallback(async () => {
    const confirmed = await deleteDialog.confirm();
    if (!confirmed) return;
    try {
      await deleteLoop(slug);
      toast.success(t("loops.deleted"));
      if (!embedded) router.push(`/${orgSlug}/loops`);
    } catch (err) {
      const message = (err as Error).message;
      const isActiveRunsError = message.includes("active runs");
      toast.error(t("loops.deleteFailed"), {
        description: isActiveRunsError ? t("loops.deleteHasActiveRuns") : message,
      });
    }
  }, [slug, deleteLoop, deleteDialog, router, orgSlug, t, embedded]);

  return {
    editOpen,
    setEditOpen,
    triggering,
    deleteDialog,
    handleTrigger,
    handleLoadMore,
    handleCancelRun,
    handleOpenRun,
    handleEnable,
    handleDisable,
    handleDelete,
  };
}
