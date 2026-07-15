"use client";

import { useEffect } from "react";
import { useTranslations } from "next-intl";
import { AlertMessage } from "@/components/ui/alert-message";
import { FormField, FormRow } from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Spinner } from "@/components/ui/spinner";
import { useWorkerCreateOptions } from "@/components/pod/hooks/useWorkerCreateOptions";
import type { WorkerTemplatePanelProps } from "./worker-template-panel-props";
import { WorkerTemplateOptionSelectField } from "./WorkerTemplateOptionSelectField";
import {
  selectWorkerTemplateType,
  synchronizeWorkerTemplateRuntime,
  workerTemplateRuntimeImageOptions,
  workerTemplateTypeOptions,
} from "./worker-template-runtime-options";

export function WorkerTemplateRuntimeChoices({
  draft,
  onChange,
}: Pick<WorkerTemplatePanelProps, "draft" | "onChange">) {
  const t = useTranslations("resourceEditor");
  const runtimeT = useTranslations("workerCreate");
  const options = useWorkerCreateOptions(true, {
    workerTypeSlug: "",
    computeTargetId: 0,
    deploymentMode: "",
  });

  useEffect(() => {
    if (options.status !== "ready") return;
    const next = synchronizeWorkerTemplateRuntime(draft, options.data);
    if (next) onChange(next);
  }, [draft, onChange, options]);

  if (options.status === "idle" || options.status === "loading") {
    return <Spinner className="my-2" />;
  }
  if (options.status === "error") {
    return <AlertMessage type="error" message={options.error} />;
  }
  const selectedType = options.data.worker_types.find(
    (option) => option.slug === draft.spec.workerType,
  );
  const interactionModes = selectedType?.supported_interaction_modes ?? [];

  return (
    <>
      <FormRow>
        <WorkerTemplateOptionSelectField
          id="worker-type"
          label={t("fields.workerType")}
          value={draft.spec.workerType}
          options={workerTemplateTypeOptions(options.data)}
          onChange={(workerType) => {
            const next = selectWorkerTemplateType(
              draft,
              options.data,
              workerType,
            );
            if (next) onChange(next);
          }}
        />
        <WorkerTemplateOptionSelectField
          id="runtime-image"
          label={runtimeT("runtime.runtimeImage")}
          value={numberValue(draft.spec.runtime.runtimeImageId)}
          options={workerTemplateRuntimeImageOptions(
            options.data,
            draft.spec.workerType,
          )}
          disabled={!selectedType?.selectable}
          onChange={(value) => onChange({
            ...draft,
            spec: {
              ...draft.spec,
              runtime: { ...draft.spec.runtime, runtimeImageId: Number(value) },
            },
          })}
        />
      </FormRow>
      <FormRow>
        <FormField
          label={t("fields.optionsRevision")}
          htmlFor="options-revision"
          required
          className="flex-1"
        >
          <Input
            id="options-revision"
            value={draft.spec.optionsRevision}
            readOnly
            disabled
          />
        </FormField>
        <FormField
          label={t("fields.interactionMode")}
          htmlFor="interaction-mode"
          required
          className="flex-1"
        >
          <Select
            value={draft.spec.typeConfig.interactionMode}
            disabled={!selectedType?.selectable}
            onValueChange={(interactionMode) => onChange({
              ...draft,
              spec: {
                ...draft.spec,
                typeConfig: {
                  ...draft.spec.typeConfig,
                  interactionMode,
                },
              },
            })}
          >
            <SelectTrigger id="interaction-mode"><SelectValue /></SelectTrigger>
            <SelectContent>
              {interactionModes.map((mode) => (
                <SelectItem key={mode} value={mode}>
                  {t(`options.${mode}`)}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </FormField>
      </FormRow>
    </>
  );
}

function numberValue(value: number): string {
  return value > 0 ? String(value) : "";
}
