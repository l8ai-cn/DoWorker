import {
  RESOURCE_API_VERSION,
  type WorkerDraft,
} from "./resource-editor-types";

export function createWorkerInvocationDraft(namespace: string): WorkerDraft {
  return {
    apiVersion: RESOURCE_API_VERSION,
    kind: "Worker",
    metadata: {
      name: "",
      namespace,
      displayName: "",
      labels: {},
    },
    spec: {
      workerTemplateRef: {
        kind: "WorkerTemplate",
        name: "",
      },
      inputs: {},
      alias: "",
    },
  };
}
