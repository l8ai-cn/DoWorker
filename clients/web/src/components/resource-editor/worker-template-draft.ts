import {
  RESOURCE_API_VERSION,
  type WorkerTemplateDraft,
} from "./resource-editor-types";

export function createWorkerTemplateDraft(
  namespace: string,
): WorkerTemplateDraft {
  return {
    apiVersion: RESOURCE_API_VERSION,
    kind: "WorkerTemplate",
    metadata: {
      name: "",
      namespace,
      displayName: "",
      labels: {},
    },
    spec: {
      optionsRevision: "",
      workerType: "",
      toolRefs: {},
      runtime: {
        runtimeImageId: 0,
        placementPolicy: "automatic",
        computeTargetRef: { kind: "ComputeTarget", name: "" },
        deploymentMode: "pooled",
      },
      typeConfig: {
        schemaVersion: 1,
        values: {},
        secretRefs: {},
        interactionMode: "acp",
        automationLevel: "autonomous",
      },
      workspace: {
        branch: "",
        skillRefs: [],
        knowledgeMounts: [],
        environmentBundleRefs: [],
        configBundleRefs: [],
        instructions: "",
      },
      lifecycle: {
        terminationPolicy: "manual",
        idleTimeoutMinutes: 0,
      },
      metadata: {
        alias: "",
      },
    },
  };
}
