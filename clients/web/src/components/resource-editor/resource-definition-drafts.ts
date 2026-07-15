import {
  RESOURCE_API_VERSION,
  type ExpertDraft,
  type GoalLoopDraft,
  type PromptDraft,
  type ResourceMetadata,
  type WorkflowDraft,
} from "./resource-editor-types";

export function createPromptDraft(namespace: string): PromptDraft {
  return {
    apiVersion: RESOURCE_API_VERSION,
    kind: "Prompt",
    metadata: resourceMetadata(namespace),
    spec: { content: "", variables: {} },
  };
}

export function createExpertDraft(namespace: string): ExpertDraft {
  return {
    apiVersion: RESOURCE_API_VERSION,
    kind: "Expert",
    metadata: resourceMetadata(namespace),
    spec: {
      workerTemplateRef: { kind: "WorkerTemplate", name: "" },
      description: "",
      category: "",
      releaseNotes: "",
    },
  };
}

export function createWorkflowDraft(namespace: string): WorkflowDraft {
  return {
    apiVersion: RESOURCE_API_VERSION,
    kind: "Workflow",
    metadata: resourceMetadata(namespace),
    spec: {
      workerTemplateRef: { kind: "WorkerTemplate", name: "" },
      promptRef: { kind: "Prompt", name: "" },
      inputs: {},
      executionMode: "direct",
      cronExpression: "",
      sandboxStrategy: "fresh",
      sessionPersistence: false,
      concurrencyPolicy: "skip",
      maxConcurrentRuns: 1,
      maxRetainedRuns: 30,
      timeoutMinutes: 60,
      idleTimeoutSeconds: 30,
      callbackUrl: "",
    },
  };
}

export function createGoalLoopDraft(namespace: string): GoalLoopDraft {
  return {
    apiVersion: RESOURCE_API_VERSION,
    kind: "GoalLoop",
    metadata: resourceMetadata(namespace),
    spec: {
      workerTemplateRef: { kind: "WorkerTemplate", name: "" },
      description: "",
      objective: "",
      acceptanceCriteria: [""],
      verificationCommand: "",
      maxIterations: 10,
      tokenBudget: undefined,
      timeoutMinutes: 60,
      noProgressLimit: 3,
      sameErrorLimit: 2,
      escalationPolicy: "pause",
    },
  };
}

function resourceMetadata(namespace: string): ResourceMetadata {
  return {
    name: "",
    namespace,
    displayName: "",
    labels: {},
  };
}
