"use client";

import { useCallback, useState } from "react";
import { useRouter } from "next/navigation";
import { useTranslations } from "next-intl";
import { toast } from "sonner";
import { useWorkflowStore } from "@/stores/workflow";
import { useConfirmDialog } from "@/components/ui/confirm-dialog";

export function useWorkflowDetailActions(slug: string, orgSlug: string, embedded?: boolean) {
  const t = useTranslations();
  const router = useRouter();

  const triggerWorkflow = useWorkflowStore((s) => s.triggerWorkflow);
  const cancelRun = useWorkflowStore((s) => s.cancelRun);
  const enableWorkflow = useWorkflowStore((s) => s.enableWorkflow);
  const disableWorkflow = useWorkflowStore((s) => s.disableWorkflow);
  const deleteWorkflow = useWorkflowStore((s) => s.deleteWorkflow);
  const loadMoreRuns = useWorkflowStore((s) => s.loadMoreRuns);

  const [editOpen, setEditOpen] = useState(false);
  const [triggering, setTriggering] = useState(false);

  const deleteDialog = useConfirmDialog({
    title: t("workflows.deleteConfirm"),
    confirmText: t("common.delete"),
    variant: "destructive",
  });

  const handleTrigger = useCallback(async () => {
    setTriggering(true);
    try {
      const result = await triggerWorkflow(slug);
      if (result.skipped) toast.info(t("workflows.triggerSkipped"), { description: result.reason });
      else if (result.run) toast.success(t("workflows.triggered"), { description: `Run #${result.run.run_number}` });
    } catch {
      toast.error(t("workflows.triggerFailed"));
    } finally {
      setTriggering(false);
    }
  }, [slug, triggerWorkflow, t]);

  const handleLoadMore = useCallback(() => loadMoreRuns(slug), [slug, loadMoreRuns]);

  const handleCancelRun = useCallback(
    async (runId: number) => {
      try {
        await cancelRun(slug, runId);
        toast.success(t("workflows.runCancelled"));
      } catch {
        toast.error(t("workflows.cancelFailed"));
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
      await enableWorkflow(slug);
      toast.success(t("workflows.enabled"));
    } catch {
      toast.error(t("workflows.enableFailed"));
    }
  }, [slug, enableWorkflow, t]);

  const handleDisable = useCallback(async () => {
    try {
      await disableWorkflow(slug);
      toast.success(t("workflows.disabled"));
    } catch {
      toast.error(t("workflows.disableFailed"));
    }
  }, [slug, disableWorkflow, t]);

  const handleDelete = useCallback(async () => {
    const confirmed = await deleteDialog.confirm();
    if (!confirmed) return;
    try {
      await deleteWorkflow(slug);
      toast.success(t("workflows.deleted"));
      if (!embedded) router.push(`/${orgSlug}/workflows`);
    } catch (err) {
      const message = (err as Error).message;
      const isActiveRunsError = message.includes("active runs");
      toast.error(t("workflows.deleteFailed"), {
        description: isActiveRunsError ? t("workflows.deleteHasActiveRuns") : message,
      });
    }
  }, [slug, deleteWorkflow, deleteDialog, router, orgSlug, t, embedded]);

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
