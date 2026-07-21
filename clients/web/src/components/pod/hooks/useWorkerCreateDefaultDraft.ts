import { useEffect } from "react";
import type { Dispatch } from "react";
import type { EnvBundleSummary, RepositoryData } from "@/lib/api";
import type { EffectiveResource } from "@/lib/api/facade/aiResource";
import type { WorkerCreateOptions } from "@/lib/api/facade/podConnect";
import type { AsyncState, WorkerCreateDraftAction, WorkerCreateDraftState } from "./workerCreateDraft";
import {
  defaultConfigDocumentPatch,
  defaultModelPatch,
  defaultToolModelPatch,
  defaultWorkerDraftPatch,
} from "./workerCreateDefaults";

export function useWorkerCreateDefaultDraft(input: {
  state: WorkerCreateDraftState;
  dispatch: Dispatch<WorkerCreateDraftAction>;
  options: AsyncState<WorkerCreateOptions>;
  modelResources: AsyncState<EffectiveResource[]>;
  toolModelResources: AsyncState<EffectiveResource[]>;
  configBundles: AsyncState<EnvBundleSummary[]>;
  repositories: RepositoryData[];
  preferredWorkerType?: string;
}) {
  const { state, dispatch } = input;
  const workerType = input.options.status === "ready"
    ? input.options.data.worker_types.find(
      (option) => option.slug === state.draft.worker_type_slug,
    )
    : undefined;

  useEffect(() => {
    if (input.options.status !== "ready") return;
    const patch = defaultWorkerDraftPatch(
      state.draft,
      input.options.data,
      input.preferredWorkerType,
    );
    if (Object.keys(patch).length > 0) dispatch({ type: "patch_draft", patch });
  }, [dispatch, input.options, input.preferredWorkerType, state.draft]);

  useEffect(() => {
    if (input.modelResources.status !== "ready") return;
    const patch = defaultModelPatch(state.draft, input.modelResources.data);
    if (Object.keys(patch).length > 0) dispatch({ type: "patch_draft", patch });
  }, [dispatch, input.modelResources, state.draft]);

  useEffect(() => {
    if (input.toolModelResources.status !== "ready") return;
    const patch = defaultToolModelPatch(
      state.draft,
      workerType,
      input.toolModelResources.data,
    );
    if (Object.keys(patch).length > 0) dispatch({ type: "patch_draft", patch });
  }, [dispatch, input.toolModelResources, state.draft, workerType]);

  useEffect(() => {
    if (input.configBundles.status !== "ready") return;
    const patch = defaultConfigDocumentPatch(
      state.draft,
      workerType,
      input.configBundles.data,
    );
    if (Object.keys(patch).length > 0) dispatch({ type: "patch_draft", patch });
  }, [dispatch, input.configBundles, state.draft, workerType]);

  useEffect(() => {
    const repositoryId = state.draft.repository_id;
    if (!repositoryId || state.draft.branch) return;
    const repository = input.repositories.find((item) => item.id === repositoryId);
    if (repository) {
      dispatch({ type: "patch_draft", patch: { branch: repository.default_branch } });
    }
  }, [dispatch, input.repositories, state.draft.branch, state.draft.repository_id]);
}
