"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { AlertMessage } from "@/components/ui/alert-message";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
} from "@/components/ui/select";
import { Spinner } from "@/components/ui/spinner";
import { Switch } from "@/components/ui/switch";
import type { EnvBundleSummary } from "@/lib/api";
import type { WorkerSpecDraft } from "@/lib/api/facade/podConnect";
import type { AsyncState } from "../hooks/workerCreateDraft";
import {
  workerTypeFieldLabel,
  workerTypeFieldOptionLabel,
  type WorkerTypeConfigField,
} from "./workerTypeConfigSchema";
import type { WorkerTypeConfigStepProps } from "./WorkerTypeConfigStep";

const DEFAULT_VALUE = "__default__";
const EMPTY_VALUE = "__empty__";

export function WorkerTypeConfigField(
  props: WorkerTypeConfigStepProps & { field: WorkerTypeConfigField },
) {
  const { draft, field, onPatch, credentialBundles, t } = props;
  const label = workerTypeFieldLabel(field.name, t);
  const value = draft.type_config_values[field.name];
  const setValue = (next: unknown) => {
    const values = { ...draft.type_config_values };
    if (next === undefined) delete values[field.name];
    else values[field.name] = next;
    onPatch({ type_config_values: values });
  };

  if (field.kind === "secret") {
    return (
      <SecretReferenceField
        field={field}
        label={label}
        draft={draft}
        bundles={credentialBundles}
        onPatch={onPatch}
        t={t}
      />
    );
  }
  if (field.kind === "boolean") {
    return (
      <div className="flex items-center justify-between gap-4">
        <FieldLabel field={field} label={label} />
        <Switch
          id={`worker-type-${field.name}`}
          checked={value === true}
          onCheckedChange={setValue}
        />
      </div>
    );
  }
  if (field.kind === "select") {
    const selectedValue =
      typeof value === "string" && value !== "" ? value : DEFAULT_VALUE;
    const selectedLabel = selectedValue === DEFAULT_VALUE
      ? t("workerCreate.typeConfig.useDefault")
      : workerTypeFieldOptionLabel(field.name, selectedValue, t);
    return (
      <LabeledField field={field} label={label}>
        <Select
          value={selectedValue}
          onValueChange={(next) => setValue(
            next === DEFAULT_VALUE ? undefined : next === EMPTY_VALUE ? "" : next,
          )}
        >
          <SelectTrigger id={`worker-type-${field.name}`}>
            <span>{selectedLabel}</span>
          </SelectTrigger>
          <SelectContent>
            <SelectItem value={DEFAULT_VALUE}>{t("workerCreate.typeConfig.useDefault")}</SelectItem>
            {field.options.filter((option) => option !== "").map((option) => {
              const optionValue = option || EMPTY_VALUE;
              return (
                <SelectItem key={optionValue} value={optionValue}>
                  {workerTypeFieldOptionLabel(field.name, option, t)}
                </SelectItem>
              );
            })}
          </SelectContent>
        </Select>
      </LabeledField>
    );
  }
  return (
    <LabeledField field={field} label={label}>
      <Input
        id={`worker-type-${field.name}`}
        type={field.kind === "number" ? "number" : "text"}
        placeholder={formatDefault(field.defaultValue, t)}
        value={typeof value === "string" || typeof value === "number" ? value : ""}
        onChange={(event) => {
          const raw = event.target.value;
          setValue(field.kind === "number" ? (raw ? Number(raw) : undefined) : raw);
        }}
      />
    </LabeledField>
  );
}

function SecretReferenceField(props: {
  field: WorkerTypeConfigField;
  label: string;
  draft: WorkerSpecDraft;
  bundles: AsyncState<EnvBundleSummary[]>;
  onPatch: (patch: Partial<WorkerSpecDraft>) => void;
  t: (key: string) => string;
}) {
  const params = useParams() ?? {};
  const orgSlug = String(params.org ?? "");
  if (props.bundles.status === "loading" || props.bundles.status === "idle") return <Spinner />;
  if (props.bundles.status === "error") return <AlertMessage type="error" message={props.bundles.error} />;
  const candidates = props.bundles.data.filter(
    (bundle) => bundle.configured_fields?.includes(props.field.name),
  );
  const selected = props.draft.secret_refs.find((reference) => reference.field === props.field.name);
  const selectedLabel = selected
    ? candidates.find((bundle) => bundle.id === selected.id)?.name
      ?? String(selected.id)
    : props.t("workerCreate.typeConfig.useDefault");
  return (
    <LabeledField field={props.field} label={props.label}>
      <Select
        value={selected ? String(selected.id) : DEFAULT_VALUE}
        onValueChange={(value) => {
          const refs = props.draft.secret_refs.filter((reference) => reference.field !== props.field.name);
          if (value !== DEFAULT_VALUE) {
            refs.push({ field: props.field.name, kind: "env-bundle", id: Number(value) });
          }
          props.onPatch({ secret_refs: refs });
        }}
      >
        <SelectTrigger id={`worker-type-${props.field.name}`}>
          <span>{selectedLabel}</span>
        </SelectTrigger>
        <SelectContent>
          <SelectItem value={DEFAULT_VALUE}>{props.t("workerCreate.typeConfig.useDefault")}</SelectItem>
          {candidates.map((bundle) => <SelectItem key={bundle.id} value={String(bundle.id)}>{bundle.name}</SelectItem>)}
        </SelectContent>
      </Select>
      <p className="mt-1 text-xs text-muted-foreground">{props.t("workerCreate.typeConfig.secretHint")}</p>
      {orgSlug && (
        <Link
          href={`/${orgSlug}/settings?tab=agents/${props.draft.worker_type_slug}`}
          className="mt-1 inline-block text-xs font-medium text-primary hover:underline"
        >
          {props.t("workerCreate.typeConfig.manageCredentials")}
        </Link>
      )}
    </LabeledField>
  );
}

function LabeledField(props: {
  field: WorkerTypeConfigField;
  label: string;
  children: React.ReactNode;
}) {
  return (
    <div>
      <FieldLabel field={props.field} label={props.label} />
      {props.children}
      {props.field.description && (
        <p className="mt-1 text-xs text-muted-foreground">{props.field.description}</p>
      )}
    </div>
  );
}

function FieldLabel(props: { field: WorkerTypeConfigField; label: string }) {
  return (
    <label
      htmlFor={`worker-type-${props.field.name}`}
      className="mb-2 block text-sm font-medium"
    >
      {props.label}
      {props.field.required && <span className="ml-1 text-destructive">*</span>}
    </label>
  );
}

function formatDefault(value: unknown, t: (key: string) => string): string | undefined {
  if (value === undefined || value === null) return undefined;
  return `${t("workerCreate.typeConfig.defaultPrefix")}${String(value)}`;
}
