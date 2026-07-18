"use client";

import { Plus, Trash2 } from "lucide-react";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import type { ResourceReference } from "./resource-editor-types";
import {
  isResourceReferenceCatalogReadOnly,
  type ResourceReferenceCatalog,
} from "./resource-reference-options";
import { ResourceReferenceField } from "./ResourceReferenceField";
import { useResourceEditorRowKeys } from "./use-resource-editor-row-keys";

interface ResourceReferenceListFieldProps {
  id: string;
  label: string;
  kind: string;
  catalogKey?: string;
  value: ResourceReference[];
  catalog: ResourceReferenceCatalog;
  onChange: (value: ResourceReference[]) => void;
}

export function ResourceReferenceListField({
  id,
  label,
  kind,
  catalogKey,
  value,
  catalog,
  onChange,
}: ResourceReferenceListFieldProps) {
  const t = useTranslations("resourceEditor");
  const rows = useResourceEditorRowKeys(value.length);
  const readOnly = isResourceReferenceCatalogReadOnly(
    catalog,
    catalogKey ?? kind,
    value.map((reference) => reference.name),
  );
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
            onChange([...value, { kind, name: "" }]);
          }}
        >
          <Plus className="h-4 w-4" />
        </Button>
      </div>
      {value.length === 0 && (
        <p className="text-sm text-muted-foreground">
          {t("collections.none")}
        </p>
      )}
      {value.map((reference, index) => (
        <div
          key={rows.keys[index]}
          className="grid gap-2 border-l-2 border-border pl-3 sm:grid-cols-[minmax(0,1fr)_2.5rem]"
        >
          <ResourceReferenceField
            id={`${id}-${index}`}
            label={`${label} ${index + 1}`}
            kind={kind}
            catalogKey={catalogKey}
            value={reference}
            catalog={catalog}
            required
            onChange={(next) => {
              const refs = [...value];
              refs[index] = next ?? { kind, name: "" };
              onChange(refs);
            }}
          />
          <Button
            type="button"
            variant="ghost"
            size="icon"
            className="self-start sm:mt-7"
            title={t("collections.remove")}
            aria-label={`${t("collections.remove")} ${label} ${index + 1}`}
            disabled={readOnly}
            onClick={() => {
              if (readOnly) return;
              rows.removeKey(index);
              onChange(value.filter((_, item) => item !== index));
            }}
          >
            <Trash2 className="h-4 w-4" />
          </Button>
        </div>
      ))}
    </div>
  );
}
