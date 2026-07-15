"use client";

import { AlertMessage } from "@/components/ui/alert-message";
import { Spinner } from "@/components/ui/spinner";
import type { EffectiveResource } from "@/lib/api/facade/aiResource";
import type {
  WorkerSpecDraft,
  WorkerToolModelRequirement,
} from "@/lib/api/facade/podConnect";
import type { AsyncState } from "../hooks/workerCreateDraft";
import {
  compatibleToolModelResources,
  modelResourceLabel,
  toolModelRoleLabel,
} from "./workerModelResources";
import { WorkerRuntimeSelectField } from "./WorkerRuntimeSelectField";

interface WorkerToolModelFieldProps {
  requirement: WorkerToolModelRequirement;
  state: AsyncState<EffectiveResource[]>;
  draft: WorkerSpecDraft;
  onPatch: (patch: Partial<WorkerSpecDraft>) => void;
  t: (key: string) => string;
}

export function WorkerToolModelField(props: WorkerToolModelFieldProps) {
  const field = props.requirement.role;
  const testId = `worker-runtime-field-${field}`;
  if (props.state.status === "idle" || props.state.status === "loading") {
    return <div data-testid={testId}><Spinner className="my-4" /></div>;
  }
  if (props.state.status === "error") {
    return <div data-testid={testId}><AlertMessage type="error" message={props.state.error} /></div>;
  }
  const resources = compatibleToolModelResources(props.requirement, props.state.data);
  if (resources.length === 0) {
    return (
      <div data-testid={testId}>
        <AlertMessage type="error" message={props.t("ide.createPod.noModelResourcesAvailableHint")} />
      </div>
    );
  }
  return (
    <div data-testid={testId}>
      <WorkerRuntimeSelectField
        field={field}
        label={toolModelRoleLabel(field)}
        value={selectedValue(props.draft.tool_model_resource_ids[field])}
        options={resources.map((item) => ({
          value: String(item.resource?.id ?? 0),
          label: modelResourceLabel(item),
          selectable: item.selectable,
          blockingReason: item.blockingReason,
        }))}
        onChange={(value) => props.onPatch({
          tool_model_resource_ids: {
            ...props.draft.tool_model_resource_ids,
            [field]: Number(value),
          },
        })}
      />
    </div>
  );
}

function selectedValue(value: number | undefined): string {
  return value && value > 0 ? String(value) : "";
}
