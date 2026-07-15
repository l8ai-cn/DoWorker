"use client";

import { Input } from "@/components/ui/input";
import type { WorkerResourceRequest } from "@/lib/api/facade/podConnect";

interface WorkerResourceEditorProps {
  value: WorkerResourceRequest;
  onChange: (value: WorkerResourceRequest) => void;
  t: (key: string) => string;
}

const GIB = 1024 ** 3;

export function WorkerResourceEditor({
  value,
  onChange,
  t,
}: WorkerResourceEditorProps) {
  return (
    <div className="grid gap-3 rounded-md border border-border bg-muted/20 p-3 sm:grid-cols-3">
      <ResourceInput
        label={t("workerCreate.runtime.resources.cpu")}
        unit="核"
        value={value.cpu_request_millicpu / 1000}
        step="0.1"
        min="0.1"
        onChange={(next) => onChange({
          ...value,
          cpu_request_millicpu: toMilliCPU(next),
          cpu_limit_millicpu: toMilliCPU(next),
        })}
      />
      <ResourceInput
        label={t("workerCreate.runtime.resources.memory")}
        unit="GiB"
        value={value.memory_request_bytes / GIB}
        step="0.5"
        min="0.5"
        onChange={(next) => onChange({
          ...value,
          memory_request_bytes: toBytes(next),
          memory_limit_bytes: toBytes(next),
        })}
      />
      <ResourceInput
        label={t("workerCreate.runtime.resources.storage")}
        unit="GiB"
        value={value.storage_request_bytes / GIB}
        step="1"
        min="1"
        onChange={(next) => onChange({
          ...value,
          storage_request_bytes: toBytes(next),
          storage_limit_bytes: toBytes(next),
        })}
      />
    </div>
  );
}

function ResourceInput(props: {
  label: string;
  unit: string;
  value: number;
  step: string;
  min: string;
  onChange: (value: number) => void;
}) {
  return (
    <label className="block text-sm">
      <span className="mb-1 block font-medium">{props.label}</span>
      <span className="flex items-center gap-2">
        <Input
          type="number"
          inputMode="decimal"
          min={props.min}
          step={props.step}
          value={Number.isFinite(props.value) && props.value > 0 ? props.value : ""}
          onChange={(event) => props.onChange(Number(event.target.value))}
          aria-label={props.label}
        />
        <span className="shrink-0 text-xs text-muted-foreground">{props.unit}</span>
      </span>
    </label>
  );
}

function toMilliCPU(value: number): number {
  return Number.isFinite(value) && value > 0 ? Math.round(value * 1000) : 0;
}

function toBytes(value: number): number {
  return Number.isFinite(value) && value > 0 ? Math.round(value * GIB) : 0;
}
