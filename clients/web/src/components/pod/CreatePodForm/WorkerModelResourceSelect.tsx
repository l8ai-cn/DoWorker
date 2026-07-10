"use client";

import { AlertCircle, Database, KeyRound } from "lucide-react";
import { cn } from "@/lib/utils";
import { Spinner } from "@/components/ui/spinner";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
} from "@/components/ui/select";
import type { EffectiveResource } from "@/lib/api/facade/aiResource";
import { modelResourceLabel } from "./workerModelResources";

interface Props {
  resources: EffectiveResource[];
  selectedResourceId: number | null;
  onSelect: (id: number | null) => void;
  loading: boolean;
  error: string | null;
  validationError?: string;
  t: (key: string) => string;
}

export function WorkerModelResourceSelect({
  resources,
  selectedResourceId,
  onSelect,
  loading,
  error,
  validationError,
  t,
}: Props) {
  const selected = resources.find((r) => r.resource?.id === selectedResourceId);

  if (loading) {
    return (
      <div className="flex items-center text-sm text-muted-foreground py-2">
        <Spinner size="sm" className="mr-2" />
        {t("common.loading")}
      </div>
    );
  }

  if (error) {
    return (
      <div role="alert" className="rounded-md border border-destructive/30 bg-destructive/10 p-3">
        <div className="flex gap-2 text-sm text-destructive">
          <AlertCircle className="mt-0.5 h-4 w-4 shrink-0" />
          <span>{error}</span>
        </div>
      </div>
    );
  }

  if (resources.length === 0) {
    return (
      <div className="rounded-md border border-dashed border-border p-3">
        <div className="flex gap-2 text-sm text-muted-foreground">
          <Database className="mt-0.5 h-4 w-4 shrink-0" />
          <span>{t("ide.createPod.noModelResourcesAvailableHint")}</span>
        </div>
      </div>
    );
  }

  return (
    <div>
      <label htmlFor="worker-model-resource-select" className="block text-sm font-medium mb-2">
        {t("ide.createPod.selectModelResource")}
      </label>
      <Select
        value={selectedResourceId ? String(selectedResourceId) : ""}
        onValueChange={(value) => onSelect(value ? Number(value) : null)}
      >
        <SelectTrigger id="worker-model-resource-select">
          <span className={cn(!selected && "text-muted-foreground")}>
            {selected ? modelResourceLabel(selected) : t("ide.createPod.selectModelResourcePlaceholder")}
          </span>
        </SelectTrigger>
        <SelectContent>
          {resources.map((item) => {
            const id = item.resource?.id;
            if (!id) return null;
            return (
              <SelectItem key={id} value={String(id)}>
                <div className="flex items-center gap-2">
                  <KeyRound className="h-3.5 w-3.5 text-muted-foreground" />
                  <span>{modelResourceLabel(item)}</span>
                </div>
              </SelectItem>
            );
          })}
        </SelectContent>
      </Select>
      <p
        role={validationError ? "alert" : undefined}
        className={cn("mt-1 text-xs", validationError ? "text-destructive" : "text-muted-foreground")}
      >
        {validationError || t("ide.createPod.modelResourceHint")}
      </p>
    </div>
  );
}
