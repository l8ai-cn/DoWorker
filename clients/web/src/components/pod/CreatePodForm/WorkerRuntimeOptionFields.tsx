"use client";

import type {
  WorkerCreateOptions,
  WorkerResourceRequest,
  WorkerSpecDraft,
} from "@/lib/api/facade/podConnect";
import { WorkerResourceEditor } from "./WorkerResourceEditor";
import {
  WorkerRuntimeSelectField,
  type WorkerRuntimeSelectOption,
} from "./WorkerRuntimeSelectField";
import { localizeWorkerRuntimeOption } from "./workerRuntimeOptionLabels";

interface WorkerRuntimeOptionFieldsProps {
  draft: WorkerSpecDraft;
  data: WorkerCreateOptions;
  onPatch: (patch: Partial<WorkerSpecDraft>) => void;
  onWorkerTypeChange: (slug: string) => void;
  t: (key: string) => string;
}

export function WorkerRuntimeOptionFields({
  draft,
  data,
  onPatch,
  onWorkerTypeChange,
  t,
}: WorkerRuntimeOptionFieldsProps) {
  const selectedProfile = data.resource_profiles.find(
    (option) => option.id === draft.resource_profile_id,
  );
  const customResources = draft.custom_resources
    ?? (selectedProfile ? resourceRequestFromProfile(selectedProfile) : undefined);

  return (
    <>
      <WorkerRuntimeSelectField
        field="worker-type"
        label={t("workerCreate.runtime.workerType")}
        value={draft.worker_type_slug}
        options={data.worker_types.map((option) => selectOption(
          "workerType",
          option.slug,
          option.name,
          option.selectable,
          option.blocking_reason,
          t,
        ))}
        onChange={onWorkerTypeChange}
      />
      <WorkerRuntimeSelectField
        field="runtime-image"
        label={t("workerCreate.runtime.runtimeImage")}
        value={numberValue(draft.runtime_image_id)}
        options={data.runtime_images.map((option) => selectOption(
          "runtimeImage",
          String(option.id),
          option.name,
          option.selectable,
          option.blocking_reason,
          t,
        ))}
        onChange={(value) => onPatch({ runtime_image_id: Number(value) })}
      />
      <WorkerRuntimeSelectField
        field="compute-target"
        label={t("workerCreate.runtime.computeTarget")}
        description={t("workerCreate.runtime.computeTargetHint")}
        value={numberValue(draft.compute_target_id)}
        options={data.compute_targets.map((option) => selectOption(
          "computeTarget",
          String(option.id),
          option.name,
          option.selectable,
          option.blocking_reason,
          t,
          option.slug,
        ))}
        onChange={(value) => onPatch({ compute_target_id: Number(value) })}
      />
      <WorkerRuntimeSelectField
        field="deployment-mode"
        label={t("workerCreate.runtime.deploymentMode")}
        description={t("workerCreate.runtime.deploymentModeHint")}
        value={draft.deployment_mode}
        options={data.deployment_modes.map((option) => selectOption(
          "deploymentMode",
          option.value,
          option.name,
          option.selectable,
          option.blocking_reason,
          t,
        ))}
        onChange={(value) => onPatch({ deployment_mode: value })}
      />
      <WorkerRuntimeSelectField
        field="resource-profile"
        label={t("workerCreate.runtime.resourceProfile")}
        description={t("workerCreate.runtime.resourceProfileHint")}
        value={draft.custom_resources ? "__custom__" : numberValue(draft.resource_profile_id)}
        options={[
          ...data.resource_profiles.map((option) => selectOption(
            "resourceProfile",
            String(option.id),
            option.name,
            option.selectable,
            option.blocking_reason,
            t,
            option.slug,
          )),
          {
            value: "__custom__",
            label: t("workerCreate.runtime.resources.custom"),
            selectable: true,
            blockingReason: "",
          },
        ]}
        onChange={(value) => {
          if (value === "__custom__") {
            onPatch({
              resource_profile_id: 0,
              custom_resources: customResources ?? emptyResourceRequest(),
            });
            return;
          }
          onPatch({
            resource_profile_id: Number(value),
            custom_resources: undefined,
          });
        }}
      />
      {draft.custom_resources && (
        <WorkerResourceEditor
          value={draft.custom_resources}
          onChange={(custom) => onPatch({ custom_resources: custom })}
          t={t}
        />
      )}
    </>
  );
}

function selectOption(
  kind: Parameters<typeof localizeWorkerRuntimeOption>[0],
  value: string,
  label: string,
  selectable: boolean,
  blockingReason: string,
  t: (key: string) => string,
  lookupValue?: string,
): WorkerRuntimeSelectOption {
  return localizeWorkerRuntimeOption(
    kind,
    value,
    label,
    selectable,
    blockingReason,
    t,
    lookupValue,
  );
}

function numberValue(value: number): string {
  return value > 0 ? String(value) : "";
}

function resourceRequestFromProfile(
  profile: WorkerCreateOptions["resource_profiles"][number],
): WorkerResourceRequest {
  return {
    cpu_request_millicpu: profile.cpu_request_millicpu,
    cpu_limit_millicpu: profile.cpu_limit_millicpu,
    memory_request_bytes: profile.memory_request_bytes,
    memory_limit_bytes: profile.memory_limit_bytes,
    storage_request_bytes: profile.storage_request_bytes,
    storage_limit_bytes: profile.storage_limit_bytes,
  };
}

function emptyResourceRequest(): WorkerResourceRequest {
  return {
    cpu_request_millicpu: 500,
    cpu_limit_millicpu: 500,
    memory_request_bytes: 1 << 30,
    memory_limit_bytes: 1 << 30,
    storage_request_bytes: 10 << 30,
    storage_limit_bytes: 10 << 30,
  };
}
