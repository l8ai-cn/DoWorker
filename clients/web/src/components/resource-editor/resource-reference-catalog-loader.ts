import { EnvironmentBundlePurpose } from "@proto/orchestration_resource/v1/orchestration_resource_pb";
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
  modelProtocolAdapters: readonly string[],
  credentialTargetNames: readonly string[],
): Promise<ResourceReferenceCatalog> {
  const results = await Promise.all(referenceQueries(
    workerType,
    [...new Set(modelProtocolAdapters)].sort(),
    [...new Set(credentialTargetNames)].sort(),
  ).map(async (query) => {
    try {
      const options = await collectResourceReferenceOptions(
        orgSlug,
        query,
        toOption,
        assertReferenceFiltersApplied,
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

function referenceQueries(
  workerType: string,
  modelProtocolAdapters: string[],
  credentialTargets: string[],
) {
  const common = REFERENCE_KINDS.map((kind) => ({
    key: kind,
    kind,
    environmentBundleFilter: undefined,
    modelBindingFilter: kind === "ModelBinding" && workerType &&
        modelProtocolAdapters.length > 0
      ? { workerType, protocolAdapters: modelProtocolAdapters }
      : undefined,
  }));
  if (!workerType) return common;
  return [
    ...common,
    ...ENVIRONMENT_BUNDLE_PURPOSES.map(({ purpose, wire }) => ({
      key: environmentBundleCatalogKey(purpose),
      kind: "EnvironmentBundle" as const,
      environmentBundleFilter: { purpose: wire, workerType },
      modelBindingFilter: undefined,
    })),
    ...credentialTargets.map((targetName) => ({
      key: environmentBundleCatalogKey("credential", targetName),
      kind: "EnvironmentBundle" as const,
      environmentBundleFilter: {
        purpose: EnvironmentBundlePurpose.CREDENTIAL,
        workerType,
        targetName,
      },
      modelBindingFilter: undefined,
    })),
  ];
}

function assertReferenceFiltersApplied(
  requested: {
    environmentBundleFilter?: {
      purpose: EnvironmentBundlePurpose;
      workerType: string;
      targetName?: string;
    };
    modelBindingFilter?: {
      workerType: string;
      protocolAdapters: readonly string[];
    };
  },
  applied: {
    appliedEnvironmentBundleFilter?: {
    purpose: EnvironmentBundlePurpose;
    workerType: string;
    targetName: string;
    };
    appliedModelBindingFilter?: {
      workerType: string;
      protocolAdapters: readonly string[];
    };
  },
) {
  const environment = requested.environmentBundleFilter;
  const appliedEnvironment = applied.appliedEnvironmentBundleFilter;
  const model = requested.modelBindingFilter;
  const appliedModel = applied.appliedModelBindingFilter;
  if (environment && (
    appliedEnvironment?.purpose !== environment.purpose ||
    appliedEnvironment.workerType !== environment.workerType ||
    appliedEnvironment.targetName !== (environment.targetName ?? "")
  )) {
    throw new Error(
      "The control plane did not apply the EnvironmentBundle reference filter.",
    );
  }
  if (model && (
    !appliedModel ||
    appliedModel.workerType !== model.workerType ||
    JSON.stringify([...appliedModel.protocolAdapters].sort()) !==
      JSON.stringify([...model.protocolAdapters].sort())
  )) {
    throw new Error(
      "The control plane did not apply the ModelBinding protocol filter.",
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
