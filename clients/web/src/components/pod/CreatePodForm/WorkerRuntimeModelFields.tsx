"use client";

import { AlertMessage } from "@/components/ui/alert-message";
import { Spinner } from "@/components/ui/spinner";
import type {
  EffectiveResource,
  ProviderDefinition,
} from "@/lib/api/facade/aiResource";
import type {
  WorkerSpecDraft,
  WorkerTypeOption,
} from "@/lib/api/facade/podConnect";
import type { AsyncState } from "../hooks/workerCreateDraft";
import {
  compatibleWorkerModelResources,
  primaryModelRequirement,
  toolModelRequirement,
} from "./workerModelResourceCompatibility";
import { modelResourceLabel } from "./workerModelResources";
import {
  WorkerRuntimeSelectField,
  type WorkerRuntimeSelectOption,
} from "./WorkerRuntimeSelectField";

interface WorkerRuntimeModelFieldsProps {
  draft: WorkerSpecDraft;
  workerType: WorkerTypeOption;
  modelResources: AsyncState<EffectiveResource[]>;
  modelProviders: AsyncState<ProviderDefinition[]>;
  onPatch: (patch: Partial<WorkerSpecDraft>) => void;
  t: (key: string) => string;
}

export function WorkerRuntimeModelFields(
  props: WorkerRuntimeModelFieldsProps,
) {
  const needsModels = props.workerType.requires_model_resource ||
    props.workerType.tool_model_requirements.length > 0;
  if (!needsModels) return null;
  if (
    props.modelResources.status === "idle" ||
    props.modelResources.status === "loading" ||
    props.modelProviders.status === "idle" ||
    props.modelProviders.status === "loading"
  ) {
    return <div data-runtime-field="models"><Spinner className="my-4" /></div>;
  }
  if (props.modelResources.status === "error") {
    return <AlertMessage type="error" message={props.modelResources.error} />;
  }
  if (props.modelProviders.status === "error") {
    return <AlertMessage type="error" message={props.modelProviders.error} />;
  }

  const resources = props.modelResources.data;
  const providers = props.modelProviders.data;
  const primaryRequirement = primaryModelRequirement(props.workerType);
  const primaryResources = primaryRequirement
    ? compatibleWorkerModelResources(resources, providers, primaryRequirement)
    : [];

  return (
    <>
      {primaryRequirement && (
        <ModelRequirementField
          field="model"
          label={props.t("workerCreate.runtime.model")}
          resources={primaryResources}
          value={props.draft.model_resource_id}
          emptyMessage={props.t("ide.createPod.noModelResourcesAvailableHint")}
          onChange={(id) => props.onPatch({ model_resource_id: id })}
        />
      )}
      {props.workerType.tool_model_requirements.map((requirement) => (
        <ModelRequirementField
          key={requirement.role}
          field={`tool-model-${requirement.role}`}
          label={`${props.t("workerCreate.runtime.toolModel")} · ${requirement.role}`}
          resources={compatibleWorkerModelResources(
            resources,
            providers,
            toolModelRequirement(requirement),
          )}
          value={props.draft.tool_model_resource_ids[requirement.role] ?? 0}
          emptyMessage={props.t("workerCreate.runtime.noCompatibleToolModel")}
          onChange={(id) => props.onPatch({
            tool_model_resource_ids: {
              ...props.draft.tool_model_resource_ids,
              [requirement.role]: id,
            },
          })}
        />
      ))}
    </>
  );
}

function ModelRequirementField(props: {
  field: string;
  label: string;
  resources: EffectiveResource[];
  value: number;
  emptyMessage: string;
  onChange: (id: number) => void;
}) {
  if (props.resources.length === 0) {
    return (
      <div data-runtime-field={props.field}>
        <p className="mb-2 text-sm font-medium">{props.label}</p>
        <AlertMessage type="error" message={props.emptyMessage} />
      </div>
    );
  }
  return (
    <WorkerRuntimeSelectField
      field={props.field}
      label={props.label}
      value={props.value > 0 ? String(props.value) : ""}
      options={props.resources.map(modelSelectOption)}
      onChange={(value) => props.onChange(Number(value))}
    />
  );
}

function modelSelectOption(resource: EffectiveResource): WorkerRuntimeSelectOption {
  return {
    value: String(resource.resource?.id ?? 0),
    label: modelResourceLabel(resource),
    selectable: true,
    blockingReason: "",
  };
}
