"use client";

import { useTranslations } from "next-intl";
import {
  FormField,
  FormFieldGroup,
  FormRow,
} from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";
import type { WorkflowDraft } from "./resource-editor-types";
import { ResourceEnumField } from "./ResourceEnumField";

interface WorkflowExecutionFieldsProps {
  draft: WorkflowDraft;
  onChange: (draft: WorkflowDraft) => void;
}

export function WorkflowExecutionFields({
  draft,
  onChange,
}: WorkflowExecutionFieldsProps) {
  const t = useTranslations("resourceEditor");
  const setSpec = (patch: Partial<WorkflowDraft["spec"]>) => {
    onChange({ ...draft, spec: { ...draft.spec, ...patch } });
  };
  const number = (
    field: keyof Pick<WorkflowDraft["spec"],
      "maxConcurrentRuns" | "maxRetainedRuns" |
      "timeoutMinutes" | "idleTimeoutSeconds">,
    value: string,
  ) => {
    const parsed = Number(value);
    setSpec({ [field]: Number.isSafeInteger(parsed) ? parsed : 0 });
  };
  return (
    <FormFieldGroup title={t("sections.execution")}>
      <FormRow>
        <div className="flex-1">
          <ResourceEnumField
            id="workflow-execution-mode"
            label={t("fields.executionMode")}
            value={draft.spec.executionMode}
            options={[
              { value: "direct", label: t("options.direct") },
              { value: "autopilot", label: t("options.autopilot") },
            ]}
            onChange={(executionMode) => setSpec({ executionMode })}
          />
        </div>
        <div className="flex-1">
          <ResourceEnumField
            id="workflow-sandbox-strategy"
            label={t("fields.sandboxStrategy")}
            value={draft.spec.sandboxStrategy}
            options={[
              { value: "fresh", label: t("options.fresh") },
              { value: "persistent", label: t("options.persistent") },
            ]}
            onChange={(sandboxStrategy) => setSpec({ sandboxStrategy })}
          />
        </div>
      </FormRow>
      <FormRow>
        <FormField
          label={t("fields.concurrencyPolicy")}
          htmlFor="workflow-concurrency-policy"
          className="flex-1"
        >
          <Input
            id="workflow-concurrency-policy"
            value={draft.spec.concurrencyPolicy}
            disabled
          />
        </FormField>
        <FormField
          label={t("fields.sessionPersistence")}
          className="flex-1"
        >
          <Switch
            checked={draft.spec.sessionPersistence}
            onCheckedChange={(sessionPersistence) => setSpec({
              sessionPersistence,
            })}
            aria-label={t("fields.sessionPersistence")}
          />
        </FormField>
      </FormRow>
      <FormRow>
        <NumberField
          id="workflow-max-concurrent"
          label={t("fields.maxConcurrentRuns")}
          value={draft.spec.maxConcurrentRuns}
          min={1}
          max={100}
          onChange={(value) => number("maxConcurrentRuns", value)}
        />
        <NumberField
          id="workflow-max-retained"
          label={t("fields.maxRetainedRuns")}
          value={draft.spec.maxRetainedRuns}
          min={0}
          max={10000}
          onChange={(value) => number("maxRetainedRuns", value)}
        />
      </FormRow>
      <FormRow>
        <NumberField
          id="workflow-timeout"
          label={t("fields.timeoutMinutes")}
          value={draft.spec.timeoutMinutes}
          min={1}
          max={1440}
          onChange={(value) => number("timeoutMinutes", value)}
        />
        <NumberField
          id="workflow-idle-timeout"
          label={t("fields.idleTimeoutSeconds")}
          value={draft.spec.idleTimeoutSeconds}
          min={1}
          max={86400}
          onChange={(value) => number("idleTimeoutSeconds", value)}
        />
      </FormRow>
      <FormField label={t("fields.cronExpression")} htmlFor="workflow-cron">
        <Input
          id="workflow-cron"
          value={draft.spec.cronExpression ?? ""}
          onChange={(event) => setSpec({
            cronExpression: event.target.value,
          })}
        />
      </FormField>
      <FormField label={t("fields.callbackUrl")} htmlFor="workflow-callback">
        <Input
          id="workflow-callback"
          type="url"
          value={draft.spec.callbackUrl ?? ""}
          onChange={(event) => setSpec({ callbackUrl: event.target.value })}
        />
      </FormField>
    </FormFieldGroup>
  );
}

function NumberField({
  id,
  label,
  value,
  min,
  max,
  onChange,
}: {
  id: string;
  label: string;
  value: number;
  min: number;
  max: number;
  onChange: (value: string) => void;
}) {
  return (
    <FormField label={label} htmlFor={id} className="flex-1">
      <Input
        id={id}
        type="number"
        min={min}
        max={max}
        value={value}
        onChange={(event) => onChange(event.target.value)}
      />
    </FormField>
  );
}
