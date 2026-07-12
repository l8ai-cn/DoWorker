import { useCallback } from "react";
import { Button } from "@/components/ui/button";
import {
  ConfirmDialog,
  useConfirmDialog,
} from "@/components/ui/confirm-dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
} from "@/components/ui/select";
import type { ExecutionCluster } from "@/lib/api/facade/executionCluster";
import type { Runner } from "@/stores/runner";
import type { TranslationFn } from "../GeneralSettings";
import { RunnerCard } from "./RunnerCard";

interface RunnersPanelProps {
  runners: Runner[];
  loading: boolean;
  onEdit: (runner: Runner) => void;
  onDelete: (id: number) => Promise<void>;
  onGenerateToken: () => void;
  clusters: ExecutionCluster[];
  selectedClusterId: string;
  onSelectCluster: (clusterId: string) => void;
  clustersLoading: boolean;
  t: TranslationFn;
}

export function RunnersPanel({
  runners,
  loading,
  onEdit,
  onDelete,
  onGenerateToken,
  clusters,
  selectedClusterId,
  onSelectCluster,
  clustersLoading,
  t,
}: RunnersPanelProps) {
  const { dialogProps, confirm } = useConfirmDialog();
  const handleDeleteWithConfirm = useCallback(
    async (id: number) => {
      const confirmed = await confirm({
        title: t("settings.runnersSection.deleteDialog.title"),
        description: t("settings.runnersSection.deleteDialog.description"),
        variant: "destructive",
        confirmText: t("settings.runnersSection.deleteDialog.delete"),
        cancelText: t("settings.runnersSection.deleteDialog.cancel"),
      });
      if (confirmed) {
        try {
          await onDelete(id);
        } catch (error) {
          console.error("Failed to delete runner:", error);
        }
      }
    },
    [confirm, onDelete, t],
  );

  return (
    <div className="surface-card p-6">
      <div className="mb-4 flex items-center justify-between">
        <div>
          <h2 className="text-lg font-semibold">
            {t("settings.runnersSection.title")}
          </h2>
          <p className="text-sm text-muted-foreground">
            {t("settings.runnersSection.description")}
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Select
            value={selectedClusterId}
            onValueChange={onSelectCluster}
            disabled={clustersLoading}
          >
            <SelectTrigger className="w-44">
              <span
                className={selectedClusterId ? "" : "text-muted-foreground"}
              >
                {selectedClusterId
                  ? clusters.find(
                      (cluster) => String(cluster.id) === selectedClusterId,
                    )?.name
                  : t("settings.runnersSection.selectCluster")}
              </span>
            </SelectTrigger>
            <SelectContent>
              {clusters.map((cluster) => (
                <SelectItem key={cluster.id} value={String(cluster.id)}>
                  {cluster.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <Button
            variant="outline"
            onClick={onGenerateToken}
            disabled={clustersLoading || !selectedClusterId}
          >
            {t("settings.runnersSection.generateToken")}
          </Button>
        </div>
      </div>

      <RunnerList
        runners={runners}
        loading={loading}
        onEdit={onEdit}
        onDelete={handleDeleteWithConfirm}
        t={t}
      />
      <ConfirmDialog {...dialogProps} />
    </div>
  );
}

function RunnerList({
  runners,
  loading,
  onEdit,
  onDelete,
  t,
}: Pick<RunnersPanelProps, "runners" | "loading" | "onEdit" | "t"> & {
  onDelete: (id: number) => Promise<void>;
}) {
  if (loading) {
    return (
      <div className="text-center py-4 text-muted-foreground">
        {t("settings.runnersSection.loading")}
      </div>
    );
  }
  if (runners.length === 0) {
    return (
      <div className="text-center py-8 text-muted-foreground">
        {t("settings.runnersSection.noRunners")}
      </div>
    );
  }
  return (
    <div className="space-y-3">
      {runners.map((runner) => (
        <RunnerCard
          key={runner.id}
          runner={runner}
          onEdit={onEdit}
          onDelete={() => onDelete(runner.id)}
          formatLastSeen={(date) => formatLastSeen(date, t)}
          t={t}
        />
      ))}
    </div>
  );
}

function formatLastSeen(dateString: string | undefined, t: TranslationFn) {
  if (!dateString) return "Never";
  const date = new Date(dateString);
  const diffSec = Math.floor((Date.now() - date.getTime()) / 1000);
  if (diffSec < 60) return t("settings.runnersSection.justNow");
  if (diffSec < 3600) return `${Math.floor(diffSec / 60)}m ago`;
  if (diffSec < 86400) return `${Math.floor(diffSec / 3600)}h ago`;
  return date.toLocaleDateString();
}
