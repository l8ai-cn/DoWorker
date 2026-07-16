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
import type { WorkerTemplatePanelProps } from "./worker-template-panel-props";

export function WorkerTemplateLifecyclePanel({
  draft,
  onChange,
}: WorkerTemplatePanelProps) {
  const t = useTranslations("resourceEditor");
  const lifecycle = draft.spec.lifecycle;
  const update = (patch: Partial<typeof lifecycle>) => {
    onChange({
      ...draft,
      spec: {
        ...draft.spec,
        lifecycle: { ...lifecycle, ...patch },
      },
    });
  };
  return (
    <FormFieldGroup
      title={t("sections.lifecycle")}
      className="border-t border-border pt-6"
    >
      <FormRow>
        <FormField
          label={t("fields.terminationPolicy")}
          htmlFor="termination-policy"
          className="flex-1"
        >
          <Select
            value={lifecycle.terminationPolicy}
            onValueChange={(terminationPolicy) => update({
              terminationPolicy,
            })}
          >
            <SelectTrigger id="termination-policy"><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="manual">{t("options.manual")}</SelectItem>
              <SelectItem value="idle">{t("options.idle")}</SelectItem>
              <SelectItem value="completed">{t("options.completed")}</SelectItem>
            </SelectContent>
          </Select>
        </FormField>
        <FormField
          label={t("fields.idleTimeoutMinutes")}
          htmlFor="idle-timeout"
          className="flex-1"
        >
          <Input
            id="idle-timeout"
            type="number"
            min={0}
            value={lifecycle.idleTimeoutMinutes}
            onChange={(event) => update({
              idleTimeoutMinutes: nonNegativeInteger(event.target.value),
            })}
          />
        </FormField>
      </FormRow>
    </FormFieldGroup>
  );
}

function nonNegativeInteger(value: string): number {
  const parsed = Number(value);
  return Number.isSafeInteger(parsed) && parsed >= 0 ? parsed : 0;
}
