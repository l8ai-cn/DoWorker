"use client";

import { Plus, Trash2 } from "lucide-react";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import { FormField } from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import { useResourceEditorRowKeys } from "./use-resource-editor-row-keys";

interface ResourceStringMapFieldProps {
  label: string;
  value: Record<string, string>;
  onChange: (value: Record<string, string>) => void;
}

export function ResourceStringMapField({
  label,
  value,
  onChange,
}: ResourceStringMapFieldProps) {
  const t = useTranslations("resourceEditor");
  const entries = Object.entries(value);
  const rows = useResourceEditorRowKeys(entries.length);
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
          onClick={() => {
            rows.appendKey();
            onChange({ ...value, [nextKey(value)]: "" });
          }}
        >
          <Plus className="h-4 w-4" />
        </Button>
      </div>
      {entries.length === 0 && (
        <p className="text-sm text-muted-foreground">{t("collections.none")}</p>
      )}
      {entries.map(([key, entryValue], index) => (
        <div
          key={rows.keys[index]}
          className="grid gap-3 border-l-2 border-border pl-3 sm:grid-cols-[minmax(0,12rem)_minmax(0,1fr)_2.5rem]"
        >
          <FormField label={t("fields.inputKey")} required>
            <Input
              value={key}
              onChange={(event) => {
                onChange(replaceEntry(entries, index, event.target.value, entryValue));
              }}
            />
          </FormField>
          <FormField label={t("fields.inputValue")}>
            <Input
              value={entryValue}
              onChange={(event) => {
                onChange(replaceEntry(entries, index, key, event.target.value));
              }}
            />
          </FormField>
          <Button
            type="button"
            variant="ghost"
            size="icon"
            className="self-start sm:mt-7"
            title={t("collections.remove")}
            aria-label={`${t("collections.remove")} ${key}`}
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

function replaceEntry(
  entries: [string, string][],
  index: number,
  key: string,
  value: string,
): Record<string, string> {
  return Object.fromEntries(entries.map(([entryKey, entryValue], item) =>
    item === index ? [key, value] : [entryKey, entryValue]));
}

function nextKey(value: Record<string, string>): string {
  let index = Object.keys(value).length + 1;
  while (`input-${index}` in value) index++;
  return `input-${index}`;
}
