"use client";

import { RefreshCw, FolderOpen } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Select, SelectContent, SelectItem, SelectTrigger } from "@/components/ui/select";
import type { RunnerData } from "@/lib/api";

interface RunnerPodsToolbarProps {
  runner: RunnerData;
  podFilter: string;
  loadingPods: boolean;
  loadingSandbox: boolean;
  t: (key: string) => string;
  onFilterChange: (filter: string) => void;
  onRefresh: () => void;
  onRefreshSandbox: () => void;
}

export function RunnerPodsToolbar({
  runner,
  podFilter,
  loadingPods,
  loadingSandbox,
  t,
  onFilterChange,
  onRefresh,
  onRefreshSandbox,
}: RunnerPodsToolbarProps) {
  return (
    <div className="flex items-center justify-between">
      <Select value={podFilter || "all"} onValueChange={onFilterChange}>
        <SelectTrigger className="w-[180px]">
          <span>
            {podFilter === "all" || !podFilter
              ? t("runners.detail.allStatus")
              : t(`pods.status.${podFilter}`)}
          </span>
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">{t("runners.detail.allStatus")}</SelectItem>
          <SelectItem value="running">{t("pods.status.running")}</SelectItem>
          <SelectItem value="terminated">{t("pods.status.terminated")}</SelectItem>
          <SelectItem value="error">{t("pods.status.error")}</SelectItem>
        </SelectContent>
      </Select>

      <div className="flex items-center gap-2">
        <Button
          variant="outline"
          onClick={onRefreshSandbox}
          disabled={loadingSandbox || runner.status !== "online"}
        >
          {loadingSandbox ? (
            <RefreshCw className="w-4 h-4 mr-2 animate-spin" />
          ) : (
            <FolderOpen className="w-4 h-4 mr-2" />
          )}
          {t("runners.detail.refreshSandbox")}
        </Button>
        <Button variant="outline" onClick={onRefresh} disabled={loadingPods}>
          {loadingPods ? (
            <RefreshCw className="w-4 h-4 mr-2 animate-spin" />
          ) : (
            <RefreshCw className="w-4 h-4 mr-2" />
          )}
          {t("common.refresh")}
        </Button>
      </div>
    </div>
  );
}
