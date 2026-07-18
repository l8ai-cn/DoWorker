"use client";

import { useCallback, useState } from "react";
import { useRouter } from "next/navigation";
import { useTranslations } from "next-intl";
import { toast } from "sonner";
import { useWorkflowStore } from "@/stores/workflow";

export function useWorkflowDetailActions(slug: string, orgSlug: string) {
  const t = useTranslations();
  const router = useRouter();

  const triggerWorkflow = useWorkflowStore((s) => s.triggerWorkflow);
  const cancelRun = useWorkflowStore((s) => s.cancelRun);
  const enableWorkflow = useWorkflowStore((s) => s.enableWorkflow);
  const disableWorkflow = useWorkflowStore((s) => s.disableWorkflow);
  const loadMoreRuns = useWorkflowStore((s) => s.loadMoreRuns);

  const [revisionOpen, setRevisionOpen] = useState(false);
  const [triggering, setTriggering] = useState(false);

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

  return {
    revisionOpen,
    setRevisionOpen,
    triggering,
    handleTrigger,
    handleLoadMore,
    handleCancelRun,
    handleOpenRun,
    handleEnable,
    handleDisable,
  };
}
