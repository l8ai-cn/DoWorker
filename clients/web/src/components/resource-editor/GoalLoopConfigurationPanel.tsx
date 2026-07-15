"use client";

import { useTranslations } from "next-intl";
import {
  FormField,
  FormFieldGroup,
  FormRow,
} from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import type { GoalLoopDraft } from "./resource-editor-types";
import { GoalLoopAcceptanceCriteriaField } from "./GoalLoopAcceptanceCriteriaField";
import { GoalLoopExecutionNumberField } from "./GoalLoopExecutionNumberField";
import { ResourceEnumField } from "./ResourceEnumField";
import { ResourceIdentityFields } from "./ResourceIdentityFields";
import { ResourceReferenceField } from "./ResourceReferenceField";
import { useResourceReferenceOptions } from "./use-resource-reference-options";

interface GoalLoopConfigurationPanelProps {
  orgSlug: string;
  draft: GoalLoopDraft;
  onChange: (draft: GoalLoopDraft) => void;
}

type RequiredNumberField =
  | "maxIterations"
  | "timeoutMinutes"
  | "noProgressLimit"
  | "sameErrorLimit";

export function GoalLoopConfigurationPanel({
  orgSlug,
  draft,
  onChange,
}: GoalLoopConfigurationPanelProps) {
  const t = useTranslations("resourceEditor");
  const catalog = useResourceReferenceOptions(orgSlug);
  const setSpec = (patch: Partial<GoalLoopDraft["spec"]>) => {
    onChange({ ...draft, spec: { ...draft.spec, ...patch } });
  };
  const setNumber = (field: RequiredNumberField, value: string) => {
    setSpec({ [field]: Number(value) });
  };

  return (
    <div className="space-y-6">
      <ResourceIdentityFields
        metadata={draft.metadata}
        onChange={(metadata) => onChange({ ...draft, metadata })}
      />
      <FormFieldGroup title={t("sections.definition")}>
        <ResourceReferenceField
          id="goal-loop-worker-template"
          label={t("fields.workerTemplateRef")}
          kind="WorkerTemplate"
          value={draft.spec.workerTemplateRef}
          catalog={catalog}
          required
          onChange={(workerTemplateRef) => {
            if (workerTemplateRef) setSpec({ workerTemplateRef });
          }}
        />
        <FormField
          label={t("fields.description")}
          htmlFor="goal-loop-description"
        >
          <Textarea
            id="goal-loop-description"
            value={draft.spec.description}
            onChange={(event) => setSpec({ description: event.target.value })}
          />
        </FormField>
        <FormField
          label={t("fields.objective")}
          htmlFor="goal-loop-objective"
          required
        >
          <Textarea
            id="goal-loop-objective"
            className="min-h-28"
            value={draft.spec.objective}
            onChange={(event) => setSpec({ objective: event.target.value })}
          />
        </FormField>
        <GoalLoopAcceptanceCriteriaField
          value={draft.spec.acceptanceCriteria}
          onChange={(acceptanceCriteria) => setSpec({ acceptanceCriteria })}
        />
        <FormField
          label={t("fields.verificationCommand")}
          htmlFor="goal-loop-verification-command"
          required
        >
          <Input
            id="goal-loop-verification-command"
            className="font-mono"
            value={draft.spec.verificationCommand}
            onChange={(event) => setSpec({
              verificationCommand: event.target.value,
            })}
          />
        </FormField>
      </FormFieldGroup>
      <FormFieldGroup title={t("sections.execution")}>
        <FormRow>
          <GoalLoopExecutionNumberField
            id="goal-loop-max-iterations"
            label={t("fields.maxIterations")}
            value={draft.spec.maxIterations}
            min={1}
            max={100}
            onChange={(value) => setNumber("maxIterations", value)}
          />
          <GoalLoopExecutionNumberField
            id="goal-loop-token-budget"
            label={t("fields.tokenBudget")}
            value={draft.spec.tokenBudget ?? ""}
            min={1}
            onChange={(value) => setSpec({
              tokenBudget: value === "" ? undefined : Number(value),
            })}
          />
        </FormRow>
        <FormRow>
          <GoalLoopExecutionNumberField
            id="goal-loop-timeout"
            label={t("fields.timeoutMinutes")}
            value={draft.spec.timeoutMinutes}
            min={1}
            max={1440}
            onChange={(value) => setNumber("timeoutMinutes", value)}
          />
          <GoalLoopExecutionNumberField
            id="goal-loop-no-progress-limit"
            label={t("fields.noProgressLimit")}
            value={draft.spec.noProgressLimit}
            min={1}
            max={20}
            onChange={(value) => setNumber("noProgressLimit", value)}
          />
        </FormRow>
        <FormRow>
          <GoalLoopExecutionNumberField
            id="goal-loop-same-error-limit"
            label={t("fields.sameErrorLimit")}
            value={draft.spec.sameErrorLimit}
            min={1}
            max={20}
            onChange={(value) => setNumber("sameErrorLimit", value)}
          />
          <div className="flex-1">
            <ResourceEnumField
              id="goal-loop-escalation-policy"
              label={t("fields.escalationPolicy")}
              value={draft.spec.escalationPolicy}
              options={[
                { value: "pause", label: t("options.pause") },
                { value: "fail", label: t("options.fail") },
              ]}
              onChange={(escalationPolicy) => setSpec({
                escalationPolicy: escalationPolicy as "pause" | "fail",
              })}
            />
          </div>
        </FormRow>
      </FormFieldGroup>
    </div>
  );
}
