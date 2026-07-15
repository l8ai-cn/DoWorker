"use client";

import { useTranslations } from "next-intl";
import { FormField, FormFieldGroup, FormRow } from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import type {
  WorkerTemplateRuntime,
} from "./resource-editor-types";
import { ResourceReferenceField } from "./ResourceReferenceField";
import { WorkerTemplateResourceLimitsField } from "./WorkerTemplateResourceLimitsField";
import type { WorkerTemplatePanelProps } from "./worker-template-panel-props";

export function WorkerTemplateRuntimePanel({
  draft,
  catalog,
  onChange,
}: WorkerTemplatePanelProps) {
  const t = useTranslations("resourceEditor");
  const runtime = draft.spec.runtime;
  const setRuntime = (patch: Partial<WorkerTemplateRuntime>) => {
    onChange({
      ...draft,
      spec: {
        ...draft.spec,
        runtime: { ...runtime, ...patch },
      },
    });
  };

  return (
    <FormFieldGroup
      title={t("sections.runtime")}
      className="border-t border-border pt-6"
    >
      <FormRow>
        <FormField
          label={t("fields.runtimeImageId")}
          htmlFor="runtime-image-id"
          required
          className="flex-1"
        >
          <Input
            id="runtime-image-id"
            type="number"
            min={1}
            value={runtime.runtimeImageId || ""}
            onChange={(event) => setRuntime({
              runtimeImageId: positiveInteger(event.target.value),
            })}
          />
        </FormField>
        <FormField
          label={t("fields.deploymentMode")}
          htmlFor="deployment-mode"
          required
          className="flex-1"
        >
          <Select
            value={runtime.deploymentMode}
            onValueChange={(deploymentMode) => setRuntime({ deploymentMode })}
          >
            <SelectTrigger id="deployment-mode"><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="pooled">{t("options.pooled")}</SelectItem>
              <SelectItem value="dedicated">{t("options.dedicated")}</SelectItem>
            </SelectContent>
          </Select>
        </FormField>
      </FormRow>
      <FormField label={t("fields.placementPolicy")}>
        <Select
          value={runtime.placementPolicy}
          onValueChange={(placementPolicy) => setRuntime({ placementPolicy })}
        >
          <SelectTrigger>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="automatic">{t("options.automatic")}</SelectItem>
            <SelectItem value="explicit">{t("options.explicit")}</SelectItem>
          </SelectContent>
        </Select>
      </FormField>
      <ResourceReferenceField
        id="compute-target-reference"
        label={t("fields.computeTargetRef")}
        kind="ComputeTarget"
        value={runtime.computeTargetRef}
        catalog={catalog}
        required
        onChange={(computeTargetRef) => setRuntime({
          computeTargetRef: computeTargetRef ?? {
            kind: "ComputeTarget",
            name: "",
          },
        })}
      />
      <WorkerTemplateResourceLimitsField
        runtime={runtime}
        catalog={catalog}
        onChange={setRuntime}
      />
      <FormRow>
        <FormField
          label={t("fields.interactionMode")}
          htmlFor="interaction-mode"
          className="flex-1"
        >
          <Select
            value={draft.spec.typeConfig.interactionMode}
            onValueChange={(interactionMode) => setTypeConfig(
              draft,
              onChange,
              { interactionMode },
            )}
          >
            <SelectTrigger id="interaction-mode"><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="pty">{t("options.pty")}</SelectItem>
              <SelectItem value="acp">{t("options.acp")}</SelectItem>
            </SelectContent>
          </Select>
        </FormField>
        <FormField
          label={t("fields.automationLevel")}
          htmlFor="automation-level"
          className="flex-1"
        >
          <Select
            value={draft.spec.typeConfig.automationLevel}
            onValueChange={(automationLevel) => setTypeConfig(
              draft,
              onChange,
              { automationLevel },
            )}
          >
            <SelectTrigger id="automation-level"><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="interactive">{t("options.interactive")}</SelectItem>
              <SelectItem value="auto_edit">{t("options.autoEdit")}</SelectItem>
              <SelectItem value="autonomous">{t("options.autonomous")}</SelectItem>
            </SelectContent>
          </Select>
        </FormField>
      </FormRow>
    </FormFieldGroup>
  );
}

function positiveInteger(value: string): number {
  const parsed = Number(value);
  return Number.isSafeInteger(parsed) && parsed > 0 ? parsed : 0;
}

function setTypeConfig(
  draft: WorkerTemplatePanelProps["draft"],
  onChange: WorkerTemplatePanelProps["onChange"],
  patch: Partial<WorkerTemplatePanelProps["draft"]["spec"]["typeConfig"]>,
) {
  onChange({
    ...draft,
    spec: {
      ...draft.spec,
      typeConfig: { ...draft.spec.typeConfig, ...patch },
    },
  });
}
