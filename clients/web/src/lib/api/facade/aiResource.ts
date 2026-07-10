import type {
  EffectiveResource as ProtoEffectiveResource,
  ModelResource as ProtoModelResource,
  ProviderConnection as ProtoProviderConnection,
  ProviderDefinition as ProtoProviderDefinition,
  UsageSummary as ProtoUsageSummary,
} from "@proto/ai_resource/v1/types_pb";

export interface ProviderDefinition {
  key: string;
  displayName: string;
  modalities: string[];
  credentialFields: Array<{ key: string; label: string; secret: boolean; required: boolean }>;
  defaultBaseUrl: string;
  protocolAdapter: string;
  supportsCustomEndpoint: boolean;
  supportsModelDiscovery: boolean;
}

export interface UsageSummary {
  quotaTotal?: number;
  usageTotal?: number;
  remaining?: number;
  unit?: string;
  period?: string;
  measuredAt?: string;
}

export interface ModelResource {
  id: number;
  providerConnectionId: number;
  identifier: string;
  modelId: string;
  displayName: string;
  modalities: string[];
  capabilities: string[];
  defaultModalities: string[];
  status: string;
  isEnabled: boolean;
  lastValidatedAt?: string;
  validationError: string;
  usageSummary?: UsageSummary;
}

export interface ProviderConnection {
  id: number;
  ownerScope: string;
  identifier: string;
  providerKey: string;
  name: string;
  baseUrl: string;
  configuredFields: string[];
  status: string;
  isEnabled: boolean;
  lastValidatedAt?: string;
  validationError: string;
  canManage: boolean;
  resources: ModelResource[];
}

export interface EffectiveResource {
  connection?: ProviderConnection;
  resource?: ModelResource;
  selectable: boolean;
  blockingReason: string;
}

export interface ConnectionInput {
  identifier: string;
  providerKey: string;
  name: string;
  baseUrl: string;
  credentials: Record<string, string>;
}

export interface ResourceInput {
  identifier: string;
  modelId: string;
  displayName: string;
  modalities: string[];
  capabilities: string[];
}

function safeId(value: bigint, name: string): number {
  const id = Number(value);
  if (!Number.isSafeInteger(id)) throw new Error(`unsafe ${name}`);
  return id;
}

export function fromProviderDefinition(value: ProtoProviderDefinition): ProviderDefinition {
  return {
    key: value.key, displayName: value.displayName, modalities: value.modalities,
    credentialFields: value.credentialFields.map((field) => ({
      key: field.key, label: field.label, secret: field.secret, required: field.required,
    })),
    defaultBaseUrl: value.defaultBaseUrl, protocolAdapter: value.protocolAdapter,
    supportsCustomEndpoint: value.supportsCustomEndpoint,
    supportsModelDiscovery: value.supportsModelDiscovery,
  };
}

function fromUsage(value: ProtoUsageSummary | undefined): UsageSummary | undefined {
  if (!value) return undefined;
  return {
    quotaTotal: value.quotaTotal,
    usageTotal: value.usageTotal,
    remaining: value.remaining,
    unit: value.unit,
    period: value.period,
    measuredAt: value.measuredAt,
  };
}

export function fromModelResource(value: ProtoModelResource): ModelResource {
  return {
    id: safeId(value.id, "model resource id"), providerConnectionId: safeId(value.providerConnectionId, "provider connection id"),
    identifier: value.identifier, modelId: value.modelId, displayName: value.displayName,
    modalities: value.modalities, capabilities: value.capabilities,
    defaultModalities: value.defaultModalities, status: value.status,
    isEnabled: value.isEnabled, lastValidatedAt: value.lastValidatedAt,
    validationError: value.validationError, usageSummary: fromUsage(value.usageSummary),
  };
}

export function fromProviderConnection(value: ProtoProviderConnection): ProviderConnection {
  return {
    id: safeId(value.id, "provider connection id"), ownerScope: value.ownerScope, identifier: value.identifier,
    providerKey: value.providerKey, name: value.name, baseUrl: value.baseUrl,
    configuredFields: value.configuredFields, status: value.status,
    isEnabled: value.isEnabled, lastValidatedAt: value.lastValidatedAt,
    validationError: value.validationError, canManage: value.canManage,
    resources: value.resources.map(fromModelResource),
  };
}

export function fromEffectiveResource(value: ProtoEffectiveResource): EffectiveResource {
  return {
    connection: value.connection ? fromProviderConnection(value.connection) : undefined,
    resource: value.resource ? fromModelResource(value.resource) : undefined,
    selectable: value.selectable, blockingReason: value.blockingReason,
  };
}
