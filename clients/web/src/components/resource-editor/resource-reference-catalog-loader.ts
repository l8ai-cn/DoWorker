import { EnvironmentBundlePurpose } from "@proto/orchestration_resource/v1/orchestration_resource_queries_pb";
import type {
  EnvironmentBundleReferencePurpose,
  ResourceReferenceCatalog,
  ResourceReferenceOption,
} from "./resource-reference-options";
import { environmentBundleCatalogKey } from "./resource-reference-options";
import { collectResourceReferenceOptions } from "./resource-reference-options-pagination";
import { safeResourceError } from "./resource-editor-source-transition";

const REFERENCE_KINDS = [
  "WorkerTemplate",
  "Prompt",
  "ModelBinding",
  "ToolBinding",
  "Repository",
  "Skill",
  "KnowledgeBase",
  "ComputeTarget",
  "ResourceProfile",
] as const;

const ENVIRONMENT_BUNDLE_PURPOSES = [
  { purpose: "runtime", wire: EnvironmentBundlePurpose.RUNTIME },
  { purpose: "config", wire: EnvironmentBundlePurpose.CONFIG },
] as const satisfies ReadonlyArray<{
  purpose: EnvironmentBundleReferencePurpose;
  wire: EnvironmentBundlePurpose;
}>;

export async function loadResourceReferenceCatalog(
  orgSlug: string,
  workerType: string,
  credentialTargetNames: readonly string[],
): Promise<ResourceReferenceCatalog> {
  const results = await Promise.all(referenceQueries(
    workerType,
    [...new Set(credentialTargetNames)].sort(),
  ).map(async (query) => {
    try {
      const options = await collectResourceReferenceOptions(
        orgSlug,
        query,
        toOption,
        assertEnvironmentBundleFilterApplied,
      );
      return { ok: true, key: query.key, options } as const;
    } catch (error) {
      return {
        ok: false,
        key: query.key,
        error: safeResourceError(
          error,
          `Failed to load ${query.kind} references.`,
        ),
      } as const;
    }
  }));
  const byKind: Record<string, ResourceReferenceOption[]> = {};
  const errorsByKind: Record<string, string> = {};
  for (const result of results) {
    if (result.ok) byKind[result.key] = result.options;
    else errorsByKind[result.key] = result.error;
  }
  return {
    loading: false,
    error: Object.keys(byKind).length === 0
      ? Object.values(errorsByKind)[0] ?? "Failed to load resource references."
      : null,
    errorsByKind,
    byKind,
  };
}

function referenceQueries(workerType: string, credentialTargets: string[]) {
  const common = REFERENCE_KINDS.map((kind) => ({
    key: kind,
    kind,
    environmentBundleFilter: undefined,
  }));
  if (!workerType) return common;
  return [
    ...common,
    ...ENVIRONMENT_BUNDLE_PURPOSES.map(({ purpose, wire }) => ({
      key: environmentBundleCatalogKey(purpose),
      kind: "EnvironmentBundle" as const,
      environmentBundleFilter: { purpose: wire, workerType },
    })),
    ...credentialTargets.map((targetName) => ({
      key: environmentBundleCatalogKey("credential", targetName),
      kind: "EnvironmentBundle" as const,
      environmentBundleFilter: {
        purpose: EnvironmentBundlePurpose.CREDENTIAL,
        workerType,
        targetName,
      },
    })),
  ];
}

function assertEnvironmentBundleFilterApplied(
  requested: {
    purpose: EnvironmentBundlePurpose;
    workerType: string;
    targetName?: string;
  } | undefined,
  applied: {
    purpose: EnvironmentBundlePurpose;
    workerType: string;
    targetName: string;
  } | undefined,
) {
  if (!requested) return;
  if (
    applied?.purpose !== requested.purpose ||
    applied.workerType !== requested.workerType ||
    applied.targetName !== (requested.targetName ?? "")
  ) {
    throw new Error(
      "The control plane did not apply the EnvironmentBundle reference filter.",
    );
  }
}

function toOption(resource: {
  identity?: { target?: { name?: string } };
  displayName: string;
  revision: bigint;
}): ResourceReferenceOption {
  return {
    name: resource.identity?.target?.name ?? "",
    displayName: resource.displayName,
    revision: Number(resource.revision),
  };
}
