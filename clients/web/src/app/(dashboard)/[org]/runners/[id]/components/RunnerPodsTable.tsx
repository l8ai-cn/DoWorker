"use client";

import { formatDistanceToNow } from "date-fns";
import {
  CheckCircle,
  XCircle,
  GitBranch,
  RotateCcw,
  Ticket,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import type { RunnerPodData, SandboxStatus } from "@/lib/api";
import { cn } from "@/lib/utils";
import { AgentStatusBadge } from "@/components/shared/AgentStatusBadge";
import { getShortPodKey } from "@/lib/pod-display-name";

function podStatusClass(status: string) {
  const colors: Record<string, string> = {
    running: "bg-success-bg text-success",
    initializing: "bg-info-bg text-info",
    terminated: "bg-muted text-muted-foreground",
    error: "bg-danger-bg text-danger",
    paused: "bg-warning-bg text-warning",
  };
  return colors[status] || "bg-muted text-muted-foreground";
}

interface RunnerPodsTableProps {
  pods: RunnerPodData[];
  sandboxStatuses: Map<string, SandboxStatus>;
  t: (key: string, params?: Record<string, string | number>) => string;
  onResume: (pod: RunnerPodData) => void;
}

export function RunnerPodsTable({ pods, sandboxStatuses, t, onResume }: RunnerPodsTableProps) {
  return (
    <div className="surface-card overflow-hidden">
      <table className="w-full">
        <thead className="bg-muted">
          <tr>
            <th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
              {t("runners.detail.podKey")}
            </th>
            <th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
              {t("runners.detail.status")}
            </th>
            <th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
              {t("runners.detail.agentStatus")}
            </th>
            <th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
              {t("runners.detail.sandbox")}
            </th>
            <th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
              {t("runners.detail.branch")}
            </th>
            <th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
              {t("runners.detail.ticket")}
            </th>
            <th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
              {t("runners.detail.createdAt")}
            </th>
            <th className="px-4 py-3 text-right text-xs font-medium text-muted-foreground uppercase tracking-wider">
              {t("runners.detail.actions")}
            </th>
          </tr>
        </thead>
        <tbody className="divide-y divide-border/25">
          {pods.map((pod) => {
            const sandboxStatus = sandboxStatuses.get(pod.pod_key);
            const isInactive = pod.status !== "running" && pod.status !== "initializing";
            const alreadyResumed = Boolean(pod.resumed_by_pod_key);
            const canResume = isInactive && sandboxStatus?.can_resume && !alreadyResumed;

            return (
              <tr
                key={pod.pod_key}
                data-testid="runner-pod-row"
                data-pod-key={pod.pod_key}
                className="motion-interactive hover:bg-surface-muted"
              >
                <td className="px-4 py-3">
                  <span className="text-sm font-medium text-foreground">{pod.pod_key}</span>
                  {pod.source_pod_key && (
                    <span className="ml-2 text-xs text-muted-foreground">
                      (resumed from {getShortPodKey(pod.source_pod_key)}...)
                    </span>
                  )}
                </td>
                <td className="px-4 py-3">
                  <span
                    className={cn(
                      "inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium",
                      podStatusClass(pod.status),
                    )}
                  >
                    {pod.status}
                  </span>
                </td>
                <td className="px-4 py-3">
                  <AgentStatusBadge agentStatus={pod.agent_status} podStatus={pod.status} variant="badge" />
                </td>
                <td className="px-4 py-3">
                  {pod.status === "running" ? (
                    <span className="flex items-center text-success text-sm">
                      <CheckCircle className="w-4 h-4 mr-1" />
                      {t("runners.detail.active")}
                    </span>
                  ) : isInactive ? (
                    sandboxStatus === undefined ? (
                      <span className="text-muted-foreground text-sm">-</span>
                    ) : sandboxStatus.exists ? (
                      <span className="flex items-center text-success text-sm">
                        <CheckCircle className="w-4 h-4 mr-1" />
                        {sandboxStatus.can_resume ? t("runners.detail.canResume") : t("runners.detail.exists")}
                      </span>
                    ) : (
                      <span className="flex items-center text-muted-foreground text-sm">
                        <XCircle className="w-4 h-4 mr-1" />
                        {t("runners.detail.notExists")}
                      </span>
                    )
                  ) : (
                    <span className="text-muted-foreground text-sm">-</span>
                  )}
                </td>
                <td className="px-4 py-3 text-sm text-muted-foreground">
                  {pod.branch_name ? (
                    <span className="flex items-center">
                      <GitBranch className="w-4 h-4 mr-1" />
                      {pod.branch_name}
                    </span>
                  ) : (
                    "-"
                  )}
                </td>
                <td className="px-4 py-3 text-sm text-muted-foreground">
                  {pod.ticket ? (
                    <div className="flex items-start gap-1.5 max-w-[200px]">
                      <Ticket className="w-4 h-4 mt-0.5 shrink-0 text-muted-foreground" />
                      <div className="min-w-0">
                        <span className="text-xs font-medium text-foreground">{pod.ticket.slug}</span>
                        <p className="text-xs text-muted-foreground truncate" title={pod.ticket.title}>
                          {pod.ticket.title}
                        </p>
                      </div>
                    </div>
                  ) : (
                    "-"
                  )}
                </td>
                <td className="px-4 py-3 text-sm text-muted-foreground">
                  {formatDistanceToNow(new Date(pod.created_at ?? ""), { addSuffix: true })}
                </td>
                <td className="px-4 py-3 text-right">
                  <div className="flex items-center justify-end gap-2">
                    {alreadyResumed && (
                      <span className="text-xs text-muted-foreground" title={pod.resumed_by_pod_key}>
                        {t("runners.detail.alreadyResumedBadge", {
                          podKey: getShortPodKey(pod.resumed_by_pod_key ?? ""),
                        })}
                      </span>
                    )}
                    {canResume && (
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => onResume(pod)}
                        title={t("runners.detail.resumeTooltip")}
                      >
                        <RotateCcw className="w-4 h-4 mr-1" />
                        {t("runners.detail.resume")}
                      </Button>
                    )}
                  </div>
                </td>
              </tr>
            );
          })}
          {pods.length === 0 && (
            <tr>
              <td colSpan={8} className="px-4 py-8 text-center text-muted-foreground">
                {t("runners.detail.noPods")}
              </td>
            </tr>
          )}
        </tbody>
      </table>
    </div>
  );
}
