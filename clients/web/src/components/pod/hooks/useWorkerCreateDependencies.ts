import { useMemo } from "react";
import type { EffectiveResource } from "@/lib/api/facade/aiResource";
import type { EnvBundleSummary } from "@/lib/api";
import type { WorkerCreateOptions } from "@/lib/api/facade/podConnect";
import { useWorkerModelResources } from "./useWorkerModelResources";
import { useWorkerCreateEnvBundles } from "./useWorkerCreateEnvBundles";
import { useWorkerCreateModelResources } from "./useWorkerCreateModelResources";
import type { AsyncState } from "./workerCreateDraft";
import { workerCreateLoadable } from "./workerCreateController";
import { useWorkerSkills } from "./useWorkerSkills";
import type { WorkerSkillOption } from "../CreatePodForm/workerSkillOption";

type WorkerTypeOption = WorkerCreateOptions["worker_types"][number];

interface WorkerCreateDependencies {
  modelResources: AsyncState<EffectiveResource[]>;
  toolModelResources: AsyncState<EffectiveResource[]>;
  runtimeBundles: AsyncState<EnvBundleSummary[]>;
  credentialBundles: AsyncState<EnvBundleSummary[]>;
  configBundles: AsyncState<EnvBundleSummary[]>;
  skills: AsyncState<WorkerSkillOption[]>;
}

export function useWorkerCreateDependencies(
  workerType: WorkerTypeOption | undefined,
  repositoryId?: number,
): WorkerCreateDependencies {
  const model = useWorkerModelResources(
    workerType?.slug,
    null,
    true,
    workerType
      ? {
        required: workerType.requires_model_resource,
        protocolAdapters: workerType.model_protocol_adapters,
      }
      : undefined,
  );
  const bundles = useWorkerCreateEnvBundles(workerType?.slug ?? "");
  const skills = useWorkerSkills(repositoryId ?? null);
  const modelResources = useMemo(
    () =>
      workerCreateLoadable(
        model.loadingModelResources,
        model.modelResourceError,
        model.modelResources,
      ),
    [
      model.loadingModelResources,
      model.modelResourceError,
      model.modelResources,
    ],
  );
  const installedSkills = useMemo(
    () =>
      workerCreateLoadable(
        skills.loading,
        skills.error,
        skills.skills,
      ),
    [skills.loading, skills.skills, skills.error],
  );
  const toolModelResources = useMemo(
    () =>
      workerCreateLoadable(
        model.loadingModelResources,
        model.modelResourceError,
        model.toolModelResources,
      ),
    [
      model.loadingModelResources,
      model.modelResourceError,
      model.toolModelResources,
    ],
  );

  return {
    modelResources,
    toolModelResources,
    runtimeBundles: bundles.runtime,
    credentialBundles: bundles.credential,
    configBundles: bundles.config,
    skills: installedSkills,
  };
}
