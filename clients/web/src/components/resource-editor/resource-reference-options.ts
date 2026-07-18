export interface ResourceReferenceOption {
  name: string;
  displayName: string;
  revision: number;
}

export type EnvironmentBundleReferencePurpose =
  | "runtime"
  | "config"
  | "credential";

export function environmentBundleCatalogKey(
  purpose: EnvironmentBundleReferencePurpose,
  targetName?: string,
): string {
  return targetName
    ? `EnvironmentBundle:${purpose}:${targetName}`
    : `EnvironmentBundle:${purpose}`;
}

export interface ResourceReferenceCatalog {
  loading: boolean;
  error: string | null;
  errorsByKind: Record<string, string>;
  byKind: Record<string, ResourceReferenceOption[]>;
}

export function isResourceReferenceCatalogReadOnly(
  catalog: ResourceReferenceCatalog,
  key: string,
  names: string[] = [],
): boolean {
  const options = catalog.byKind[key] ?? [];
  const error = catalog.errorsByKind[key] ?? catalog.error;
  return catalog.loading ||
    Boolean(error) ||
    options.length === 0 ||
    names.some((name) =>
      Boolean(name) && !options.some((option) => option.name === name)
    );
}
