"use client";

import { AlertMessage } from "@/components/ui/alert-message";
import type { EnvBundleSummary } from "@/lib/api";
import type {
  WorkerCreateOptions,
  WorkerSpecDraft,
} from "@/lib/api/facade/podConnect";
import type { AsyncState } from "../hooks/workerCreateDraft";
import {
  parseWorkerTypeConfigSchema,
} from "./workerTypeConfigSchema";
import { WorkerTypeConfigField } from "./WorkerTypeConfigField";

export interface WorkerTypeConfigStepProps {
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
        <div>
          <h3 className="text-base font-semibold">{props.t("workerCreate.typeConfig.title")}</h3>
          <p className="mt-1 text-sm text-muted-foreground">
            {props.t("workerCreate.typeConfig.description")}
          </p>
        </div>
        {schema.fields.map((field) => (
          <WorkerTypeConfigField key={field.name} field={field} {...props} />
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
