import { useCallback, useState } from "react";
import type { PodData } from "@/lib/api";
import type { DestroyPolicy } from "../CreatePodForm/podLifecycleOptions";
import type { KnowledgeMountSelection } from "@/lib/api/facade/knowledgeBaseApi";
import { submitCreatePod } from "./useCreatePodFormSubmit";

export function useCreatePodSubmitAction(params: {
  selectedAgent: string | null;
  selectedRepository: number | null;
  selectedBranch: string;
  selectedRuntimeBundleNames: string[];
  selectedSkillSlugs: string[];
  alias: string;
  perpetual: boolean;
  destroyPolicy: DestroyPolicy;
  destroyAfterMinutes: number;
  selectedKnowledgeMounts: KnowledgeMountSelection[];
  modelResourceId: number | null;
  agentfileLayer: string;
  automationLevel: string;
  validate: () => boolean;
  setLastChoices: (choices: Record<string, unknown>) => void;
  onSuccess?: (pod: PodData) => void;
}) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [warning, setWarning] = useState<string | null>(null);

  const submit = useCallback(
    async (
      selectedRunnerId: number | null | undefined,
      pluginConfig: Record<string, unknown>,
      options?: { ticketSlug?: string; cols?: number; rows?: number },
    ): Promise<PodData | null> => {
      void pluginConfig;
      if (!params.validate()) return null;
      if (!params.selectedAgent) {
        setError("Please select an agent");
        return null;
      }
      setLoading(true);
      setError(null);
      setWarning(null);
      try {
        const result = await submitCreatePod({
          selectedAgent: params.selectedAgent,
          alias: params.alias,
          perpetual: params.perpetual,
          selectedRunnerId,
          agentfileLayer: params.agentfileLayer || undefined,
          automationLevel: params.automationLevel,
          repositoryId: params.selectedRepository,
          modelResourceId: params.modelResourceId,
          options,
        });
        if (result) {
          params.setLastChoices({
            lastAgentSlug: params.selectedAgent,
            lastRepositoryId: params.selectedRepository,
            lastRuntimeBundleNames: params.selectedRuntimeBundleNames,
            lastBranchName: params.selectedBranch || null,
            lastSkillSlugs: params.selectedSkillSlugs,
            lastDestroyPolicy: params.destroyPolicy,
            lastDestroyAfterMinutes: params.destroyAfterMinutes,
            lastKnowledgeMounts: params.selectedKnowledgeMounts,
          });
          if (result.warning) setWarning(result.warning);
          params.onSuccess?.(result.pod);
        }
        return result?.pod ?? null;
      } catch (err) {
        const message = err instanceof Error ? err.message : "Failed to create pod";
        setError(message);
        console.error("Failed to create pod:", err);
        return null;
      } finally {
        setLoading(false);
      }
    },
    [params],
  );

  return { loading, error, warning, submit };
}
