import { useMemo } from "react";
import type { EffectiveResource } from "@/lib/api/facade/aiResource";
import type { EnvBundleSummary, InstalledSkill } from "@/lib/api";
import { useRepoSkills } from "./useCreatePodFormEffects";
import { useWorkerModelResources } from "./useWorkerModelResources";
import { useWorkerCreateEnvBundles } from "./useWorkerCreateEnvBundles";
import type { AsyncState } from "./workerCreateDraft";
import { workerCreateLoadable } from "./workerCreateController";

interface WorkerCreateDependencies {
  modelResources: AsyncState<EffectiveResource[]>;
  runtimeBundles: AsyncState<EnvBundleSummary[]>;
  credentialBundles: AsyncState<EnvBundleSummary[]>;
  skills: AsyncState<InstalledSkill[]>;
}

export function useWorkerCreateDependencies(
  workerTypeSlug: string,
  repositoryId?: number,
): WorkerCreateDependencies {
  const model = useWorkerModelResources(workerTypeSlug);
  const bundles = useWorkerCreateEnvBundles(workerTypeSlug);
  const skills = useRepoSkills(repositoryId ?? null);
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
        skills.loadingSkills,
        skills.skillLoadError,
        skills.repoSkills,
      ),
    [skills.loadingSkills, skills.repoSkills, skills.skillLoadError],
  );

  return {
    modelResources,
    runtimeBundles: bundles.runtime,
    credentialBundles: bundles.credential,
    skills: installedSkills,
  };
}
