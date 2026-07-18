"use client";

import { useTranslations } from "next-intl";
import { FormField } from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";
import type { ResourceReference } from "./resource-editor-types";
import {
  isResourceReferenceCatalogReadOnly,
  type ResourceReferenceCatalog,
} from "./resource-reference-options";

interface ResourceReferenceFieldProps {
  id: string;
  label: string;
  kind: string;
  catalogKey?: string;
  value?: ResourceReference;
  catalog: ResourceReferenceCatalog;
  required?: boolean;
  onChange: (value: ResourceReference | undefined) => void;
}

export function ResourceReferenceField({
  id,
  label,
  kind,
  catalogKey,
  value,
  catalog,
  required,
  onChange,
}: ResourceReferenceFieldProps) {
  const t = useTranslations("resourceEditor");
  const key = catalogKey ?? kind;
  const options = catalog.byKind[key] ?? [];
  const error = catalog.errorsByKind[key] ?? catalog.error;
  const resolved = !value?.name ||
    options.some((option) => option.name === value.name);
  const readOnly = isResourceReferenceCatalogReadOnly(
    catalog,
    key,
    value?.name ? [value.name] : [],
  );
  const hint = catalog.loading
    ? t("references.loading")
    : error
      ? error
      : options.length === 0
        ? t("references.empty", { kind })
        : t("references.available", { count: options.length });
  return (
    <FormField
      label={label}
      htmlFor={id}
      required={required}
      hint={hint}
      error={error ?? undefined}
    >
      <div className="grid gap-2 sm:grid-cols-[minmax(0,1fr)_7rem]">
        <select
          id={id}
          required={required}
          aria-required={required}
          aria-readonly={readOnly}
          disabled={readOnly}
          className={cn(
            "h-9 w-full rounded-md bg-surface-raised px-3 text-sm",
            "ring-1 ring-border/35 focus-visible:outline-none",
            "focus-visible:ring-2 focus-visible:ring-ring/35",
            readOnly && "cursor-not-allowed bg-muted/40 text-muted-foreground",
          )}
          value={value?.name ?? ""}
          onChange={(event) => {
            if (readOnly) return;
            const name = event.target.value;
            if (!name) {
              onChange(undefined);
              return;
            }
            const identityChanged = value?.kind !== kind || value.name !== name;
            onChange(identityChanged
              ? { ...value, kind, name, revision: undefined }
              : { ...value, kind, name });
          }}
        >
          <option value="">{label}</option>
          {!resolved && value?.name && (
            <option value={value.name}>{value.name}</option>
          )}
          {options.filter((option) => option.name).map((option) => (
            <option
              key={option.name}
              value={option.name}
            >
              {option.displayName || option.name} (r{option.revision})
            </option>
          ))}
        </select>
        <Input
          id={`${id}-revision`}
          type="number"
          min={1}
          aria-label={`${label} ${t("fields.revision")}`}
          placeholder={t("fields.revision")}
          value={value?.revision ?? ""}
          disabled={!value?.name}
          readOnly={readOnly}
          aria-readonly={readOnly}
          onChange={(event) => {
            if (readOnly || !value) return;
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
