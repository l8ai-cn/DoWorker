"use client";

import { useEffect, useState } from "react";
import { useLoopStore, useCurrentLoop, useLoopRuns } from "@/stores/loop";
import { CenteredSpinner } from "@/components/ui/spinner";
import { LoopCreateDialog } from "@/components/loops/LoopCreateDialog";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { useTranslations } from "next-intl";
import { LoopHeader, LoopConfigSection } from "@/app/(dashboard)/[org]/loops/[slug]/components";
import { LoopDetailTabs } from "@/app/(dashboard)/[org]/loops/[slug]/components/LoopDetailTabs";
import { LoopDetailErrorState } from "@/components/loops/LoopDetailErrorState";
import { LoopDetailRunsPanel } from "@/components/loops/LoopDetailRunsPanel";
import { useLoopDetailActions } from "@/components/loops/loop-detail-actions";

interface LoopDetailPaneProps {
  slug: string;
  orgSlug: string;
  embedded?: boolean;
}

type TabId = "runs" | "prompt" | "autopilot";

export function LoopDetailPane({ slug, orgSlug, embedded }: LoopDetailPaneProps) {
  const t = useTranslations();
  const currentLoop = useCurrentLoop();
  const runs = useLoopRuns();
  const runsLoading = useLoopStore((s) => s.runsLoading);
  const runsTotalCount = useLoopStore((s) => s.runsTotalCount);
  const loopLoading = useLoopStore((s) => s.loopLoading);
  const error = useLoopStore((s) => s.error);
  const fetchLoop = useLoopStore((s) => s.fetchLoop);
  const fetchRuns = useLoopStore((s) => s.fetchRuns);
  const clearError = useLoopStore((s) => s.clearError);
  const setCurrentLoop = useLoopStore((s) => s.setCurrentLoop);

  const [activeTab, setActiveTab] = useState<TabId>("runs");
  const actions = useLoopDetailActions(slug, orgSlug, embedded);

  useEffect(() => {
    fetchLoop(slug);
    fetchRuns(slug, { limit: 20, offset: 0 });
    return () => setCurrentLoop(null);
  }, [slug, fetchLoop, fetchRuns, setCurrentLoop]);

  if (loopLoading && !currentLoop) return <CenteredSpinner className="h-full" />;

  if (error && !currentLoop) {
    return (
      <LoopDetailErrorState
        error={error}
        retryLabel={t("loops.retry")}
        onRetry={() => {
          clearError();
          fetchLoop(slug);
        }}
      />
    );
  }

  if (!currentLoop) return null;

  const tabs = [
    { id: "runs", label: t("loops.tabs.runs") },
    { id: "prompt", label: t("loops.tabs.prompt") },
    { id: "autopilot", label: t("loops.tabs.autopilot") },
  ];

  return (
    <div className="flex h-full flex-col overflow-hidden">
      <div className="px-8 pt-6">
        <LoopHeader
          loop={currentLoop}
          triggering={actions.triggering}
          t={t}
          onTrigger={actions.handleTrigger}
          onEdit={() => actions.setEditOpen(true)}
          onEnable={actions.handleEnable}
          onDisable={actions.handleDisable}
          onDelete={actions.handleDelete}
        />
      </div>

      <div className="px-8">
        <LoopDetailTabs active={activeTab} onChange={(id) => setActiveTab(id as TabId)} tabs={tabs} />
      </div>

      <div className="flex-1 overflow-y-auto px-8 py-6">
        {activeTab === "runs" && (
          <LoopDetailRunsPanel
            loop={currentLoop}
            runs={runs}
            runsLoading={runsLoading}
            runsTotalCount={runsTotalCount}
            t={t}
            onOpenRun={actions.handleOpenRun}
            onCancelRun={actions.handleCancelRun}
            onLoadMore={actions.handleLoadMore}
            onEdit={() => actions.setEditOpen(true)}
          />
        )}

        {activeTab === "prompt" && (
          <LoopConfigSection loop={currentLoop} orgSlug={orgSlug} t={t} />
        )}

        {activeTab === "autopilot" && (
          <div className="surface-card bg-muted/30 p-6 text-sm text-muted-foreground">
            {t("loops.tabs.autopilotEmpty")}
          </div>
        )}
      </div>

      <LoopCreateDialog
        open={actions.editOpen}
        onOpenChange={actions.setEditOpen}
        onCreated={() => {
          actions.setEditOpen(false);
          fetchLoop(slug);
        }}
        editLoop={currentLoop}
      />
      <ConfirmDialog {...actions.deleteDialog.dialogProps} />
    </div>
  );
}
