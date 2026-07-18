"use client";

import { useTranslations } from "next-intl";
import { FormField, FormFieldGroup } from "@/components/ui/form-field";
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
import { WorkerTemplateRuntimeChoices } from "./WorkerTemplateRuntimeChoices";
import type { AsyncState } from "@/components/pod/hooks/workerCreateDraft";
import type { WorkerCreateOptions } from "@/lib/api/facade/podConnect";

export function WorkerTemplateRuntimePanel({
  draft,
  catalog,
  workerOptions,
  onChange,
}: WorkerTemplatePanelProps & {
  workerOptions: AsyncState<WorkerCreateOptions>;
}) {
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
      <WorkerTemplateRuntimeChoices
        draft={draft}
        workerOptions={workerOptions}
        onChange={onChange}
      />
      <FormField
        label={t("fields.deploymentMode")}
        htmlFor="deployment-mode"
        required
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
      <FormField
        label={t("fields.placementPolicy")}
        htmlFor="placement-policy"
      >
        <Select
          value={runtime.placementPolicy}
          onValueChange={(placementPolicy) => setRuntime({ placementPolicy })}
        >
          <SelectTrigger id="placement-policy">
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
      <FormField
        label={t("fields.automationLevel")}
        htmlFor="automation-level"
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
    </FormFieldGroup>
  );
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
