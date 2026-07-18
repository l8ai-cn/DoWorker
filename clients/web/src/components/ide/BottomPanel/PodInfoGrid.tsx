"use client";

import type { PodData } from "@/lib/api/facade/pod";
import { AgentStatusBadge } from "@/components/shared/AgentStatusBadge";
import {
  AlertCircle,
  Bot,
  Clock,
  FolderGit2,
  GitBranch,
  Server,
  Terminal,
  Ticket,
  User,
  Wrench,
} from "lucide-react";
import { InfoRow } from "./InfoRow";

interface PodInfoGridProps {
  pod: PodData;
  orgSlug: string;
  t: (key: string, params?: Record<string, string | number>) => string;
}

export function PodInfoGrid({ pod, orgSlug, t }: PodInfoGridProps) {
  return (
    <div className="grid grid-cols-2 gap-x-6 gap-y-1.5">
      <InfoRow
        icon={<Terminal className="w-3 h-3" />}
        label={t("ide.bottomPanel.infoTab.podKey")}
        value={pod.pod_key}
        mono
      />

      {pod.agent && (
        <InfoRow
          icon={<Bot className="w-3 h-3" />}
          label={t("ide.bottomPanel.infoTab.agent")}
          value={pod.agent.name}
        />
      )}

      {pod.worker_skill_slugs && pod.worker_skill_slugs.length > 0 && (
        <InfoRow
          icon={<Wrench className="w-3 h-3" />}
          label={t("ide.bottomPanel.infoTab.skills")}
          value={pod.worker_skill_slugs.join(", ")}
          mono
          className="col-span-2"
        />
      )}

      {pod.agent_status && pod.status === "running" && (
        <InfoRow
          icon={<Bot className="w-3 h-3" />}
          label={t("ide.bottomPanel.infoTab.agentStatus")}
          value={
            <AgentStatusBadge
              agentStatus={pod.agent_status}
              podStatus={pod.status}
              variant="inline"
            />
          }
        />
      )}

      {pod.runner && (
        <InfoRow
          icon={<Server className="w-3 h-3" />}
          label={t("ide.bottomPanel.infoTab.runner")}
          value={pod.runner.node_id}
          mono
        />
      )}

      {pod.repository && (
        <InfoRow
          icon={<FolderGit2 className="w-3 h-3" />}
          label={t("ide.bottomPanel.infoTab.repository")}
          value={pod.repository.slug}
          href={orgSlug ? `/${orgSlug}/infra?tab=repositories&id=${pod.repository.id}` : undefined}
        />
      )}

      {pod.branch_name && (
        <InfoRow
          icon={<GitBranch className="w-3 h-3" />}
          label={t("ide.bottomPanel.infoTab.branch")}
          value={pod.branch_name}
          mono
        />
      )}

      {pod.sandbox_path && (
        <InfoRow
          icon={<FolderGit2 className="w-3 h-3" />}
          label={t("ide.bottomPanel.infoTab.worktree")}
          value={pod.sandbox_path}
          mono
        />
      )}

      {pod.ticket && (
        <InfoRow
          icon={<Ticket className="w-3 h-3" />}
          label={t("ide.bottomPanel.infoTab.ticket")}
          value={`${pod.ticket.slug} - ${pod.ticket.title}`}
          href={orgSlug ? `/${orgSlug}/tickets/${pod.ticket.slug}` : undefined}
        />
      )}

      {pod.created_by && (
        <InfoRow
          icon={<User className="w-3 h-3" />}
          label={t("ide.bottomPanel.infoTab.createdBy")}
          value={pod.created_by.name || pod.created_by.username}
        />
      )}

      {pod.started_at && (
        <InfoRow
          icon={<Clock className="w-3 h-3" />}
          label={t("ide.bottomPanel.infoTab.startedAt")}
          value={new Date(pod.started_at).toLocaleString()}
        />
      )}

      <InfoRow
        icon={<Clock className="w-3 h-3" />}
        label={t("ide.bottomPanel.infoTab.createdAt")}
        value={pod.created_at ? new Date(pod.created_at).toLocaleString() : "-"}
      />

      {pod.error_message && (
        <InfoRow
          icon={<AlertCircle className="w-3 h-3 text-danger" />}
          label={t("ide.bottomPanel.infoTab.error")}
          value={`${pod.error_code ? `[${pod.error_code}] ` : ""}${pod.error_message}`}
          className="col-span-2"
          valueClassName="text-danger"
        />
      )}
    </div>
  );
}
