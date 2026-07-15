"use client";

import { useState } from "react";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { AlertMessage } from "@/components/ui/alert-message";
import { Spinner } from "@/components/ui/spinner";
import type {
  EffectiveResource,
  ProviderDefinition,
} from "@/lib/api/facade/aiResource";
import type { WorkerCreateOptions, WorkerSpecDraft } from "@/lib/api/facade/podConnect";
import type { AsyncState } from "../hooks/workerCreateDraft";
import { WorkerRuntimeModelFields } from "./WorkerRuntimeModelFields";
import {
  WorkerRuntimeSelectField,
  type WorkerRuntimeSelectOption,
} from "./WorkerRuntimeSelectField";

interface WorkerRuntimeStepProps {
  draft: WorkerSpecDraft;
  options: AsyncState<WorkerCreateOptions>;
  modelResources: AsyncState<EffectiveResource[]>;
  modelProviders: AsyncState<ProviderDefinition[]>;
  onPatch: (patch: Partial<WorkerSpecDraft>) => void;
  onWorkerTypeChange: (slug: string, schemaVersion: number) => void;
  t: (key: string) => string;
}

export function WorkerRuntimeStep(props: WorkerRuntimeStepProps) {
  const {
    draft,
    options,
    modelResources,
    modelProviders,
    onPatch,
    onWorkerTypeChange,
    t,
  } = props;
  const [pendingType, setPendingType] = useState<string | null>(null);

  if (options.status === "idle" || options.status === "loading") {
    return <Spinner className="my-8" />;
  }
  if (options.status === "error") {
    return <AlertMessage type="error" message={options.error} />;
  }
  const data = options.data;
  const workerTypes = data.worker_types.map((option) => selectOption(
    option.slug,
    option.name,
    option.selectable,
    option.blocking_reason,
  ));
  const selectedPending = data.worker_types.find((option) => option.slug === pendingType);
  const selectedWorkerType = data.worker_types.find(
    (option) => option.slug === draft.worker_type_slug,
  );

  const changeWorkerType = (slug: string) => {
    if (slug === draft.worker_type_slug) return;
    if (hasTypeSpecificValues(draft)) {
      setPendingType(slug);
      return;
    }
    const selected = data.worker_types.find((option) => option.slug === slug);
    if (selected) onWorkerTypeChange(selected.slug, selected.schema_version);
  };

  return (
    <div className="space-y-5">
      {selectedWorkerType && (
        <WorkerRuntimeModelFields
          draft={draft}
          workerType={selectedWorkerType}
          modelResources={modelResources}
          modelProviders={modelProviders}
          onPatch={onPatch}
          t={t}
        />
      )}
      <WorkerRuntimeSelectField
        field="worker-type"
        label={t("workerCreate.runtime.workerType")}
        value={draft.worker_type_slug}
        options={workerTypes}
        onChange={changeWorkerType}
      />
      <WorkerRuntimeSelectField
        field="runtime-image"
        label={t("workerCreate.runtime.runtimeImage")}
        value={numberValue(draft.runtime_image_id)}
        options={data.runtime_images.map((option) => selectOption(
          String(option.id), option.name, option.selectable, option.blocking_reason,
        ))}
        onChange={(value) => onPatch({ runtime_image_id: Number(value) })}
      />
      <WorkerRuntimeSelectField
        field="compute-target"
        label={t("workerCreate.runtime.computeTarget")}
        value={numberValue(draft.compute_target_id)}
        options={data.compute_targets.map((option) => selectOption(
          String(option.id), option.name, option.selectable, option.blocking_reason,
        ))}
        onChange={(value) => onPatch({ compute_target_id: Number(value) })}
      />
      <WorkerRuntimeSelectField
        field="deployment-mode"
        label={t("workerCreate.runtime.deploymentMode")}
        value={draft.deployment_mode}
        options={data.deployment_modes.map((option) => selectOption(
          option.value, option.name, option.selectable, option.blocking_reason,
        ))}
        onChange={(value) => onPatch({ deployment_mode: value })}
      />
      <WorkerRuntimeSelectField
        field="resource-profile"
        label={t("workerCreate.runtime.resourceProfile")}
        value={numberValue(draft.resource_profile_id)}
        options={data.resource_profiles.map((option) => selectOption(
          String(option.id), option.name, option.selectable, option.blocking_reason,
        ))}
        onChange={(value) => onPatch({ resource_profile_id: Number(value) })}
      />

      <AlertDialog open={pendingType !== null} onOpenChange={(open) => !open && setPendingType(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{t("workerCreate.typeChange.title")}</AlertDialogTitle>
            <AlertDialogDescription>
              {t("workerCreate.typeChange.description")}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>{t("workerCreate.typeChange.cancel")}</AlertDialogCancel>
            <AlertDialogAction
              onClick={() => {
                if (selectedPending) {
                  onWorkerTypeChange(selectedPending.slug, selectedPending.schema_version);
                }
                setPendingType(null);
              }}
            >
              {t("workerCreate.typeChange.confirm")}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}

function selectOption(
  value: string,
  label: string,
  selectable: boolean,
  blockingReason: string,
): WorkerRuntimeSelectOption {
  return { value, label, selectable, blockingReason };
}

function numberValue(value: number): string {
  return value > 0 ? String(value) : "";
}

function hasTypeSpecificValues(draft: WorkerSpecDraft): boolean {
  return Boolean(
    draft.worker_type_slug ||
      draft.model_resource_id ||
      Object.keys(draft.tool_model_resource_ids).length ||
      draft.runtime_image_id ||
      Object.keys(draft.type_config_values).length ||
      draft.secret_refs.length ||
      draft.skill_ids.length ||
      draft.env_bundle_ids.length,
  );
}
