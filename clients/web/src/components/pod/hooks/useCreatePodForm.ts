import { useState, useCallback, useMemo } from "react";
import { PodData, AgentData, RepositoryData } from "@/lib/api";
import { usePodCreationStore } from "@/stores/podCreation";
import { POD_MODE_PTY } from "@/lib/pod-modes";
import { usePrefsAutoFill, useEnvBundles, useRepoSkills } from "./useCreatePodFormEffects";
import { useCreatePodGeneratedLayer } from "./useCreatePodGeneratedLayer";
import { useCreatePodInteractionMode } from "./useCreatePodInteractionMode";
import { useCreatePodSubmitAction } from "./useCreatePodSubmitAction";
import { useCreatePodValidation } from "./useCreatePodValidation";
import { requiresModelResource, useWorkerModelResources } from "./useWorkerModelResources";
import type { CreatePodFormState } from "./useCreatePodFormTypes";
import type { DestroyPolicy } from "../CreatePodForm/podLifecycleOptions";
import type { KnowledgeMountSelection } from "@/lib/api/facade/knowledgeBaseApi";
import type { CustomEnvEntry } from "@/components/settings/envBundleCredentialForms/types";

export type { CreatePodFormState, FormValidationErrors } from "./useCreatePodFormTypes";

export function useCreatePodForm(
  availableAgents: AgentData[],
  repositories: RepositoryData[],
  onSuccess?: (pod: PodData) => void,
  configValues?: Record<string, unknown>,
  overrides?: { repositoryId?: number | null },
): CreatePodFormState {
  const { lastDestroyAfterMinutes, lastDestroyPolicy, lastKnowledgeMounts, setLastChoices } =
    usePodCreationStore();

  const [selectedAgent, setSelectedAgent] = useState<string | null>(null);
  const [selectedRepository, setSelectedRepository] = useState<number | null>(null);
  const [selectedBranch, setSelectedBranch] = useState<string>("");
  const [automationLevel, setAutomationLevel] = useState<string>("autonomous");
  const [prompt, setPrompt] = useState<string>("");
  const [alias, setAlias] = useState<string>("");
  const [perpetual, setPerpetual] = useState(false);
  const [destroyPolicy, setDestroyPolicy] = useState<DestroyPolicy>(lastDestroyPolicy);
  const [destroyAfterMinutes, setDestroyAfterMinutes] = useState(lastDestroyAfterMinutes);
  const [selectedKnowledgeMounts, setSelectedKnowledgeMounts] =
    useState<KnowledgeMountSelection[]>(lastKnowledgeMounts ?? []);
  const [tokenBudget, setTokenBudget] = useState<number | null>(null);
  const [customEnv, setCustomEnv] = useState<CustomEnvEntry[]>([]);
  const [rawLayerMode, setRawLayerModeState] = useState(false);
  const [rawLayerText, setRawLayerText] = useState("");
  const selectedAgentSlug = useMemo(() => {
    if (!selectedAgent) return "";
    return availableAgents.find((a) => a.slug === selectedAgent)?.slug || "";
  }, [selectedAgent, availableAgents]);

  const bundles = useEnvBundles(selectedAgent);
  const modelResources = useWorkerModelResources(selectedAgentSlug);
  const skills = useRepoSkills(selectedRepository);
  const mode = useCreatePodInteractionMode(selectedAgent, availableAgents, automationLevel);

  const handleSelectedAgent = useCallback((slug: string | null) => {
    setSelectedAgent(slug);
    bundles.setSelectedRuntimeBundleNames([]);
    modelResources.setSelectedModelResourceId(null);
    skills.setSelectedSkillSlugs([]);
    mode.setInteractionMode(POD_MODE_PTY);
    if (!slug) bundles.setEnvBundles([]);
  }, [bundles, modelResources, skills, mode]);

  const handleSelectedRepository = useCallback((id: number | null) => {
    setSelectedRepository(id);
    const repo = id ? repositories.find((r) => r.id === id) : undefined;
    setSelectedBranch(repo?.default_branch ?? "");
  }, [repositories]);

  const prefsInitializedRef = usePrefsAutoFill(
    availableAgents, repositories, handleSelectedAgent, handleSelectedRepository, setSelectedBranch,
    overrides,
  );

  const isValid = useMemo(() => {
    if (!selectedAgent || !selectedAgentSlug) return false;
    if (bundles.loadingBundles || bundles.bundleLoadError) return false;
    if (!requiresModelResource(selectedAgentSlug)) return true;
    return Boolean(
      modelResources.selectedModelResourceId &&
        modelResources.selectedModelResource &&
        !modelResources.loadingModelResources &&
        !modelResources.modelResourceError,
    );
  }, [selectedAgent, selectedAgentSlug, bundles.loadingBundles, bundles.bundleLoadError, modelResources]);

  const validation = useCreatePodValidation({
    selectedAgent,
    selectedRepository,
    selectedBranch,
    customEnv,
    bundleLoadError: bundles.bundleLoadError,
    selectedAgentSlug,
    modelResourceError: modelResources.modelResourceError,
    loadingModelResources: modelResources.loadingModelResources,
    selectedModelResourceId: modelResources.selectedModelResourceId,
    selectedModelResource: modelResources.selectedModelResource,
  });

  const reset = useCallback(() => {
    setSelectedAgent(null);
    setSelectedRepository(null);
    setSelectedBranch("");
    bundles.setSelectedRuntimeBundleNames([]);
    modelResources.setSelectedModelResourceId(null);
    skills.setSelectedSkillSlugs([]);
    bundles.setEnvBundles([]);
    mode.setInteractionMode(POD_MODE_PTY);
    setAutomationLevel("autonomous");
    setPrompt("");
    setAlias("");
    setPerpetual(false);
    setDestroyPolicy("manual");
    setDestroyAfterMinutes(120);
    setSelectedKnowledgeMounts([]);
    setTokenBudget(null);
    setCustomEnv([]);
    validation.setValidationErrors({});
    setRawLayerModeState(false);
    setRawLayerText("");
    prefsInitializedRef.current = false;
  }, [bundles, modelResources, skills, mode, prefsInitializedRef, validation]);

  const generatedLayer = useCreatePodGeneratedLayer({
    configValues,
    repositories,
    selectedRepository,
    selectedBranch,
    interactionMode: mode.interactionMode,
    selectedRuntimeBundleNames: bundles.selectedRuntimeBundleNames,
    selectedSkillSlugs: skills.selectedSkillSlugs,
    selectedKnowledgeMounts,
    tokenBudget,
    prompt,
    customEnv,
  });

  const agentfileLayer = rawLayerMode ? rawLayerText : generatedLayer;

  const setRawLayerMode = useCallback((enabled: boolean) => {
    if (enabled && !rawLayerText) {
      setRawLayerText(generatedLayer);
    }
    setRawLayerModeState(enabled);
  }, [generatedLayer, rawLayerText]);

  const submission = useCreatePodSubmitAction({
    selectedAgent,
    selectedRepository,
    selectedBranch,
    selectedRuntimeBundleNames: bundles.selectedRuntimeBundleNames,
    selectedSkillSlugs: skills.selectedSkillSlugs,
    alias,
    perpetual,
    destroyPolicy,
    destroyAfterMinutes,
    selectedKnowledgeMounts,
    modelResourceId: modelResources.selectedModelResourceId,
    agentfileLayer,
    automationLevel,
    validate: validation.validate,
    setLastChoices,
    onSuccess,
  });

  return {
    selectedAgent, selectedRepository, selectedBranch,
    selectedModelResourceId: modelResources.selectedModelResourceId,
    selectedRuntimeBundleNames: bundles.selectedRuntimeBundleNames,
    selectedSkillSlugs: skills.selectedSkillSlugs,
    interactionMode: mode.interactionMode, automationLevel, prompt, alias, perpetual,
    destroyPolicy, destroyAfterMinutes, selectedKnowledgeMounts, tokenBudget,
    customEnv,
    modelResources: modelResources.modelResources,
    loadingModelResources: modelResources.loadingModelResources,
    modelResourceError: modelResources.modelResourceError,
    envBundles: bundles.envBundles, loadingBundles: bundles.loadingBundles,
    bundleLoadError: bundles.bundleLoadError,
    repoSkills: skills.repoSkills, loadingSkills: skills.loadingSkills,
    setSelectedAgent: handleSelectedAgent,
    setSelectedRepository: handleSelectedRepository,
    setSelectedBranch,
    setSelectedModelResourceId: modelResources.setSelectedModelResourceId,
    setSelectedRuntimeBundleNames: bundles.setSelectedRuntimeBundleNames,
    setSelectedSkillSlugs: skills.setSelectedSkillSlugs,
    setInteractionMode: mode.setInteractionMode, setAutomationLevel, setPrompt, setAlias, setPerpetual,
    selectedAgentSlug, supportedModes: mode.supportedModes,
    setDestroyPolicy, setDestroyAfterMinutes, setSelectedKnowledgeMounts, setTokenBudget,
    setCustomEnv,
    loading: submission.loading,
    error: submission.error,
    warning: submission.warning,
    validationErrors: validation.validationErrors,
    isValid,
    reset,
    validate: validation.validate,
    submit: submission.submit,
    rawLayerMode, rawLayerText, agentfileLayer, setRawLayerMode, setRawLayerText,
  };
}
