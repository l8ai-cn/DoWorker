import { PodData } from "@/lib/api";
import type { CustomEnvEntry } from "@/components/settings/AgentCredentialsSettings/credentialForms/types";
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
}

export interface CreatePodFormState {
  // Selection state (order: Runner -> Agent -> Others)
  selectedAgent: string | null;
  selectedRepository: number | null;
  selectedBranch: string;
  // Credential bundle (kind='credential') — single-select. Empty string
  // means "use the Agent's default authentication" (OAuth / CLI login etc.).
  selectedCredentialName: string;
  // Runtime bundle names (kind='runtime') — ordered multi-select. Each name
  // maps to a `USE_ENV_BUNDLE "..."` directive emitted AFTER the credential
  // line, so runtime preferences (model, log level, proxy) can override
  // credential defaults when keys conflict.
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
  // Virtual API key binding for quota/billing attribution. Null = agent default.
  selectedVirtualKeyId: number | null;
  // Per-Worker custom environment variables. Emitted as `ENV KEY = "value"`
  // lines in the generated AgentFile layer.
  customEnv: CustomEnvEntry[];

  // EnvBundles (credential + runtime kinds) available for the selected agent
  envBundles: EnvBundleSummary[];
  loadingBundles: boolean;
  repoSkills: InstalledSkill[];
  loadingSkills: boolean;

  // Actions
  setSelectedAgent: (slug: string | null) => void;
  setSelectedRepository: (id: number | null) => void;
  setSelectedBranch: (branch: string) => void;
  setSelectedCredentialName: (name: string) => void;
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
  setSelectedVirtualKeyId: (id: number | null) => void;
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
