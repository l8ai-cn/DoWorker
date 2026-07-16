import type {
  ResourceManifest,
  ResourceReference,
} from "./resource-manifest-types";

export interface PromptVariableDraft {
  required: boolean;
  default?: string;
}

export interface PromptSpec {
  content: string;
  variables: Record<string, PromptVariableDraft>;
}

export type PromptDraft = ResourceManifest<PromptSpec> & {
  kind: "Prompt";
};

export interface ExpertSpec {
  workerTemplateRef: ResourceReference;
  promptRef?: ResourceReference;
  description: string;
  category: string;
  releaseNotes: string;
}

export type ExpertDraft = ResourceManifest<ExpertSpec> & {
  kind: "Expert";
};

export interface WorkflowSpec {
  workerTemplateRef: ResourceReference;
  promptRef: ResourceReference;
  inputs: Record<string, string>;
  executionMode: string;
  cronExpression?: string;
  sandboxStrategy: string;
  sessionPersistence: boolean;
  concurrencyPolicy: string;
  maxConcurrentRuns: number;
  maxRetainedRuns: number;
  timeoutMinutes: number;
  idleTimeoutSeconds: number;
  callbackUrl?: string;
}

export type WorkflowDraft = ResourceManifest<WorkflowSpec> & {
  kind: "Workflow";
};

export type GoalLoopIntegerDraft = number | string;

export interface GoalLoopSpec {
  workerTemplateRef: ResourceReference;
  description: string;
  objective: string;
  acceptanceCriteria: string[];
  verificationCommand: string;
  maxIterations: GoalLoopIntegerDraft;
  tokenBudget?: GoalLoopIntegerDraft;
  timeoutMinutes: GoalLoopIntegerDraft;
  noProgressLimit: GoalLoopIntegerDraft;
  sameErrorLimit: GoalLoopIntegerDraft;
  escalationPolicy: "pause" | "fail";
}

export type GoalLoopDraft = ResourceManifest<GoalLoopSpec> & {
  kind: "GoalLoop";
};

export type DefinitionResourceDraft =
  | PromptDraft
  | ExpertDraft
  | WorkflowDraft
  | GoalLoopDraft;
export type DefinitionResourceKind = DefinitionResourceDraft["kind"];
