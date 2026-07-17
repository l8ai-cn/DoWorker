"use client";

import { useTranslations } from "next-intl";
import { FormField } from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import type { ResourceReference } from "./resource-editor-types";
import type { ResourceReferenceCatalog } from "./resource-reference-options";

interface ResourceReferenceFieldProps {
  id: string;
  label: string;
  kind: string;
  value?: ResourceReference;
  catalog: ResourceReferenceCatalog;
  required?: boolean;
  onChange: (value: ResourceReference | undefined) => void;
}

export function ResourceReferenceField({
  id,
  label,
  kind,
  value,
  catalog,
  required,
  onChange,
}: ResourceReferenceFieldProps) {
  const t = useTranslations("resourceEditor");
  const options = catalog.byKind[kind] ?? [];
  const hint = catalog.loading
    ? t("references.loading")
    : catalog.error
      ? catalog.error
      : options.length === 0
        ? t("references.empty", { kind })
        : t("references.available", { count: options.length });
  return (
    <FormField
      label={label}
      htmlFor={id}
      required={required}
      hint={hint}
      error={catalog.error ?? undefined}
    >
      <div className="grid gap-2 sm:grid-cols-[minmax(0,1fr)_7rem]">
        <Input
          id={id}
          list={`${id}-options`}
          value={value?.name ?? ""}
          onChange={(event) => {
            const name = event.target.value;
            onChange(name ? { ...value, kind, name } : undefined);
          }}
        />
        <datalist id={`${id}-options`}>
          {options.filter((option) => option.name).map((option) => (
            <option
              key={option.name}
              value={option.name}
              label={option.displayName || `r${option.revision}`}
            />
          ))}
        </datalist>
        <Input
          type="number"
          min={1}
          aria-label={`${label} ${t("fields.revision")}`}
          placeholder={t("fields.revision")}
          value={value?.revision ?? ""}
          disabled={!value?.name}
          onChange={(event) => {
            if (!value) return;
            const revision = Number(event.target.value);
            onChange({
              ...value,
              revision: Number.isSafeInteger(revision) && revision > 0
                ? revision
                : undefined,
            });
          }}
        />
      </div>
    </FormField>
  );
}
