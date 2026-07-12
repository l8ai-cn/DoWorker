"use client";

import { useState, useEffect } from "react";
import { useCurrentOrg } from "@/stores/auth";
import { isApiErrorCode, getLocalizedErrorMessage } from "@/lib/api/errors";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
} from "@/components/ui/select";
import { ShieldAlert } from "lucide-react";
import { useTranslations } from "next-intl";
import {
  createRegistrationCommand,
  listExecutionClusters,
} from "@/lib/api/facade/executionClusterApi";
import type { ExecutionCluster } from "@/lib/api/facade/executionCluster";
import { RunnerRegistrationInstructions } from "./RunnerRegistrationInstructions";

interface AddRunnerModalProps {
  open: boolean;
  onClose: () => void;
  onCreated?: () => void;
}

export function AddRunnerModal({
  open,
  onClose,
  onCreated,
}: AddRunnerModalProps) {
  const t = useTranslations();
  const currentOrg = useCurrentOrg();
  const [loading, setLoading] = useState(false);
  const [generatedCommand, setGeneratedCommand] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [clusters, setClusters] = useState<ExecutionCluster[]>([]);
  const [selectedClusterId, setSelectedClusterId] = useState("");
  const [clustersLoading, setClustersLoading] = useState(false);
  const orgSlug = currentOrg?.slug ?? "";

  useEffect(() => {
    if (!open) {
      setGeneratedCommand(null);
      setLoading(false);
      setError(null);
      setClusters([]);
      setSelectedClusterId("");
    }
  }, [open]);

  useEffect(() => {
    if (!open || !orgSlug) return;
    let active = true;
    setClustersLoading(true);
    void listExecutionClusters(orgSlug)
      .then((items) => {
        if (active) setClusters(items);
      })
      .catch(() => {
        if (active) setError(t("runners.addRunnerModal.loadClustersFailed"));
      })
      .finally(() => {
        if (active) setClustersLoading(false);
      });
    return () => {
      active = false;
    };
  }, [open, orgSlug, t]);

  if (!open) return null;

  const handleGenerate = async () => {
    if (!selectedClusterId) {
      setError(t("runners.addRunnerModal.clusterRequired"));
      return;
    }
    setLoading(true);
    setError(null);
    try {
      const result = await createRegistrationCommand(
        orgSlug,
        Number(selectedClusterId),
      );
      setGeneratedCommand(result.command);
    } catch (err) {
      if (
        isApiErrorCode(err, "ADMIN_REQUIRED") ||
        isApiErrorCode(err, "INSUFFICIENT_PERMISSIONS")
      ) {
        setError(t("apiErrors.INSUFFICIENT_PERMISSIONS"));
      } else {
        setError(
          getLocalizedErrorMessage(err, t, t("apiErrors.INTERNAL_ERROR")),
        );
      }
    } finally {
      setLoading(false);
    }
  };

  const handleDone = () => {
    setGeneratedCommand(null);
    onCreated?.();
    onClose();
  };

  const handleClose = () => {
    setGeneratedCommand(null);
    onClose();
  };

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <div className="bg-background border border-border rounded-lg w-full max-w-lg p-4 md:p-6 max-h-[90vh] overflow-y-auto">
        <h2 className="text-lg md:text-xl font-semibold mb-2">
          {t("runners.addRunnerModal.title")}
        </h2>
        <p className="text-sm text-muted-foreground mb-4">
          {t("runners.addRunnerModal.subtitle")}
        </p>

        {generatedCommand ? (
          <RunnerRegistrationInstructions
            command={generatedCommand}
            onDone={handleDone}
          />
        ) : (
          <div className="space-y-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">
                {t("runners.addRunnerModal.clusterLabel")}
              </label>
              <Select
                value={selectedClusterId}
                onValueChange={setSelectedClusterId}
                disabled={clustersLoading}
              >
                <SelectTrigger>
                  <span
                    className={selectedClusterId ? "" : "text-muted-foreground"}
                  >
                    {selectedClusterId
                      ? clusters.find(
                          (cluster) => String(cluster.id) === selectedClusterId,
                        )?.name
                      : t("runners.addRunnerModal.clusterPlaceholder")}
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
              {clustersLoading && (
                <p className="text-xs text-muted-foreground">
                  {t("runners.addRunnerModal.clusterLoading")}
                </p>
              )}
            </div>

            {error ? (
              <div className="flex items-start gap-2 p-3 bg-destructive/10 border border-destructive/20 rounded-lg">
                <ShieldAlert className="w-5 h-5 text-destructive flex-shrink-0 mt-0.5" />
                <p className="text-sm text-destructive">{error}</p>
              </div>
            ) : (
              <p className="text-sm text-muted-foreground">
                {t("runners.addRunnerModal.generateHint")}
              </p>
            )}

            <div className="flex flex-col-reverse sm:flex-row justify-end gap-3 mt-6">
              <Button variant="outline" onClick={handleClose}>
                {t("runners.addRunnerModal.cancel")}
              </Button>
              <Button
                onClick={handleGenerate}
                disabled={loading || clustersLoading || !selectedClusterId}
              >
                {loading
                  ? t("runners.addRunnerModal.generating")
                  : t("runners.addRunnerModal.generate")}
              </Button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

export default AddRunnerModal;
