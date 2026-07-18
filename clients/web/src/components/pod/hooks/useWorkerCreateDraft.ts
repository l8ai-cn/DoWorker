import { useCallback, useEffect, useReducer } from "react";
import type { PodData, RepositoryData } from "@/lib/api";
import { podApi } from "@/lib/api";
import { readCurrentOrg } from "@/stores/auth";
import {
  createInitialWorkerDraftState,
  type WorkerCreateStepId,
  workerCreateDraftReducer,
} from "./workerCreateDraft";
import {
  assembleWorkerCreateController,
  type WorkerCreateController,
  workerCreateError,
  workerCreateInitialDraft,
} from "./workerCreateController";
import { useWorkerCreateDependencies } from "./useWorkerCreateDependencies";
import { defaultWorkerDraftPatch } from "./workerCreateDefaults";
import { useWorkerCreateModelBindings } from "./useWorkerCreateModelBindings";
import { useWorkerCreateOptions } from "./useWorkerCreateOptions";
import { useWorkerCreateSubmission } from "./useWorkerCreateSubmission";
import { workerCreateValidity } from "./workerCreateValidity";
import {
  clearWorkerCreateDraft,
  loadWorkerCreateDraft,
  persistWorkerCreateDraft,
} from "./workerCreateDraftPersistence";

export type { WorkerCreateController } from "./workerCreateController";
interface UseWorkerCreateDraftParams {
  enabled: boolean;
  repositories: RepositoryData[];
  initialWorkerTypeSlug?: string;
  initialTask?: string;
  initialRepositoryId?: number | null;
  ticketSlug?: string;
  onSuccess?: (pod: PodData) => void;
  onError?: (error: Error) => void;
}

export function useWorkerCreateDraft(
  params: UseWorkerCreateDraftParams,
): WorkerCreateController {
  const orgSlug = readCurrentOrg()?.slug ?? "";
  const [state, dispatch] = useReducer(
    workerCreateDraftReducer,
    undefined,
    () => {
      const initial = workerCreateInitialDraft(params);
      const persisted = loadWorkerCreateDraft(orgSlug);
      if (!persisted) return createInitialWorkerDraftState(initial);
      return {
        ...createInitialWorkerDraftState({ ...persisted.draft, ...initial }),
        step: persisted.step,
        fillPrompt: persisted.fillPrompt,
      };
    },
  );
  const options = useWorkerCreateOptions(params.enabled, orgSlug, {
    workerTypeSlug: state.draft.worker_type_slug,
    computeTargetId: state.draft.compute_target_id,
    deploymentMode: state.draft.deployment_mode,
  });
  const selectedWorkerType = options.status === "ready"
    ? options.data.worker_types.find(
      (option) => option.slug === state.draft.worker_type_slug,
    )
    : undefined;
  const dependencies = useWorkerCreateDependencies(
    selectedWorkerType,
    state.draft.repository_id,
  );
  const modelBindings = useWorkerCreateModelBindings({
    draft: state.draft,
    generationModelResourceId: state.generationModelResourceId,
    options,
    modelResources: dependencies.modelResources,
    modelProviders: dependencies.modelProviders,
    dispatch,
  });
  const validity = workerCreateValidity(
    state.draft,
    options,
    dependencies.modelResources.status === "ready" &&
      dependencies.toolModelResources.status === "ready" &&
      dependencies.runtimeBundles.status === "ready" &&
      dependencies.credentialBundles.status === "ready" &&
      dependencies.configBundles.status === "ready" &&
      dependencies.skills.status === "ready",
    modelBindings.modelDependenciesReady,
  );

  useEffect(() => {
    if (state.create.status === "ready") {
      clearWorkerCreateDraft(orgSlug);
      return;
    }
    persistWorkerCreateDraft(orgSlug, {
      step: state.step,
      fillPrompt: state.fillPrompt,
      draft: state.draft,
    });
  }, [orgSlug, state.create.status, state.draft, state.fillPrompt, state.step]);

  useEffect(() => {
    if (options.status !== "ready") return;
    const patch = defaultWorkerDraftPatch(
      state.draft,
      options.data,
      params.initialWorkerTypeSlug,
    );
    if (Object.keys(patch).length > 0) {
      dispatch({ type: "patch_draft", patch });
    }
  }, [options, params.initialWorkerTypeSlug, state.draft]);

  useEffect(() => {
    const repositoryId = state.draft.repository_id;
    if (!repositoryId || state.draft.branch) return;
    const repository = params.repositories.find((item) => item.id === repositoryId);
    if (repository) {
      dispatch({ type: "patch_draft", patch: { branch: repository.default_branch } });
    }
  }, [params.repositories, state.draft.branch, state.draft.repository_id]);

  useEffect(() => {
    if (!params.enabled) {
      dispatch({ type: "reset", draft: workerCreateInitialDraft(params) });
    }
  }, [params.enabled]); // eslint-disable-line react-hooks/exhaustive-deps

  const runPreflight = useCallback(async () => {
    const requestId = crypto.randomUUID();
    dispatch({ type: "preflight_loading", requestId });
    try {
      const result = await podApi.preflightWorker(state.draft);
      dispatch({ type: "preflight_succeeded", requestId, result });
      return result;
    } catch (error) {
      const resolved = workerCreateError(error);
      dispatch({
        type: "preflight_failed",
        requestId,
        error: resolved.message,
      });
      params.onError?.(resolved);
      return null;
    }
  }, [params, state.draft]);

  const goToStep = useCallback(async (step: WorkerCreateStepId) => {
    if (!validity.accessible(step)) return;
    dispatch({ type: "set_step", step });
    if (step === 4) await runPreflight();
  }, [runPreflight, validity]);

  const fillWithAI = useCallback(async (prompt: string) => {
    const requestId = crypto.randomUUID();
    dispatch({ type: "fill_loading", requestId });
    try {
      const result = await podApi.fillWorkerDraft(
        prompt,
        state.generationModelResourceId,
        state.draft,
      );
      dispatch({ type: "fill_succeeded", requestId, result });
    } catch (error) {
      const resolved = workerCreateError(error);
      dispatch({ type: "fill_failed", requestId, error: resolved.message });
      params.onError?.(resolved);
    }
  }, [params, state.draft, state.generationModelResourceId]);

  const createWorker = useWorkerCreateSubmission({
    dispatch,
    state,
    ticketSlug: params.ticketSlug,
    onSuccess: params.onSuccess,
    onError: params.onError,
  });

  return assembleWorkerCreateController({
    state,
    options,
    modelResources: dependencies.modelResources,
    toolModelResources: dependencies.toolModelResources,
    runtimeBundles: dependencies.runtimeBundles,
    credentialBundles: dependencies.credentialBundles,
    configBundles: dependencies.configBundles,
    skills: dependencies.skills,
    repositories: params.repositories,
    validity,
    dispatch,
    initial: workerCreateInitialDraft(params),
    fillWithAI,
    goToStep,
    runPreflight,
    createWorker,
  });
}
