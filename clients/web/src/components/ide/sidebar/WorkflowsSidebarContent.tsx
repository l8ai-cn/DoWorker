"use client";

import React, { useEffect, useState, useCallback } from "react";
import { useRouter, usePathname } from "next/navigation";
import { cn } from "@/lib/utils";
import { useCurrentOrg } from "@/stores/auth";
import { useWorkflowStore, useWorkflows, WorkflowData } from "@/stores/workflow";
import { WorkflowCreateDialog } from "@/components/workflows/WorkflowCreateDialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Loader2,
  Plus,
  Search,
  RefreshCw,
  Clock,
  Bot,
  Zap,
  Repeat,
} from "lucide-react";
import { useTranslations } from "next-intl";
import { formatTimeAgo } from "@/lib/utils/time";

export function WorkflowsSidebarContent({ className }: { className?: string }) {
  const t = useTranslations();
  const router = useRouter();
  const pathname = usePathname();
  const currentOrg = useCurrentOrg();
  const workflows = useWorkflows();
  const loading = useWorkflowStore((s) => s.loading);
  const fetchWorkflows = useWorkflowStore((s) => s.fetchWorkflows);

  const [searchQuery, setSearchQuery] = useState("");
  const [refreshing, setRefreshing] = useState(false);
  const [createDialogOpen, setCreateDialogOpen] = useState(false);

  useEffect(() => {
    if (currentOrg) {
      fetchWorkflows();
    }
  }, [currentOrg, fetchWorkflows]);

  const handleRefresh = useCallback(async () => {
    setRefreshing(true);
    try {
      await fetchWorkflows();
    } finally {
      setRefreshing(false);
    }
  }, [fetchWorkflows]);

  const handleWorkflowClick = useCallback(
    (slug: string) => {
      router.push(`/${currentOrg?.slug}/workflows/${slug}`);
    },
    [router, currentOrg]
  );

  const handleCreated = useCallback(() => {
    setCreateDialogOpen(false);
    fetchWorkflows();
  }, [fetchWorkflows]);

  const filteredWorkflows = searchQuery
    ? workflows.filter(
        (l) =>
          l.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
          l.slug.toLowerCase().includes(searchQuery.toLowerCase())
      )
    : workflows;

  const activeSlug = pathname?.match(/\/workflows\/([^/]+)/)?.[1] || null;

  return (
    <div className={cn("flex flex-col h-full", className)}>
      <WorkflowCreateDialog
        open={createDialogOpen}
        onOpenChange={setCreateDialogOpen}
        onCreated={handleCreated}
      />

      {/* Search */}
      <div className="px-2 py-2">
        <div className="relative">
          <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-muted-foreground" />
          <Input
            placeholder={t("workflows.searchPlaceholder")}
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-8 h-8 text-sm bg-muted/50 border-transparent focus:border-border focus:bg-background transition-colors"
          />
        </div>
      </div>

      {/* Action buttons */}
      <div className="flex items-center gap-1 px-2 pb-2">
        <Button
          size="sm"
          variant="outline"
          className="flex-1 h-7 text-xs gap-1"
          onClick={() => setCreateDialogOpen(true)}
        >
          <Plus className="w-3 h-3" />
          {t("workflows.createWorkflow")}
        </Button>
        <Button
          size="sm"
          variant="ghost"
          className="h-7 w-7 p-0"
          onClick={handleRefresh}
          disabled={refreshing}
        >
          <RefreshCw className={cn("w-3.5 h-3.5", refreshing && "animate-spin")} />
        </Button>
      </div>

      {/* Workflow list */}
      <div className="flex-1 overflow-y-auto">
        {/* Count header */}
        <div className="px-3 py-1.5 text-[10px] uppercase tracking-wider text-muted-foreground font-medium">
          {t("workflows.workflowCount", { count: filteredWorkflows.length })}
        </div>

        {loading ? (
          <div className="flex items-center justify-center py-8">
            <Loader2 className="w-4 h-4 animate-spin text-muted-foreground" />
          </div>
        ) : filteredWorkflows.length === 0 ? (
          <div className="flex flex-col items-center px-3 py-8 text-center">
            <Repeat className="w-8 h-8 text-muted-foreground/40 mb-2" />
            <p className="text-xs text-muted-foreground">
              {searchQuery ? t("workflows.noMatch") : t("workflows.emptyState")}
            </p>
          </div>
        ) : (
          <div className="px-1 pb-1">
            {filteredWorkflows.map((workflow) => (
              <WorkflowListItem
                key={workflow.id}
                workflow={workflow}
                onClick={handleWorkflowClick}
                isActive={activeSlug === workflow.slug}
              />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

function WorkflowListItem({
  workflow,
  onClick,
  isActive,
}: {
  workflow: WorkflowData;
  onClick: (slug: string) => void;
  isActive: boolean;
}) {
  const t = useTranslations();
  const isEnabled = workflow.status === "enabled";
  const isRunning = workflow.active_run_count > 0;
  const successRate =
    workflow.total_runs > 0
      ? Math.round((workflow.successful_runs / workflow.total_runs) * 100)
      : null;

  return (
    <button
      data-testid="workflow-row"
      data-workflow-slug={workflow.slug}
      className={cn(
        "w-full text-left px-2.5 py-2 rounded-md",
        "transition-colors duration-150 cursor-pointer",
        isActive
          ? "bg-accent text-accent-foreground"
          : "motion-interactive hover:bg-surface-muted text-foreground"
      )}
      onClick={() => onClick(workflow.slug)}
    >
      {/* Name row */}
      <div className="flex items-center gap-2 mb-1">
        <span className="relative flex-shrink-0">
          <span
            className={cn(
              "block w-2 h-2 rounded-full",
              isRunning ? "bg-info" : isEnabled ? "bg-success" : "bg-muted-foreground"
            )}
          />
          {isRunning && (
            <span className="absolute inset-0 w-2 h-2 rounded-full animate-ping opacity-30 bg-info" />
          )}
        </span>
        <span className="text-sm font-medium truncate">{workflow.name}</span>
      </div>

      {/* Trigger + Mode row */}
      <div className="flex items-center gap-1.5 ml-4 mb-1">
        {workflow.cron_expression ? (
          <span className="inline-flex items-center gap-0.5 text-[10px] text-muted-foreground font-mono">
            <Clock className="w-2.5 h-2.5" />
            {workflow.cron_expression}
          </span>
        ) : (
          <span className="inline-flex items-center gap-0.5 text-[10px] text-muted-foreground">
            <Repeat className="w-2.5 h-2.5" />
            {t("workflows.onDemand")}
          </span>
        )}
        <span className="text-[10px] text-muted-foreground/60 mx-0.5">|</span>
        <span className="inline-flex items-center gap-0.5 text-[10px] text-muted-foreground">
          {workflow.execution_mode === "autopilot" ? (
            <Bot className="w-2.5 h-2.5" />
          ) : (
            <Zap className="w-2.5 h-2.5" />
          )}
          {workflow.execution_mode === "autopilot" ? t("workflows.modeAutoShort") : t("workflows.modeDirect")}
        </span>
      </div>

      {/* Last run + stats */}
      {workflow.last_run_at && (
        <div className="flex items-center gap-1.5 ml-4 text-[10px] text-muted-foreground">
          <Clock className="w-2.5 h-2.5" />
          <span>{formatTimeAgo(workflow.last_run_at, t)}</span>
          {successRate !== null && (
            <>
              <span className="text-muted-foreground/40">|</span>
              <span>{successRate}%</span>
            </>
          )}
        </div>
      )}
    </button>
  );
}

export default WorkflowsSidebarContent;
