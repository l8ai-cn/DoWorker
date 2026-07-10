import { useMemo } from "react";
import { buildAgentfileLayer } from "@/lib/agentfile-layer";
import type { PodMode } from "@/lib/pod-modes";
import type { RepositoryData } from "@/lib/api";
import type { CustomEnvEntry } from "@/components/settings/AgentCredentialsSettings/credentialForms/types";
import type { KnowledgeMountSelection } from "@/lib/api/facade/knowledgeBaseApi";

export function useCreatePodGeneratedLayer(params: {
  configValues?: Record<string, unknown>;
  repositories: RepositoryData[];
  selectedRepository: number | null;
  selectedBranch: string;
  interactionMode: PodMode;
  selectedRuntimeBundleNames: string[];
  selectedSkillSlugs: string[];
  selectedKnowledgeMounts: KnowledgeMountSelection[];
  tokenBudget: number | null;
  prompt: string;
  customEnv: CustomEnvEntry[];
}) {
  const {
    configValues,
    repositories,
    selectedRepository,
    selectedBranch,
    interactionMode,
    selectedRuntimeBundleNames,
    selectedSkillSlugs,
    selectedKnowledgeMounts,
    tokenBudget,
    prompt,
    customEnv,
  } = params;

  return useMemo(() => {
    const repoSlug = selectedRepository
      ? repositories.find((r) => r.id === selectedRepository)?.slug
      : undefined;
    return buildAgentfileLayer({
      configValues: configValues ?? {},
      repositorySlug: repoSlug,
      branchName: selectedBranch || undefined,
      interactionMode,
      runtimeBundleNames: selectedRuntimeBundleNames.length > 0
        ? selectedRuntimeBundleNames
        : undefined,
      skillSlugs: selectedSkillSlugs.length > 0 ? selectedSkillSlugs : undefined,
      knowledgeMounts: selectedKnowledgeMounts.length > 0 ? selectedKnowledgeMounts : undefined,
      tokenBudget,
      prompt: prompt.trim() || undefined,
      customEnv: customEnv.length > 0
        ? customEnv.map((e) => ({ key: e.key, value: e.value }))
        : undefined,
    });
  }, [
    configValues,
    repositories,
    selectedRepository,
    selectedBranch,
    interactionMode,
    selectedRuntimeBundleNames,
    selectedSkillSlugs,
    selectedKnowledgeMounts,
    tokenBudget,
    prompt,
    customEnv,
  ]);
}
