import {
  RESOURCE_API_VERSION,
  type ResourceBindingDraft,
  type ResourceBindingKind,
  type ResourceManifest,
} from "./resource-editor-types";

export function createResourceBindingDraft(
  kind: ResourceBindingKind,
  namespace: string,
): ResourceBindingDraft {
  switch (kind) {
    case "ModelBinding":
      return bindingManifest(kind, namespace, { resourceId: 0 });
    case "Repository":
      return bindingManifest(kind, namespace, { repositoryId: 0 });
    case "Skill":
      return bindingManifest(kind, namespace, { skillId: 0 });
    case "KnowledgeBase":
      return bindingManifest(kind, namespace, { knowledgeBaseId: 0 });
    case "EnvironmentBundle":
      return bindingManifest(kind, namespace, { environmentBundleId: 0 });
    case "ComputeTarget":
      return bindingManifest(kind, namespace, { computeTargetId: 0 });
    case "ResourceProfile":
      return bindingManifest(kind, namespace, { resourceProfileId: 0 });
    case "ToolBinding":
      return bindingManifest(kind, namespace, {
        modelRef: { kind: "ModelBinding", name: "" },
      });
  }
}

function bindingManifest<K extends ResourceBindingKind, S>(
  kind: K,
  namespace: string,
  spec: S,
): ResourceManifest<S> & { kind: K } {
  return {
    apiVersion: RESOURCE_API_VERSION,
    kind,
    metadata: {
      name: "",
      namespace,
      displayName: "",
      labels: {},
    },
    spec,
  };
}
