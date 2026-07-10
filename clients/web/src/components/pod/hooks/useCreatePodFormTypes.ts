import { PodData } from "@/lib/api";
import type { CustomEnvEntry } from "@/components/settings/envBundleCredentialForms/types";
import type { EffectiveResource } from "@/lib/api/facade/aiResource";
import type { KnowledgeMountSelection } from "@/lib/api/facade/knowledgeBaseApi";
import type { PodMode } from "@/lib/pod-modes";
import type { EnvBundleSummary } from "@/lib/viewModels/envBundleSummary";
import type { InstalledSkill } from "@/lib/viewModels/extension";
import type { DestroyPolicy } from "../CreatePodForm/podLifecycleOptions";

/**
 * Validation errors for the form
 */
export interface FormValidationErrors {
  runner?: string;
  agent?: string;
  repository?: string;
  branch?: string;
  prompt?: string;
  env?: string;
  modelResource?: string;
  runtimeBundles?: string;
}

export interface CreatePodFormState {
  selectedAgent: string | null;
  selectedRepository: number | null;
  selectedBranch: string;
  selectedModelResourceId: number | null;
  selectedRuntimeBundleNames: string[];
  selectedSkillSlugs: string[];
  interactionMode: PodMode;
  // Unified permission/automation tier: interactive | auto_edit | autonomous.
  // autonomous (default) forces ACP so the Worker runs non-interactively.
  automationLevel: string;
  prompt: string;
  alias: string;
  perpetual: boolean;
  destroyPolicy: DestroyPolicy;
  destroyAfterMinutes: number;
  selectedKnowledgeMounts: KnowledgeMountSelection[];
  // Optional per-Worker token budget cap. null = no cap (org quota still
  // applies). Emitted as CONFIG token_budget in the AgentFile layer.
  tokenBudget: number | null;
  customEnv: CustomEnvEntry[];

  modelResources: EffectiveResource[];
  loadingModelResources: boolean;
  modelResourceError: string | null;
  envBundles: EnvBundleSummary[];
  loadingBundles: boolean;
  bundleLoadError: string | null;
  repoSkills: InstalledSkill[];
  loadingSkills: boolean;

  // Actions
  setSelectedAgent: (slug: string | null) => void;
  setSelectedRepository: (id: number | null) => void;
  setSelectedBranch: (branch: string) => void;
  setSelectedModelResourceId: (id: number | null) => void;
  setSelectedRuntimeBundleNames: (names: string[]) => void;
  setSelectedSkillSlugs: (slugs: string[]) => void;
  setInteractionMode: (mode: PodMode) => void;
  setAutomationLevel: (level: string) => void;
  setPrompt: (prompt: string) => void;
  setAlias: (alias: string) => void;
  setPerpetual: (perpetual: boolean) => void;
  setDestroyPolicy: (policy: DestroyPolicy) => void;
  setDestroyAfterMinutes: (minutes: number) => void;
  setSelectedKnowledgeMounts: (mounts: KnowledgeMountSelection[]) => void;
  setTokenBudget: (budget: number | null) => void;
  setCustomEnv: (entries: CustomEnvEntry[]) => void;

  // AgentFile Layer
  rawLayerMode: boolean;
  rawLayerText: string;
  agentfileLayer: string;
  setRawLayerMode: (enabled: boolean) => void;
  setRawLayerText: (text: string) => void;

  // Computed
  selectedAgentSlug: string;
  supportedModes: string[]; // parsed from agent type's supported_modes

  // Form state
  loading: boolean;
  error: string | null;
  // Non-fatal note returned by the server (e.g. "pod created, but X is degraded").
  // Distinct from `error`, which represents a request failure.
  warning: string | null;
  validationErrors: FormValidationErrors;
  isValid: boolean;

  // Actions
  reset: () => void;
  validate: () => boolean;
  submit: (
    selectedRunnerId: number | null | undefined,
    pluginConfig: Record<string, unknown>,
    options?: { ticketSlug?: string; cols?: number; rows?: number }
  ) => Promise<PodData | null>;
}
