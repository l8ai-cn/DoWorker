"use client";

import { Plus, Trash2 } from "lucide-react";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import { FormField } from "@/components/ui/form-field";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import type { WorkerTemplateKnowledgeMount } from "./resource-editor-types";
import type { ResourceReferenceCatalog } from "./resource-reference-options";
import { ResourceReferenceField } from "./ResourceReferenceField";
import { useResourceEditorRowKeys } from "./use-resource-editor-row-keys";

interface WorkerTemplateKnowledgeMountsFieldProps {
  value: WorkerTemplateKnowledgeMount[];
  catalog: ResourceReferenceCatalog;
  onChange: (value: WorkerTemplateKnowledgeMount[]) => void;
}

export function WorkerTemplateKnowledgeMountsField({
  value,
  catalog,
  onChange,
}: WorkerTemplateKnowledgeMountsFieldProps) {
  const t = useTranslations("resourceEditor");
  const rows = useResourceEditorRowKeys(value.length);
  const replace = (
    index: number,
    next: WorkerTemplateKnowledgeMount,
  ) => {
    onChange(value.map((mount, item) => item === index ? next : mount));
  };
  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between gap-3">
        <h4 className="text-sm font-medium">{t("fields.knowledgeMounts")}</h4>
        <Button
          type="button"
          variant="outline"
          size="icon"
          title={t("collections.add")}
          aria-label={`${t("collections.add")} ${t("fields.knowledgeMounts")}`}
          onClick={() => {
            rows.appendKey();
            onChange([
              ...value,
              { ref: { kind: "KnowledgeBase", name: "" }, mode: "ro" },
            ]);
          }}
        >
          <Plus className="h-4 w-4" />
        </Button>
      </div>
      {value.length === 0 && (
        <p className="text-sm text-muted-foreground">{t("collections.none")}</p>
      )}
      {value.map((mount, index) => (
        <div
          key={rows.keys[index]}
          className="grid gap-3 border-l-2 border-border pl-3 md:grid-cols-[minmax(0,1fr)_8rem_2.5rem]"
        >
          <ResourceReferenceField
            id={`knowledge-mount-${index}`}
            label={`${t("fields.knowledgeBase")} ${index + 1}`}
            kind="KnowledgeBase"
            value={mount.ref}
            catalog={catalog}
            required
            onChange={(ref) => replace(index, {
              ...mount,
              ref: ref ?? { kind: "KnowledgeBase", name: "" },
            })}
          />
          <FormField label={t("fields.mountMode")}>
            <Select
              value={mount.mode}
              onValueChange={(mode) => replace(index, { ...mount, mode })}
            >
              <SelectTrigger><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectItem value="ro">{t("options.readOnly")}</SelectItem>
                <SelectItem value="rw">{t("options.readWrite")}</SelectItem>
              </SelectContent>
            </Select>
          </FormField>
          <Button
            type="button"
            variant="ghost"
            size="icon"
            className="self-start md:mt-7"
            title={t("collections.remove")}
            aria-label={`${t("collections.remove")} ${index + 1}`}
            onClick={() => {
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
