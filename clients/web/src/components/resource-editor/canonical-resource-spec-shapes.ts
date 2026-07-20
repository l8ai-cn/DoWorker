import type { ResourceEditorKind } from "./resource-editor-types";

export type CanonicalShape =
  | "any"
  | "boolean"
  | "integer"
  | "string"
  | CanonicalArrayShape
  | CanonicalMapShape
  | CanonicalObjectShape;

export interface CanonicalArrayShape {
  type: "array";
  item: CanonicalShape;
}

export interface CanonicalMapShape {
  type: "map";
  value: CanonicalShape;
}

export interface CanonicalObjectShape {
  type: "object";
  fields: Record<string, CanonicalShape>;
  optional: readonly string[];
}

const array = (item: CanonicalShape): CanonicalArrayShape => ({
  type: "array",
  item,
});
const map = (value: CanonicalShape): CanonicalMapShape => ({
  type: "map",
  value,
});
const object = (
  fields: Record<string, CanonicalShape>,
  optional: readonly string[] = [],
): CanonicalObjectShape => ({ type: "object", fields, optional });

const reference = object({
  apiVersion: "string",
  kind: "string",
  namespace: "string",
  name: "string",
  revision: "integer",
}, ["apiVersion", "namespace", "revision"]);

const resourceLimits = object({
  cpuRequestMilliCPU: "integer",
  cpuLimitMilliCPU: "integer",
  memoryRequestBytes: "integer",
  memoryLimitBytes: "integer",
  storageRequestBytes: "integer",
  storageLimitBytes: "integer",
  gpuRequest: "integer",
  gpuLimit: "integer",
}, ["storageRequestBytes", "storageLimitBytes", "gpuRequest", "gpuLimit"]);

export const canonicalResourceSpecShapes = {
  WorkerTemplate: object({
    optionsRevision: "string",
    workerType: "string",
    modelRef: reference,
    toolRefs: map(reference),
    runtime: object({
      runtimeImageId: "integer",
      placementPolicy: "string",
      computeTargetRef: reference,
      deploymentMode: "string",
      resourceProfileRef: reference,
      customResources: resourceLimits,
    }, ["resourceProfileRef", "customResources"]),
    typeConfig: object({
      schemaVersion: "integer",
      values: map("any"),
      secretRefs: map(reference),
      interactionMode: "string",
      automationLevel: "string",
    }),
    workspace: object({
      repositoryRef: reference,
      branch: "string",
      skillRefs: array(reference),
      knowledgeMounts: array(object({
        ref: reference,
        mode: "string",
      })),
      environmentBundleRefs: array(reference),
      configDocumentBindings: array(object({
        documentId: "string",
        configBundleRef: reference,
      })),
      instructions: "string",
    }, ["repositoryRef"]),
    lifecycle: object({
      terminationPolicy: "string",
      idleTimeoutMinutes: "integer",
    }),
    metadata: object({ alias: "string" }),
  }, ["modelRef"]),
  Worker: object({
    workerTemplateRef: reference,
    promptRef: reference,
    inputs: map("string"),
    alias: "string",
  }, ["promptRef"]),
  Prompt: object({
    content: "string",
    variables: map(object({
      required: "boolean",
      default: "string",
    }, ["default"])),
  }),
  Expert: object({
    workerTemplateRef: reference,
    promptRef: reference,
    description: "string",
    category: "string",
    releaseNotes: "string",
  }, ["promptRef"]),
  Workflow: object({
    workerTemplateRef: reference,
    promptRef: reference,
    inputs: map("string"),
    executionMode: "string",
    cronExpression: "string",
    sandboxStrategy: "string",
    sessionPersistence: "boolean",
    concurrencyPolicy: "string",
    maxConcurrentRuns: "integer",
    maxRetainedRuns: "integer",
    timeoutMinutes: "integer",
    idleTimeoutSeconds: "integer",
    callbackUrl: "string",
  }, ["cronExpression", "callbackUrl"]),
  GoalLoop: object({
    workerTemplateRef: reference,
    description: "string",
    objective: "string",
    acceptanceCriteria: array("string"),
    verificationCommand: "string",
    maxIterations: "integer",
    tokenBudget: "integer",
    timeoutMinutes: "integer",
    noProgressLimit: "integer",
    sameErrorLimit: "integer",
    escalationPolicy: "string",
    loopProgram: object({
      canonicalSource: "string",
      customBlock: object({
        nodeId: "string",
        definitionId: "string",
        slug: "string",
        version: "integer",
        definitionDigest: "string",
      }, []),
    }, ["customBlock"]),
  }, ["tokenBudget", "loopProgram"]),
  ModelBinding: object({ resourceId: "integer" }),
  Repository: object({ repositoryId: "integer" }),
  Skill: object({ skillId: "integer" }),
  KnowledgeBase: object({ knowledgeBaseId: "integer" }),
  EnvironmentBundle: object({ environmentBundleId: "integer" }),
  ComputeTarget: object({ computeTargetId: "integer" }),
  ResourceProfile: object({ resourceProfileId: "integer" }),
  ToolBinding: object({ modelRef: reference }),
} satisfies Record<ResourceEditorKind, CanonicalObjectShape>;
