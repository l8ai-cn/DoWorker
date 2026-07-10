import type { CredentialFormSpec, CredentialFieldSpec } from "./types";
import { buildCredentialFormSpec } from "./buildCredentialFormSpec";
import { getBuiltinCredentialFallback } from "./credentialBuiltinFallbacks";
import type { CredentialField } from "@/lib/viewModels/agent";

// e2e-echo fallback is registered only in E2E builds.
if (process.env.NEXT_PUBLIC_E2E === "true") {
  // eslint-disable-next-line @typescript-eslint/no-require-imports, no-restricted-imports
  require("./e2e-echo-fallback");
}

function makeFallback(agentSlug: string): CredentialFormSpec {
  return buildCredentialFormSpec(agentSlug, []);
}

export function getCredentialFormSpec(agentSlug: string): CredentialFormSpec {
  const fallback = getBuiltinCredentialFallback(agentSlug);
  if (fallback.length === 0) {
    return makeFallback(agentSlug);
  }
  return buildCredentialFormSpec(agentSlug, fallback);
}

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

export function getEnvKeyLabel(
  agentSlug: string,
  envKey: string,
  t: (key: string) => string
): string {
  const spec = getCredentialFormSpec(agentSlug);
  for (const field of spec.fields) {
    if (field.kind === "oneof") {
      const opt = field.options.find((o) => o.envKey === envKey);
      if (opt) {
        const translated = t(opt.label);
        return translated === opt.label ? envKey : translated;
      }
    } else if (field.envKey === envKey) {
      const translated = t(field.label);
      return translated === field.label ? envKey : translated;
    }
  }
  return envKey;
}

export type { CredentialFormSpec, CredentialFieldSpec, CustomEnvEntry } from "./types";
