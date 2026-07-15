export interface ResourceReferenceOption {
  name: string;
  displayName: string;
  revision: number;
}

export interface ResourceReferenceCatalog {
  loading: boolean;
  error: string | null;
  byKind: Record<string, ResourceReferenceOption[]>;
}
