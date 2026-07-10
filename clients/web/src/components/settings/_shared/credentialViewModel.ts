export interface CredentialProfileViewModel {
  id: number;
  agent_slug: string;
  name: string;
  description?: string;
  is_default: boolean;
  is_active: boolean;
  configured_fields?: string[];
  configured_values?: Record<string, string>;
  agent_name?: string;
  created_at: string;
  updated_at: string;
}

export interface CredentialProfilesByAgent {
  agent_slug: string;
  agent_name: string;
  profiles: CredentialProfileViewModel[];
}

export function getConfiguredKeys(profile: {
  configured_fields?: string[];
  configured_values?: Record<string, string>;
}): string[] {
  return [
    ...new Set([
      ...(profile.configured_fields ?? []),
      ...Object.keys(profile.configured_values ?? {}),
    ]),
  ].sort();
}
