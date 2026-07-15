import { useMemo } from "react";
import type { EffectiveResource } from "@/lib/api/facade/aiResource";
import type { EnvBundleSummary, InstalledSkill } from "@/lib/api";
import { useRepoSkills } from "./useCreatePodFormEffects";
import { useWorkerCreateEnvBundles } from "./useWorkerCreateEnvBundles";
import { useWorkerCreateModelResources } from "./useWorkerCreateModelResources";
import type { AsyncState } from "./workerCreateDraft";
import { workerCreateLoadable } from "./workerCreateController";
import type { ProviderDefinition } from "@/lib/api/facade/aiResource";

interface WorkerCreateDependencies {
  modelResources: AsyncState<EffectiveResource[]>;
  modelProviders: AsyncState<ProviderDefinition[]>;
  runtimeBundles: AsyncState<EnvBundleSummary[]>;
  credentialBundles: AsyncState<EnvBundleSummary[]>;
  skills: AsyncState<InstalledSkill[]>;
}

export function useWorkerCreateDependencies(
  workerTypeSlug: string,
  repositoryId?: number,
): WorkerCreateDependencies {
  const model = useWorkerCreateModelResources();
  const bundles = useWorkerCreateEnvBundles(workerTypeSlug);
  const skills = useRepoSkills(repositoryId ?? null);
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
    modelResources: model.modelResources,
    modelProviders: model.modelProviders,
    runtimeBundles: bundles.runtime,
    credentialBundles: bundles.credential,
    skills: installedSkills,
  };
}
