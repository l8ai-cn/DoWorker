import type {
  ResourceManifest,
  ResourceReference,
} from "./resource-manifest-types";

export interface WorkerTemplateResources {
  cpuRequestMilliCPU: number;
  cpuLimitMilliCPU: number;
  memoryRequestBytes: number;
  memoryLimitBytes: number;
  storageRequestBytes: number;
  storageLimitBytes: number;
  gpuRequest?: number;
  gpuLimit?: number;
}

export interface WorkerTemplateRuntime {
  runtimeImageId: number;
  placementPolicy: string;
  computeTargetRef: ResourceReference;
  deploymentMode: string;
  resourceProfileRef?: ResourceReference;
  customResources?: WorkerTemplateResources;
}

export interface WorkerTemplateTypeConfig {
  schemaVersion: number;
  values: Record<string, unknown>;
  secretRefs: Record<string, ResourceReference>;
  interactionMode: string;
  automationLevel: string;
}

export interface WorkerTemplateKnowledgeMount {
  ref: ResourceReference;
  mode: string;
}

export interface WorkerTemplateConfigDocumentBinding {
  documentId: string;
  configBundleRef: ResourceReference;
}

export interface WorkerTemplateWorkspace {
  repositoryRef?: ResourceReference;
  branch: string;
  skillRefs: ResourceReference[];
  knowledgeMounts: WorkerTemplateKnowledgeMount[];
  environmentBundleRefs: ResourceReference[];
  configDocumentBindings: WorkerTemplateConfigDocumentBinding[];
  instructions: string;
}

export interface WorkerTemplateSpec {
  optionsRevision: string;
  workerType: string;
  modelRef?: ResourceReference;
  toolRefs: Record<string, ResourceReference>;
  runtime: WorkerTemplateRuntime;
  typeConfig: WorkerTemplateTypeConfig;
  workspace: WorkerTemplateWorkspace;
  lifecycle: {
    terminationPolicy: string;
    idleTimeoutMinutes: number;
  };
  metadata: {
    alias: string;
  };
}

export type WorkerTemplateDraft = ResourceManifest<WorkerTemplateSpec> & {
  kind: "WorkerTemplate";
};

export interface WorkerInvocationSpec {
  workerTemplateRef: ResourceReference;
  promptRef?: ResourceReference;
  inputs: Record<string, string>;
  alias: string;
}

export type WorkerDraft = ResourceManifest<WorkerInvocationSpec> & {
  kind: "Worker";
};

export type WorkerResourceDraft = WorkerTemplateDraft | WorkerDraft;
export type WorkerResourceKind = WorkerResourceDraft["kind"];
