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
import type { EffectiveResource } from "@/lib/api/facade/aiResource";
import type {
  WorkerCreateOptions,
  WorkerSpecDraft,
} from "@/lib/api/facade/podConnect";
import type { AsyncState } from "../hooks/workerCreateDraft";
import { WorkerPrimaryModelField } from "./WorkerPrimaryModelField";
import { WorkerToolModelField } from "./WorkerToolModelField";
import { WorkerRuntimeOptionFields } from "./WorkerRuntimeOptionFields";

interface WorkerRuntimeStepProps {
  draft: WorkerSpecDraft;
  options: AsyncState<WorkerCreateOptions>;
  modelResources: AsyncState<EffectiveResource[]>;
  toolModelResources: AsyncState<EffectiveResource[]>;
  onPatch: (patch: Partial<WorkerSpecDraft>) => void;
  onWorkerTypeChange: (slug: string, schemaVersion: number) => void;
  t: (key: string) => string;
}

export function WorkerRuntimeStep(props: WorkerRuntimeStepProps) {
  const {
    draft,
    options,
    modelResources,
    toolModelResources,
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
      {selectedWorkerType?.requires_model_resource && (
        <WorkerPrimaryModelField
          state={modelResources}
          draft={draft}
          onPatch={onPatch}
          t={t}
        />
      )}
      {selectedWorkerType?.tool_model_requirements.map((requirement) => (
        <WorkerToolModelField
          key={requirement.role}
          requirement={requirement}
          state={toolModelResources}
          draft={draft}
          onPatch={onPatch}
          t={t}
        />
      ))}
      <WorkerRuntimeOptionFields
        draft={draft}
        data={data}
        onPatch={onPatch}
        onWorkerTypeChange={changeWorkerType}
        t={t}
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

function hasTypeSpecificValues(draft: WorkerSpecDraft): boolean {
  return Boolean(
    draft.worker_type_slug ||
      draft.model_resource_id ||
      Object.keys(draft.tool_model_resource_ids).length ||
      draft.runtime_image_id ||
      Object.keys(draft.type_config_values).length ||
      draft.secret_refs.length ||
      draft.skill_ids.length ||
      draft.env_bundle_ids.length ||
      draft.config_document_bindings.length ||
      draft.custom_resources,
  );
}
