"use client";

import { useEffect, useState } from "react";
import { useWorkflowStore, useCurrentWorkflow, useWorkflowRuns } from "@/stores/workflow";
import { CenteredSpinner } from "@/components/ui/spinner";
import { WorkflowCreateDialog } from "@/components/workflows/WorkflowCreateDialog";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { useTranslations } from "next-intl";
import { WorkflowHeader, WorkflowConfigSection } from "@/app/(dashboard)/[org]/workflows/[slug]/components";
import { WorkflowDetailTabs } from "@/app/(dashboard)/[org]/workflows/[slug]/components/WorkflowDetailTabs";
import { WorkflowDetailErrorState } from "@/components/workflows/WorkflowDetailErrorState";
import { WorkflowDetailRunsPanel } from "@/components/workflows/WorkflowDetailRunsPanel";
import { useWorkflowDetailActions } from "@/components/workflows/workflow-detail-actions";

interface WorkflowDetailPaneProps {
  slug: string;
  orgSlug: string;
  embedded?: boolean;
}

type TabId = "runs" | "prompt" | "autopilot";

export function WorkflowDetailPane({ slug, orgSlug, embedded }: WorkflowDetailPaneProps) {
  const t = useTranslations();
  const currentWorkflow = useCurrentWorkflow();
  const runs = useWorkflowRuns();
  const runsLoading = useWorkflowStore((s) => s.runsLoading);
  const runsTotalCount = useWorkflowStore((s) => s.runsTotalCount);
  const workflowLoading = useWorkflowStore((s) => s.workflowLoading);
  const error = useWorkflowStore((s) => s.error);
  const fetchWorkflow = useWorkflowStore((s) => s.fetchWorkflow);
  const fetchRuns = useWorkflowStore((s) => s.fetchRuns);
  const clearError = useWorkflowStore((s) => s.clearError);
  const setCurrentWorkflow = useWorkflowStore((s) => s.setCurrentWorkflow);

  const [activeTab, setActiveTab] = useState<TabId>("runs");
  const actions = useWorkflowDetailActions(slug, orgSlug, embedded);

  useEffect(() => {
    fetchWorkflow(slug);
    fetchRuns(slug, { limit: 20, offset: 0 });
    return () => setCurrentWorkflow(null);
  }, [slug, fetchWorkflow, fetchRuns, setCurrentWorkflow]);

  if (workflowLoading && !currentWorkflow) return <CenteredSpinner className="h-full" />;

  if (error && !currentWorkflow) {
    return (
      <WorkflowDetailErrorState
        error={error}
        retryLabel={t("workflows.retry")}
        onRetry={() => {
          clearError();
          fetchWorkflow(slug);
        }}
      />
    );
  }

  if (!currentWorkflow) return null;

  const tabs = [
    { id: "runs", label: t("workflows.tabs.runs") },
    { id: "prompt", label: t("workflows.tabs.prompt") },
    { id: "autopilot", label: t("workflows.tabs.autopilot") },
  ];

  return (
    <div className="flex h-full flex-col overflow-hidden">
      <div className="px-8 pt-6">
        <WorkflowHeader
          workflow={currentWorkflow}
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
        <WorkflowDetailTabs active={activeTab} onChange={(id) => setActiveTab(id as TabId)} tabs={tabs} />
      </div>

      <div className="flex-1 overflow-y-auto px-8 py-6">
        {activeTab === "runs" && (
          <WorkflowDetailRunsPanel
            workflow={currentWorkflow}
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
          <WorkflowConfigSection workflow={currentWorkflow} orgSlug={orgSlug} t={t} />
        )}

        {activeTab === "autopilot" && (
          <div className="surface-card bg-muted/30 p-6 text-sm text-muted-foreground">
            {t("workflows.tabs.autopilotEmpty")}
          </div>
        )}
      </div>

      <WorkflowCreateDialog
        open={actions.editOpen}
        onOpenChange={actions.setEditOpen}
        onCreated={() => {
          actions.setEditOpen(false);
          fetchWorkflow(slug);
        }}
        editWorkflow={currentWorkflow}
      />
      <ConfirmDialog {...actions.deleteDialog.dialogProps} />
    </div>
  );
}
