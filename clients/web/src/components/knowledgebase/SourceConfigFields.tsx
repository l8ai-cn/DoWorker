"use client";

import { FormField } from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import {
  SOURCE_FIELD_DEFS,
  type KBSourceType,
  type SourceConfigForm,
} from "./sourceConfig";

interface SourceConfigFieldsProps {
  sourceType: Exclude<KBSourceType, "git">;
  value: SourceConfigForm;
  onChange: (next: SourceConfigForm) => void;
  idPrefix?: string;
}

export function SourceConfigFields({
  sourceType,
  value,
  onChange,
  idPrefix = "kb-source",
}: SourceConfigFieldsProps) {
  const fields = SOURCE_FIELD_DEFS[sourceType];

  return (
    <div className="space-y-3 rounded-lg border border-border bg-surface-muted/40 p-4">
      <p className="text-sm font-medium">外部数据源凭证</p>
      {fields.map((field) => (
        <FormField
          key={field.key}
          label={field.label}
          htmlFor={`${idPrefix}-${field.key}`}
          hint={field.help}
        >
          <Input
            id={`${idPrefix}-${field.key}`}
            type={field.secret ? "password" : "text"}
            value={value[field.key] ?? ""}
            onChange={(e) => onChange({ ...value, [field.key]: e.target.value })}
            placeholder={
              field.secret && value[field.key] === "***"
                ? "已配置，留空保持不变"
                : field.placeholder
            }
            autoComplete="off"
          />
        </FormField>
      ))}
    </div>
  );
}
