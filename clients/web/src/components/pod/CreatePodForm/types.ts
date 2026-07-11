import type { PodData } from "@/lib/api/facade/pod";

export type PodCreationScenario = "workspace" | "ticket";

export interface TicketContext {
  id: number;
  slug: string;
  title: string;
  description?: string;
  repositoryId?: number;
}

export interface ScenarioContext {
  ticket?: TicketContext;
}

export type PromptGenerator = (context: ScenarioContext) => string;

export interface CreatePodFormConfig {
  scenario: PodCreationScenario;
  context?: ScenarioContext;
  promptGenerator?: PromptGenerator;
  promptPlaceholder?: string;
  // Workspace recipe pre-fill: agent is selected when the runner exposes it,
  // prompt seeds the input. Both override saved-preference auto-fill.
  initialAgentSlug?: string;
  initialPrompt?: string;
  /** Test / deep-link hook: open wizard on a specific step. */
  initialWizardStep?: 1 | 2 | 3 | 4;
  initialExpertSlug?: string;
  onSuccess?: (pod: PodData) => void;
  onError?: (error: Error) => void;
  onCancel?: () => void;
}

export interface CreatePodFormProps {
  config: CreatePodFormConfig;
  enabled?: boolean;
  className?: string;
}
