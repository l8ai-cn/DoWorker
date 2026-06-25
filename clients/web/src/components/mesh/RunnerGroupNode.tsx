"use client";

import { memo } from "react";
import { type NodeProps } from "@xyflow/react";
import { useTranslations } from "next-intl";

interface RunnerGroupData {
  runnerNodeId: string;
  runnerStatus: string;
  podCount: number;
}

function RunnerGroupNode({ data }: NodeProps) {
  const t = useTranslations("mesh");
  const { runnerNodeId, runnerStatus, podCount } = data as unknown as RunnerGroupData;
  const isOnline = runnerStatus === "online";

  return (
    <div className="surface-card bg-muted/30 shadow-sm w-full h-full">
      <div className="flex items-center gap-2 px-4 py-2 bg-surface-muted/30">
        <div
          className={`w-2 h-2 rounded-full ${
            isOnline ? "bg-success" : "bg-muted-foreground"
          }`}
        />
        <span className="text-sm font-medium text-foreground truncate">
          {runnerNodeId}
        </span>
        <span className={`text-xs px-1.5 py-0.5 rounded ${
          isOnline
            ? "bg-success-bg text-success"
            : "bg-muted text-muted-foreground"
        }`}>
          {isOnline ? t("runnerGroup.online") : t("runnerGroup.offline")}
        </span>
        <span className="ml-auto text-xs text-muted-foreground bg-muted px-2 py-0.5 rounded-full">
          {t("runnerGroup.podCount", { count: podCount })}
        </span>
      </div>
      <div className="p-4" />
    </div>
  );
}

export default memo(RunnerGroupNode);
