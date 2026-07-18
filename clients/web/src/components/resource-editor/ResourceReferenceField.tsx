"use client";

import { useTranslations } from "next-intl";
import { FormField } from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
} from "@/components/ui/select";
import type { ResourceReference } from "./resource-editor-types";
import {
  isResourceReferenceCatalogReadOnly,
  type ResourceReferenceCatalog,
} from "./resource-reference-options";

const EMPTY_REFERENCE = "__none__";

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
  const options = (catalog.byKind[key] ?? []).filter((option) => option.name);
  const error = catalog.errorsByKind[key] ?? catalog.error;
  const selected = options.find((option) => option.name === value?.name);
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
        <Select
          value={value?.name ?? ""}
          disabled={readOnly}
          onValueChange={(name) => {
            if (name === EMPTY_REFERENCE) {
              onChange(undefined);
              return;
            }
            onChange({
              kind,
              name,
              revision: value?.name === name ? value.revision : undefined,
            });
          }}
        >
          <SelectTrigger
            id={id}
            role="combobox"
            aria-label={label}
            aria-required={required}
          >
            <span className={value?.name ? "truncate" : "truncate text-muted-foreground"}>
              {selected
                ? `${selected.displayName || selected.name} · ${selected.name}`
                : value?.name || label}
            </span>
          </SelectTrigger>
          <SelectContent>
            {!required && (
              <SelectItem value={EMPTY_REFERENCE}>
                {t("collections.none")}
              </SelectItem>
            )}
            {options.map((option) => (
              <SelectItem
                key={option.name}
                value={option.name}
                aria-label={`${option.displayName || option.name} ${option.name}`}
              >
                <span className="flex min-w-0 flex-col">
                  <span className="truncate">
                    {option.displayName || option.name}
                  </span>
                  <span className="truncate text-xs text-muted-foreground">
                    {option.name}
                  </span>
                </span>
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
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
