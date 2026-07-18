"use client";

import { Plus, Trash2 } from "lucide-react";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import { FormField } from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";
import type { PromptVariableDraft } from "./resource-editor-types";
import { useResourceEditorRowKeys } from "./use-resource-editor-row-keys";

interface PromptVariablesFieldProps {
  value: Record<string, PromptVariableDraft>;
  onChange: (value: Record<string, PromptVariableDraft>) => void;
}

export function PromptVariablesField({
  value,
  onChange,
}: PromptVariablesFieldProps) {
  const t = useTranslations("resourceEditor");
  const entries = Object.entries(value);
  const rows = useResourceEditorRowKeys(entries.length);
  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between gap-3">
        <h4 className="text-sm font-medium">{t("fields.variables")}</h4>
        <Button
          type="button"
          variant="outline"
          size="icon"
          title={t("collections.add")}
          aria-label={`${t("collections.add")} ${t("fields.variables")}`}
          onClick={() => {
            rows.appendKey();
            onChange({
              ...value,
              [nextVariableName(value)]: { required: false },
            });
          }}
        >
          <Plus className="h-4 w-4" />
        </Button>
      </div>
      {entries.length === 0 && (
        <p className="text-sm text-muted-foreground">{t("collections.none")}</p>
      )}
      {entries.map(([name, variable], index) => (
        <div
          key={rows.keys[index]}
          className="grid gap-3 border-l-2 border-border pl-3 lg:grid-cols-[minmax(0,12rem)_7rem_8rem_minmax(0,1fr)_2.5rem]"
        >
          <FormField label={t("fields.variableName")} required>
            <Input
              value={name}
              onChange={(event) => onChange(replaceVariable(
                entries,
                index,
                event.target.value,
                variable,
              ))}
            />
          </FormField>
          <BooleanField
            label={t("fields.required")}
            checked={variable.required}
            onChange={(required) => onChange(replaceVariable(
              entries,
              index,
              name,
              { ...variable, required },
            ))}
          />
          <BooleanField
            label={t("fields.hasDefault")}
            checked={variable.default !== undefined}
            onChange={(enabled) => onChange(replaceVariable(
              entries,
              index,
              name,
              enabled
                ? { ...variable, default: variable.default ?? "" }
                : { required: variable.required },
            ))}
          />
          <FormField
            label={t("fields.defaultValue")}
            disabled={variable.default === undefined}
          >
            <Input
              disabled={variable.default === undefined}
              value={variable.default ?? ""}
              onChange={(event) => onChange(replaceVariable(
                entries,
                index,
                name,
                { ...variable, default: event.target.value },
              ))}
            />
          </FormField>
          <Button
            type="button"
            variant="ghost"
            size="icon"
            className="self-start lg:mt-7"
            title={t("collections.remove")}
            aria-label={`${t("collections.remove")} ${name}`}
            onClick={() => {
              rows.removeKey(index);
              onChange(Object.fromEntries(
                entries.filter((_, item) => item !== index),
              ));
            }}
          >
            <Trash2 className="h-4 w-4" />
          </Button>
        </div>
      ))}
    </div>
  );
}

function BooleanField({
  label,
  checked,
  onChange,
}: {
  label: string;
  checked: boolean;
  onChange: (checked: boolean) => void;
}) {
  return (
    <FormField label={label}>
      <Switch
        className="mt-1"
        checked={checked}
        onCheckedChange={onChange}
        aria-label={label}
      />
    </FormField>
  );
}

function replaceVariable(
  entries: [string, PromptVariableDraft][],
  index: number,
  name: string,
  variable: PromptVariableDraft,
) {
  return Object.fromEntries(entries.map(([key, value], item) =>
    item === index ? [name, variable] : [key, value]));
}

function nextVariableName(value: Record<string, PromptVariableDraft>): string {
  let index = Object.keys(value).length + 1;
  while (`variable-${index}` in value) index++;
  return `variable-${index}`;
}
