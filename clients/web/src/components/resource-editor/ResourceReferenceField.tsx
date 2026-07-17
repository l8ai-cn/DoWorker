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
import type { ResourceReferenceCatalog } from "./resource-reference-options";

const EMPTY_REFERENCE = "__none__";

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
  const options = (catalog.byKind[kind] ?? []).filter((option) => option.name);
  const selected = options.find((option) => option.name === value?.name);
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
        <Select
          value={value?.name ?? ""}
          disabled={catalog.loading || Boolean(catalog.error) || options.length === 0}
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
          <SelectTrigger id={id} role="combobox" aria-label={label}>
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
