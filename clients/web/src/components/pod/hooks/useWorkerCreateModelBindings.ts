import { useEffect, useMemo } from "react";
import type { Dispatch } from "react";
import type {
  EffectiveResource,
  ProviderDefinition,
} from "@/lib/api/facade/aiResource";
import type {
  WorkerCreateOptions,
  WorkerSpecDraft,
} from "@/lib/api/facade/podConnect";
import {
  compatibleWorkerModelResources,
  generationModelRequirement,
} from "../CreatePodForm/workerModelResourceCompatibility";
import { defaultWorkerModelBindingsPatch } from "./workerCreateDefaults";
import type {
  AsyncState,
  WorkerCreateDraftAction,
} from "./workerCreateDraft";

interface WorkerCreateModelBindingInput {
  draft: WorkerSpecDraft;
  generationModelResourceId: number;
  options: AsyncState<WorkerCreateOptions>;
  modelResources: AsyncState<EffectiveResource[]>;
  modelProviders: AsyncState<ProviderDefinition[]>;
  dispatch: Dispatch<WorkerCreateDraftAction>;
}

export function useWorkerCreateModelBindings(
  input: WorkerCreateModelBindingInput,
) {
  const {
    draft,
    generationModelResourceId,
    options,
    modelResources,
    modelProviders,
    dispatch,
  } = input;
  const selectedWorkerType = options.status === "ready"
    ? options.data.worker_types.find(
      (option) => option.slug === draft.worker_type_slug,
    )
    : undefined;
  const generationModelResources = useMemo(
    () => compatibleState(
      modelResources,
      modelProviders,
      generationModelRequirement,
    ),
    [modelProviders, modelResources],
  );

  useEffect(() => {
    if (
      !selectedWorkerType ||
      modelResources.status !== "ready" ||
      modelProviders.status !== "ready"
    ) {
      return;
    }
    const patch = defaultWorkerModelBindingsPatch(
      draft,
      selectedWorkerType,
      modelResources.data,
      modelProviders.data,
    );
    if (Object.keys(patch).length > 0) {
      dispatch({ type: "patch_draft", patch });
    }
  }, [
    dispatch,
    draft,
    modelProviders,
    modelResources,
    selectedWorkerType,
  ]);

  useEffect(() => {
    if (generationModelResources.status !== "ready") return;
    const currentExists = generationModelResources.data.some(
      (item) => item.resource?.id === generationModelResourceId,
    );
    const nextId = currentExists
      ? generationModelResourceId
      : generationModelResources.data[0]?.resource?.id ?? 0;
    if (nextId !== generationModelResourceId) {
      dispatch({ type: "set_generation_model", resourceId: nextId });
    }
  }, [
    generationModelResources,
    dispatch,
    generationModelResourceId,
  ]);

  const runtimeModelsRequired = Boolean(
    selectedWorkerType?.requires_model_resource ||
      selectedWorkerType?.tool_model_requirements.length,
  );
  return {
    generationModelResources,
    modelDependenciesReady: !runtimeModelsRequired || (
      modelResources.status === "ready" &&
      modelProviders.status === "ready"
    ),
  };
}

function compatibleState(
  resources: AsyncState<EffectiveResource[]>,
  providers: AsyncState<ProviderDefinition[]>,
  requirement: typeof generationModelRequirement,
): AsyncState<EffectiveResource[]> {
  if (resources.status === "error") return resources;
  if (providers.status === "error") return providers;
  if (resources.status === "loading" || providers.status === "loading") {
    return { status: "loading" };
  }
  if (resources.status !== "ready" || providers.status !== "ready") {
    return { status: "idle" };
  }
  return {
    status: "ready",
    data: compatibleWorkerModelResources(
      resources.data,
      providers.data,
      requirement,
    ),
  };
}
