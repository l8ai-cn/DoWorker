export const RESOURCE_API_VERSION = "agentcloud.io/v1alpha1";

export interface ResourceReference {
  apiVersion?: string;
  kind: string;
  namespace?: string;
  name: string;
  revision?: number;
}

export interface ResourceMetadata {
  name: string;
  namespace: string;
  displayName?: string;
  labels?: Record<string, string>;
}

export interface ResourceManifest<TSpec> {
  apiVersion: string;
  kind: string;
  metadata: ResourceMetadata;
  spec: TSpec;
}
