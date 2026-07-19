"use client";

import { useEffect, useState } from "react";
import { WorkerModelResourceSelect } from "@/components/pod/CreatePodForm/WorkerModelResourceSelect";
import { useWorkerModelResources } from "@/components/pod/hooks/useWorkerModelResources";
import { getSessionImportWorkerRequirement } from "@/lib/api/sessionImportWorkerPlan";

interface Props {
  open: boolean;
  orgSlug: string;
  workerTypeSlug: string | null;
  selectedResourceId: number | null;
  onSelect: (resourceId: number | null) => void;
  t: (key: string) => string;
}

export function ImportCodexModelResourceSelect({
  open,
  orgSlug,
  workerTypeSlug,
  selectedResourceId,
  onSelect,
  t,
}: Props) {
  const [requirement, setRequirement] = useState<{
    required: boolean;
    protocolAdapters: string[];
  } | null>(null);
  const [requirementError, setRequirementError] = useState<string | null>(null);
  const [checking, setChecking] = useState(false);
  const resources = useWorkerModelResources(
    requirement ? workerTypeSlug : null,
    selectedResourceId,
    false,
    requirement ?? { required: false, protocolAdapters: [] },
  );

  useEffect(() => {
    if (!open || !workerTypeSlug) {
      setRequirement(null);
      setRequirementError(null);
      return;
    }
    let cancelled = false;
    setChecking(true);
    setRequirementError(null);
    onSelect(null);
    void getSessionImportWorkerRequirement(orgSlug, workerTypeSlug)
      .then((value) => {
        if (!cancelled) setRequirement({
          required: value.requiresModelResource,
          protocolAdapters: value.modelProtocolAdapters,
        });
      })
      .catch((error: unknown) => {
        if (!cancelled) {
          setRequirement(null);
          setRequirementError(error instanceof Error ? error.message : "无法加载 Worker 创建选项");
        }
      })
      .finally(() => {
        if (!cancelled) setChecking(false);
      });
    return () => {
      cancelled = true;
    };
  }, [onSelect, open, orgSlug, workerTypeSlug]);

  if (!workerTypeSlug) return null;
  if (checking) return <p className="text-xs text-muted-foreground">{t("common.loading")}</p>;
  if (requirementError) return <p className="text-xs text-destructive">{requirementError}</p>;
  if (!requirement?.required) return null;
  return (
    <WorkerModelResourceSelect
      resources={resources.modelResources}
      selectedResourceId={resources.selectedModelResourceId}
      onSelect={onSelect}
      loading={resources.loadingModelResources}
      error={resources.modelResourceError}
      validationError={selectedResourceId === null ? "请选择模型资源" : undefined}
      t={t}
    />
  );
}
