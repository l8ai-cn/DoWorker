"use client";

import { useTranslations } from "next-intl";
import { FormField, FormRow } from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import type {
  WorkerTemplateResources,
  WorkerTemplateRuntime,
} from "./resource-editor-types";
import type { ResourceReferenceCatalog } from "./resource-reference-options";
import { ResourceReferenceField } from "./ResourceReferenceField";

interface WorkerTemplateResourceLimitsFieldProps {
  runtime: WorkerTemplateRuntime;
  catalog: ResourceReferenceCatalog;
  onChange: (patch: Partial<WorkerTemplateRuntime>) => void;
}

const DEFAULT_RESOURCES: WorkerTemplateResources = {
  cpuRequestMillicpu: 500,
  cpuLimitMillicpu: 1000,
  memoryRequestBytes: 536870912,
  memoryLimitBytes: 1073741824,
  storageRequestBytes: 1073741824,
  storageLimitBytes: 10737418240,
};

export function WorkerTemplateResourceLimitsField({
  runtime,
  catalog,
  onChange,
}: WorkerTemplateResourceLimitsFieldProps) {
  const t = useTranslations("resourceEditor");
  const mode = runtime.customResources ? "custom" : "profile";
  const resources = runtime.customResources ?? DEFAULT_RESOURCES;
  const setResource = (
    field: keyof WorkerTemplateResources,
    value: string,
  ) => {
    const next = positiveInteger(value);
    onChange({
      customResources: { ...resources, [field]: next || undefined },
      resourceProfileRef: undefined,
    });
  };
  return (
    <div className="space-y-4">
      <FormField label={t("fields.resourceMode")}>
        <Select
          value={mode}
          onValueChange={(next) => onChange(next === "custom"
            ? {
              resourceProfileRef: undefined,
              customResources: { ...DEFAULT_RESOURCES },
            }
            : { customResources: undefined })}
        >
          <SelectTrigger><SelectValue /></SelectTrigger>
          <SelectContent>
            <SelectItem value="profile">{t("options.profile")}</SelectItem>
            <SelectItem value="custom">{t("options.custom")}</SelectItem>
          </SelectContent>
        </Select>
      </FormField>
      {mode === "profile" ? (
        <ResourceReferenceField
          id="resource-profile-reference"
          label={t("fields.resourceProfileRef")}
          kind="ResourceProfile"
          value={runtime.resourceProfileRef}
          catalog={catalog}
          required
          onChange={(resourceProfileRef) => onChange({ resourceProfileRef })}
        />
      ) : (
        <div className="space-y-3">
          <ResourceRow
            labels={[t("fields.cpuRequest"), t("fields.cpuLimit")]}
            values={[resources.cpuRequestMillicpu, resources.cpuLimitMillicpu]}
            onChange={[
              (value) => setResource("cpuRequestMillicpu", value),
              (value) => setResource("cpuLimitMillicpu", value),
            ]}
          />
          <ResourceRow
            labels={[t("fields.memoryRequest"), t("fields.memoryLimit")]}
            values={[resources.memoryRequestBytes, resources.memoryLimitBytes]}
            onChange={[
              (value) => setResource("memoryRequestBytes", value),
              (value) => setResource("memoryLimitBytes", value),
            ]}
          />
          <ResourceRow
            labels={[t("fields.storageRequest"), t("fields.storageLimit")]}
            values={[resources.storageRequestBytes, resources.storageLimitBytes]}
            onChange={[
              (value) => setResource("storageRequestBytes", value),
              (value) => setResource("storageLimitBytes", value),
            ]}
          />
          <ResourceRow
            labels={[t("fields.gpuRequest"), t("fields.gpuLimit")]}
            values={[resources.gpuRequest ?? 0, resources.gpuLimit ?? 0]}
            onChange={[
              (value) => setResource("gpuRequest", value),
              (value) => setResource("gpuLimit", value),
            ]}
            optional
          />
        </div>
      )}
    </div>
  );
}

function ResourceRow({
  labels,
  values,
  onChange,
  optional,
}: {
  labels: [string, string];
  values: [number, number];
  onChange: [(value: string) => void, (value: string) => void];
  optional?: boolean;
}) {
  return (
    <FormRow>
      {labels.map((label, index) => (
        <FormField key={label} label={label} className="flex-1">
          <Input
            type="number"
            min={optional ? 0 : 1}
            value={values[index] || ""}
            onChange={(event) => onChange[index](event.target.value)}
          />
        </FormField>
      ))}
    </FormRow>
  );
}

function positiveInteger(value: string): number {
  const parsed = Number(value);
  return Number.isSafeInteger(parsed) && parsed > 0 ? parsed : 0;
}
