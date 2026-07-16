"use client";

import { Plus, Trash2 } from "lucide-react";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import { FormField } from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

interface WorkerTemplateValuesFieldProps {
  value: Record<string, unknown>;
  onChange: (value: Record<string, unknown>) => void;
}

export function WorkerTemplateValuesField({
  value,
  onChange,
}: WorkerTemplateValuesFieldProps) {
  const t = useTranslations("resourceEditor");
  const entries = Object.entries(value);
  const replace = (index: number, key: string, next: unknown) => {
    onChange(Object.fromEntries(entries.map(([entryKey, entryValue], item) =>
      item === index ? [key, next] : [entryKey, entryValue])));
  };
  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between gap-3">
        <h4 className="text-sm font-medium">{t("fields.configValues")}</h4>
        <Button
          type="button"
          variant="outline"
          size="icon"
          title={t("collections.add")}
          aria-label={`${t("collections.add")} ${t("fields.configValues")}`}
          onClick={() => onChange({ ...value, [nextKey(value)]: "" })}
        >
          <Plus className="h-4 w-4" />
        </Button>
      </div>
      {entries.length === 0 && (
        <p className="text-sm text-muted-foreground">{t("collections.none")}</p>
      )}
      {entries.map(([key, entryValue], index) => {
        const type = valueType(entryValue);
        return (
          <div
            key={`${key}-${index}`}
            className="grid gap-3 border-l-2 border-border pl-3 md:grid-cols-[12rem_8rem_minmax(0,1fr)_2.5rem]"
          >
            <FormField label={t("fields.configKey")} required>
              <Input
                value={key}
                onChange={(event) => {
                  replace(index, event.target.value, entryValue);
                }}
              />
            </FormField>
            <FormField label={t("fields.valueType")}>
              <Select
                value={type}
                disabled={type === "json"}
                onValueChange={(next) => {
                  replace(index, key, defaultValue(next));
                }}
              >
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="string">{t("valueTypes.string")}</SelectItem>
                  <SelectItem value="number">{t("valueTypes.number")}</SelectItem>
                  <SelectItem value="boolean">{t("valueTypes.boolean")}</SelectItem>
                  {type === "json" && (
                    <SelectItem value="json">{t("valueTypes.json")}</SelectItem>
                  )}
                </SelectContent>
              </Select>
            </FormField>
            <ConfigValueInput
              value={entryValue}
              type={type}
              onChange={(next) => replace(index, key, next)}
            />
            <Button
              type="button"
              variant="ghost"
              size="icon"
              className="self-start md:mt-7"
              title={t("collections.remove")}
              aria-label={`${t("collections.remove")} ${key}`}
              onClick={() => onChange(Object.fromEntries(
                entries.filter((_, item) => item !== index),
              ))}
            >
              <Trash2 className="h-4 w-4" />
            </Button>
          </div>
        );
      })}
    </div>
  );
}

function ConfigValueInput({
  value,
  type,
  onChange,
}: {
  value: unknown;
  type: "string" | "number" | "boolean" | "json";
  onChange: (value: unknown) => void;
}) {
  const t = useTranslations("resourceEditor");
  if (type === "boolean") {
    return (
      <FormField label={t("fields.configValue")}>
        <Select
          value={String(value)}
          onValueChange={(next) => onChange(next === "true")}
        >
          <SelectTrigger><SelectValue /></SelectTrigger>
          <SelectContent>
            <SelectItem value="true">{t("valueTypes.true")}</SelectItem>
            <SelectItem value="false">{t("valueTypes.false")}</SelectItem>
          </SelectContent>
        </Select>
      </FormField>
    );
  }
  return (
    <FormField
      label={t("fields.configValue")}
      hint={type === "json" ? t("valueTypes.yamlOnly") : undefined}
    >
      <Input
        type={type === "number" ? "number" : "text"}
        readOnly={type === "json"}
        value={type === "json" ? JSON.stringify(value) : String(value)}
        onChange={(event) => onChange(
          type === "number" ? Number(event.target.value) : event.target.value,
        )}
      />
    </FormField>
  );
}

function valueType(value: unknown): "string" | "number" | "boolean" | "json" {
  if (typeof value === "string") return "string";
  if (typeof value === "number") return "number";
  if (typeof value === "boolean") return "boolean";
  return "json";
}

function defaultValue(type: string): string | number | boolean {
  if (type === "number") return 0;
  if (type === "boolean") return false;
  return "";
}

function nextKey(value: Record<string, unknown>): string {
  let index = Object.keys(value).length + 1;
  while (`config-${index}` in value) index++;
  return `config-${index}`;
}
