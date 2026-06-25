import type { EnvBundle as ProtoEnvBundle } from "@proto/env_bundle/v1/env_bundle_pb";

import type { CredentialProfileViewModel } from "../_shared/credentialViewModel";
import type { ConfigFileBundleViewModel, RuntimeBundleViewModel } from "./types";
import { CONFIG_BUNDLE_JSON_KEY } from "./configBundleKeys";

export function toCredentialProfile(
  b: ProtoEnvBundle,
  fallbackAgentSlug: string,
): CredentialProfileViewModel {
  return {
    id: Number(b.id),
    agent_slug: b.agentSlug ?? fallbackAgentSlug,
    name: b.name,
    description: b.description ?? undefined,
    is_default: b.kindPrimary,
    is_active: b.isActive,
    configured_fields: b.configuredFields.length > 0 ? b.configuredFields : undefined,
    configured_values:
      Object.keys(b.configuredValues).length > 0 ? b.configuredValues : undefined,
    created_at: b.createdAt,
    updated_at: b.updatedAt,
  };
}

export function toRuntimeBundle(
  b: ProtoEnvBundle,
  fallbackAgentSlug: string,
): RuntimeBundleViewModel {
  return {
    id: Number(b.id),
    agent_slug: b.agentSlug ?? fallbackAgentSlug,
    name: b.name,
    description: b.description ?? undefined,
    is_default: b.kindPrimary,
    is_active: b.isActive,
    configured_fields: b.configuredFields.length > 0 ? b.configuredFields : undefined,
    configured_values:
      Object.keys(b.configuredValues).length > 0 ? b.configuredValues : undefined,
    created_at: b.createdAt,
    updated_at: b.updatedAt,
  };
}

export function toConfigFileBundle(
  b: ProtoEnvBundle,
  fallbackAgentSlug: string,
): ConfigFileBundleViewModel {
  const json = b.configuredValues[CONFIG_BUNDLE_JSON_KEY];
  return {
    id: Number(b.id),
    agent_slug: b.agentSlug ?? fallbackAgentSlug,
    name: b.name,
    description: b.description ?? undefined,
    is_default: b.kindPrimary,
    is_active: b.isActive,
    json_content: json ?? undefined,
    created_at: b.createdAt,
    updated_at: b.updatedAt,
  };
}
