import type { Dispatch } from "react";
import type {
  EnvBundleSummary,
  InstalledSkill,
  PodData,
  RepositoryData,
} from "@/lib/api";
import type { EffectiveResource } from "@/lib/api/facade/aiResource";
import type {
  WorkerCreateOptions,
  WorkerPreflightResult,
  WorkerSpecDraft,
} from "@/lib/api/facade/podConnect";
import type {
  AsyncState,
  WorkerCreateDraftAction,
  WorkerCreateDraftState,
  WorkerCreateStepId,
} from "./workerCreateDraft";
import type { WorkerCreateValidity } from "./workerCreateValidity";

export interface WorkerCreateController {
  state: WorkerCreateDraftState;
  options: AsyncState<WorkerCreateOptions>;
  modelResources: AsyncState<EffectiveResource[]>;
  runtimeBundles: AsyncState<EnvBundleSummary[]>;
  credentialBundles: AsyncState<EnvBundleSummary[]>;
  skills: AsyncState<InstalledSkill[]>;
  repositories: RepositoryData[];
  validity: WorkerCreateValidity;
  patchDraft: (patch: Partial<WorkerSpecDraft>) => void;
  changeWorkerType: (slug: string, schemaVersion: number) => void;
  setLifecycle: (policy: string, minutes: number) => void;
  setFillPrompt: (prompt: string) => void;
  fillWithAI: (prompt: string) => Promise<void>;
  goToStep: (step: WorkerCreateStepId) => Promise<void>;
  runPreflight: () => Promise<WorkerPreflightResult | null>;
  createWorker: () => Promise<PodData | null>;
  reset: () => void;
}

interface ControllerInput {
  state: WorkerCreateDraftState;
  dispatch: Dispatch<WorkerCreateDraftAction>;
  options: AsyncState<WorkerCreateOptions>;
  modelResources: AsyncState<EffectiveResource[]>;
  runtimeBundles: AsyncState<EnvBundleSummary[]>;
  credentialBundles: AsyncState<EnvBundleSummary[]>;
  skills: AsyncState<InstalledSkill[]>;
  repositories: RepositoryData[];
  validity: WorkerCreateValidity;
  initial: Partial<WorkerSpecDraft>;
  fillWithAI: WorkerCreateController["fillWithAI"];
  goToStep: WorkerCreateController["goToStep"];
  runPreflight: WorkerCreateController["runPreflight"];
  createWorker: WorkerCreateController["createWorker"];
}

export function assembleWorkerCreateController(
  input: ControllerInput,
): WorkerCreateController {
  const { dispatch } = input;
  return {
    state: input.state,
    options: input.options,
    modelResources: input.modelResources,
    runtimeBundles: input.runtimeBundles,
    credentialBundles: input.credentialBundles,
    skills: input.skills,
    repositories: input.repositories,
    validity: input.validity,
    patchDraft: (patch) => dispatch({ type: "patch_draft", patch }),
    changeWorkerType: (slug, schemaVersion) => dispatch({
      type: "change_worker_type",
      workerTypeSlug: slug,
      schemaVersion,
    }),
    setLifecycle: (policy, minutes) => dispatch({
      type: "set_lifecycle",
      terminationPolicy: policy,
      idleTimeoutMinutes: minutes,
    }),
    setFillPrompt: (prompt) => dispatch({ type: "set_fill_prompt", prompt }),
    fillWithAI: input.fillWithAI,
    goToStep: input.goToStep,
    runPreflight: input.runPreflight,
    createWorker: input.createWorker,
    reset: () => dispatch({ type: "reset", draft: input.initial }),
  };
}

export function workerCreateInitialDraft(params: {
  initialWorkerTypeSlug?: string;
  initialTask?: string;
  initialRepositoryId?: number | null;
}): Partial<WorkerSpecDraft> {
  return {
    worker_type_slug: params.initialWorkerTypeSlug ?? "",
    repository_id: params.initialRepositoryId ?? undefined,
    initial_task: params.initialTask ?? "",
  };
}

export function workerCreateLoadable<T>(
  loading: boolean,
  error: string | null,
  data: T,
): AsyncState<T> {
  if (loading) return { status: "loading" };
  if (error) return { status: "error", error };
  return { status: "ready", data };
}

export function workerPreflightHasBlockingIssues(
  result: WorkerPreflightResult,
): boolean {
  return result.issues.some((issue) => issue.severity === "blocking");
}

export function workerCreateError(error: unknown): Error {
  return error instanceof Error ? error : new Error("Worker creation failed");
}
