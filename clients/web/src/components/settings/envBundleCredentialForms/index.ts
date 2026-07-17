import type { CredentialFormSpec, CredentialFieldSpec } from "./types";
import { buildCredentialFormSpec } from "./buildCredentialFormSpec";
import type { CredentialField } from "@/lib/viewModels/agent";

export function getCredentialFormSpecFromFields(
  agentSlug: string,
  credentialFields: CredentialField[]
): CredentialFormSpec {
  return buildCredentialFormSpec(agentSlug, credentialFields);
}

export function getEnvKeysFromSpec(spec: CredentialFormSpec): Set<string> {
  const keys = new Set<string>();
  for (const field of spec.fields) {
    if (field.kind === "oneof") {
      for (const opt of field.options) keys.add(opt.envKey);
    } else {
      keys.add(field.envKey);
    }
  }
  return keys;
}

export function findFieldByEnvKey(
  spec: CredentialFormSpec,
  envKey: string
): CredentialFieldSpec | undefined {
  for (const field of spec.fields) {
    if (field.kind === "oneof") {
      if (field.options.some((o) => o.envKey === envKey)) return field;
    } else if (field.envKey === envKey) {
      return field;
    }
  }
  return undefined;
}

export type { CredentialFormSpec, CredentialFieldSpec, CustomEnvEntry } from "./types";
