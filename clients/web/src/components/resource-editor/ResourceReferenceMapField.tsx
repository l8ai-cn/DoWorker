"use client";

import { Plus, Trash2 } from "lucide-react";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import { FormField } from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import type { ResourceReference } from "./resource-editor-types";
import {
  isResourceReferenceCatalogReadOnly,
  type ResourceReferenceCatalog,
} from "./resource-reference-options";
import { ResourceReferenceField } from "./ResourceReferenceField";
import { useResourceEditorRowKeys } from "./use-resource-editor-row-keys";

interface ResourceReferenceMapFieldProps {
  id: string;
  label: string;
  keyLabel: string;
  kind: string;
  value: Record<string, ResourceReference>;
  catalog: ResourceReferenceCatalog;
  onChange: (value: Record<string, ResourceReference>) => void;
}

export function ResourceReferenceMapField({
  id,
  label,
  keyLabel,
  kind,
  value,
  catalog,
  onChange,
}: ResourceReferenceMapFieldProps) {
  const t = useTranslations("resourceEditor");
  const entries = Object.entries(value);
  const rows = useResourceEditorRowKeys(entries.length);
  const readOnly = isResourceReferenceCatalogReadOnly(
    catalog,
    kind,
    entries.map(([, reference]) => reference.name),
  );
  const replace = (
    index: number,
    nextKey: string,
    nextReference: ResourceReference,
  ) => {
    onChange(Object.fromEntries(entries.map(([key, reference], item) =>
      item === index ? [nextKey, nextReference] : [key, reference])));
  };

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between gap-3">
        <h4 className="text-sm font-medium">{label}</h4>
        <Button
          type="button"
          variant="outline"
          size="icon"
          title={t("collections.add")}
          aria-label={`${t("collections.add")} ${label}`}
          disabled={readOnly}
          onClick={() => {
            if (readOnly) return;
            rows.appendKey();
            onChange({
              ...value,
              [nextMapKey(value)]: { kind, name: "" },
            });
          }}
        >
          <Plus className="h-4 w-4" />
        </Button>
      </div>
      {entries.length === 0 && (
        <p className="text-sm text-muted-foreground">
          {t("collections.none")}
        </p>
      )}
      {entries.map(([key, reference], index) => (
        <div
          key={rows.keys[index]}
          className="grid gap-3 border-l-2 border-border pl-3 md:grid-cols-[12rem_minmax(0,1fr)_2.5rem]"
        >
          <FormField
            label={keyLabel}
            htmlFor={`${id}-key-${index}`}
            required
          >
            <Input
              id={`${id}-key-${index}`}
              value={key}
              disabled={readOnly}
              onChange={(event) => {
                if (readOnly) return;
                replace(index, event.target.value, reference);
              }}
            />
          </FormField>
          <ResourceReferenceField
            id={`${id}-ref-${index}`}
            label={t("fields.resourceReference")}
            kind={kind}
            value={reference}
            catalog={catalog}
            required
            onChange={(next) => {
              replace(index, key, next ?? { kind, name: "" });
            }}
          />
          <Button
            type="button"
            variant="ghost"
            size="icon"
            className="self-start md:mt-7"
            title={t("collections.remove")}
            aria-label={`${t("collections.remove")} ${label} ${index + 1}`}
            disabled={readOnly}
            onClick={() => {
              if (readOnly) return;
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

function nextMapKey(value: Record<string, ResourceReference>): string {
  let index = Object.keys(value).length + 1;
  while (`binding-${index}` in value) index++;
  return `binding-${index}`;
}
