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

interface WorkerRequirement {
  required: boolean;
  protocolAdapters: string[];
}

type RequirementResult =
  | { error: null; key: string; requirement: WorkerRequirement }
  | { error: string; key: string; requirement: null };

export function ImportCodexModelResourceSelect({
  open,
  orgSlug,
  workerTypeSlug,
  selectedResourceId,
  onSelect,
  t,
}: Props) {
  const [result, setResult] = useState<RequirementResult | null>(null);
  const requestKey =
    open && workerTypeSlug ? `${orgSlug}\u0000${workerTypeSlug}` : null;
  const currentResult = result?.key === requestKey ? result : null;
  const requirement = currentResult?.requirement ?? null;
  const requirementError = currentResult?.error ?? null;
  const checking = requestKey !== null && currentResult === null;
  const resources = useWorkerModelResources(
    requirement ? workerTypeSlug : null,
    selectedResourceId,
    false,
    requirement ?? { required: false, protocolAdapters: [] },
  );

  useEffect(() => {
    if (!requestKey || !workerTypeSlug) return;
    let cancelled = false;
    void getSessionImportWorkerRequirement(orgSlug, workerTypeSlug)
      .then((value) => {
        if (cancelled) return;
        setResult({
          error: null,
          key: requestKey,
          requirement: {
            required: value.requiresModelResource,
            protocolAdapters: value.modelProtocolAdapters,
          },
        });
      })
      .catch((error: unknown) => {
        if (cancelled) return;
        setResult({
          error: error instanceof Error ? error.message : "无法加载 Worker 创建选项",
          key: requestKey,
          requirement: null,
        });
      });
    return () => {
      cancelled = true;
    };
  }, [orgSlug, requestKey, workerTypeSlug]);

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
