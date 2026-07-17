import type {
  ResourceManifest,
  ResourceReference,
} from "./resource-manifest-types";

export const RESOURCE_ID_BINDING_FIELDS = {
  ModelBinding: "resourceId",
  Repository: "repositoryId",
  Skill: "skillId",
  KnowledgeBase: "knowledgeBaseId",
  EnvironmentBundle: "environmentBundleId",
  ComputeTarget: "computeTargetId",
  ResourceProfile: "resourceProfileId",
} as const;

export type ResourceIDBindingKind = keyof typeof RESOURCE_ID_BINDING_FIELDS;
export type ResourceBindingKind = ResourceIDBindingKind | "ToolBinding";

export type ModelBindingDraft = ResourceManifest<{ resourceId: number }> & {
  kind: "ModelBinding";
};
export type RepositoryBindingDraft =
  ResourceManifest<{ repositoryId: number }> & { kind: "Repository" };
export type SkillBindingDraft = ResourceManifest<{ skillId: number }> & {
  kind: "Skill";
};
export type KnowledgeBaseBindingDraft =
  ResourceManifest<{ knowledgeBaseId: number }> & { kind: "KnowledgeBase" };
export type EnvironmentBundleBindingDraft =
  ResourceManifest<{ environmentBundleId: number }> & {
    kind: "EnvironmentBundle";
  };
export type ComputeTargetBindingDraft =
  ResourceManifest<{ computeTargetId: number }> & { kind: "ComputeTarget" };
export type ResourceProfileBindingDraft =
  ResourceManifest<{ resourceProfileId: number }> & {
    kind: "ResourceProfile";
  };
export type ToolBindingDraft =
  ResourceManifest<{ modelRef: ResourceReference }> & { kind: "ToolBinding" };

export type ResourceBindingDraft =
  | ModelBindingDraft
  | RepositoryBindingDraft
  | SkillBindingDraft
  | KnowledgeBaseBindingDraft
  | EnvironmentBundleBindingDraft
  | ComputeTargetBindingDraft
  | ResourceProfileBindingDraft
  | ToolBindingDraft;

export function isResourceBindingKind(
  kind: string,
): kind is ResourceBindingKind {
  return kind === "ToolBinding" || kind in RESOURCE_ID_BINDING_FIELDS;
}
