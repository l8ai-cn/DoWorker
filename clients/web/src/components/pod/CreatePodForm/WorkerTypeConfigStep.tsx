"use client";

import { AlertMessage } from "@/components/ui/alert-message";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Spinner } from "@/components/ui/spinner";
import { Switch } from "@/components/ui/switch";
import type { EnvBundleSummary } from "@/lib/api";
import type {
  WorkerCreateOptions,
  WorkerSpecDraft,
} from "@/lib/api/facade/podConnect";
import type { AsyncState } from "../hooks/workerCreateDraft";
import {
  parseWorkerTypeConfigSchema,
  workerTypeFieldLabel,
  type WorkerTypeConfigField,
} from "./workerTypeConfigSchema";

interface WorkerTypeConfigStepProps {
  draft: WorkerSpecDraft;
  options: AsyncState<WorkerCreateOptions>;
  credentialBundles: AsyncState<EnvBundleSummary[]>;
  onPatch: (patch: Partial<WorkerSpecDraft>) => void;
  t: (key: string) => string;
}

export function WorkerTypeConfigStep(props: WorkerTypeConfigStepProps) {
  if (props.options.status !== "ready") {
    return <AlertMessage type="error" message={props.t("workerCreate.typeConfig.optionsRequired")} />;
  }
  const selected = props.options.data.worker_types.find(
    (option) => option.slug === props.draft.worker_type_slug,
  );
  if (!selected) {
    return <AlertMessage type="error" message={props.t("workerCreate.typeConfig.typeRequired")} />;
  }

  try {
    const schema = parseWorkerTypeConfigSchema(selected.config_schema);
    if (schema.version !== selected.schema_version) {
      throw new Error(props.t("workerCreate.typeConfig.versionMismatch"));
    }
    if (schema.fields.length === 0) {
      return <p className="text-sm text-muted-foreground">{props.t("workerCreate.typeConfig.empty")}</p>;
    }
    return (
      <div className="space-y-5">
        {schema.fields.map((field) => (
          <TypeConfigField key={field.name} field={field} {...props} />
        ))}
      </div>
    );
  } catch (error) {
    return (
      <AlertMessage
        type="error"
        message={error instanceof Error ? error.message : props.t("workerCreate.typeConfig.invalid")}
      />
    );
  }
}

function TypeConfigField(
  props: WorkerTypeConfigStepProps & { field: WorkerTypeConfigField },
) {
  const { draft, field, onPatch, credentialBundles, t } = props;
  const label = workerTypeFieldLabel(field.name);
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
        <label htmlFor={`worker-type-${field.name}`} className="text-sm font-medium">{label}</label>
        <Switch
          id={`worker-type-${field.name}`}
          checked={value === true}
          onCheckedChange={setValue}
        />
      </div>
    );
  }
  if (field.kind === "select") {
    return (
      <LabeledField label={label} name={field.name}>
        <Select
          value={typeof value === "string" ? value : "__default__"}
          onValueChange={(next) => setValue(next === "__default__" ? undefined : next)}
        >
          <SelectTrigger id={`worker-type-${field.name}`}>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="__default__">{t("workerCreate.typeConfig.useDefault")}</SelectItem>
            {field.options.map((option) => <SelectItem key={option} value={option}>{option}</SelectItem>)}
          </SelectContent>
        </Select>
      </LabeledField>
    );
  }
  return (
    <LabeledField label={label} name={field.name}>
      <Input
        id={`worker-type-${field.name}`}
        type={field.kind === "number" ? "number" : "text"}
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
  if (props.bundles.status === "loading" || props.bundles.status === "idle") return <Spinner />;
  if (props.bundles.status === "error") return <AlertMessage type="error" message={props.bundles.error} />;
  const candidates = props.bundles.data.filter(
    (bundle) => bundle.configured_fields?.includes(props.field.name),
  );
  const selected = props.draft.secret_refs.find((reference) => reference.field === props.field.name);
  return (
    <LabeledField label={props.label} name={props.field.name}>
      <Select
        value={selected ? String(selected.id) : "__default__"}
        onValueChange={(value) => {
          const refs = props.draft.secret_refs.filter((reference) => reference.field !== props.field.name);
          if (value !== "__default__") {
            refs.push({ field: props.field.name, kind: "env-bundle", id: Number(value) });
          }
          props.onPatch({ secret_refs: refs });
        }}
      >
        <SelectTrigger id={`worker-type-${props.field.name}`}><SelectValue /></SelectTrigger>
        <SelectContent>
          <SelectItem value="__default__">{props.t("workerCreate.typeConfig.useDefault")}</SelectItem>
          {candidates.map((bundle) => <SelectItem key={bundle.id} value={String(bundle.id)}>{bundle.name}</SelectItem>)}
        </SelectContent>
      </Select>
      <p className="mt-1 text-xs text-muted-foreground">{props.t("workerCreate.typeConfig.secretHint")}</p>
    </LabeledField>
  );
}

function LabeledField(props: { label: string; name: string; children: React.ReactNode }) {
  return (
    <div>
      <label htmlFor={`worker-type-${props.name}`} className="mb-2 block text-sm font-medium">{props.label}</label>
      {props.children}
    </div>
  );
}
