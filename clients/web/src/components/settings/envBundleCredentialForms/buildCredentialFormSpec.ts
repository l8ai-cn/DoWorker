import type { CredentialField } from "@/lib/viewModels/agent";
import type {
  CredentialFieldSpec,
  CredentialFormSpec,
  OneOfCredentialField,
  SimpleCredentialField,
} from "./types";
import { getCredentialUxOverride } from "./credentialUxOverrides";

function defaultLabelKey(envKey: string): string {
  if (envKey === "ANTHROPIC_API_KEY" || envKey === "ANTHROPIC_AUTH_TOKEN") {
    return "settings.credentialForm.anthropic.apiKey";
  }
  if (envKey === "OPENAI_API_KEY") return "settings.credentialForm.openai.apiKey";
  if (envKey === "GOOGLE_API_KEY") return "settings.credentialForm.google.apiKey";
  if (envKey === "CURSOR_API_KEY") return "settings.credentialForm.cursor.apiKey";
  return envKey;
}

function toSimpleField(
  field: CredentialField,
  override: ReturnType<typeof getCredentialUxOverride>
): SimpleCredentialField {
  return {
    kind: field.type === "secret" ? "secret" : "text",
    envKey: field.name,
    label: override?.labels?.[field.name] ?? defaultLabelKey(field.name),
    description: override?.descriptions?.[field.name],
    placeholder: override?.placeholders?.[field.name],
    securityHint: override?.securityHints?.[field.name],
  };
}

function buildOneOfFields(
  override: NonNullable<ReturnType<typeof getCredentialUxOverride>>,
  byName: Map<string, CredentialField>
): { fields: CredentialFieldSpec[]; consumed: Set<string> } {
  const consumed = new Set<string>();
  const fields: CredentialFieldSpec[] = [];

  for (const group of override.oneofGroups ?? []) {
    const options = group.envKeys
      .filter((key) => byName.has(key))
      .map((key) => {
        consumed.add(key);
        const src = byName.get(key)!;
        return {
          kind: src.type === "secret" ? "secret" : "text",
          envKey: key,
          label: override.labels?.[key] ?? defaultLabelKey(key),
          placeholder: override.placeholders?.[key],
        } as OneOfCredentialField["options"][number];
      });
    if (options.length === 0) continue;
    fields.push({
      kind: "oneof",
      group: group.group,
      label: group.label,
      description: group.description,
      options,
    });
  }
  return { fields, consumed };
}

// Merges API-derived credential_fields (AgentFile SSOT) with per-agent UX
// overrides for field ordering and i18n labels.
export function buildCredentialFormSpec(
  agentSlug: string,
  credentialFields: CredentialField[]
): CredentialFormSpec {
  const override = getCredentialUxOverride(agentSlug);
  const byName = new Map(credentialFields.map((f) => [f.name, f]));

  const oneOfResult = override
    ? buildOneOfFields(override, byName)
    : { fields: [] as CredentialFieldSpec[], consumed: new Set<string>() };

  const simpleFields: CredentialFieldSpec[] = [];
  const orderedKeys = override?.fieldOrder ?? credentialFields.map((f) => f.name);

  for (const key of orderedKeys) {
    if (oneOfResult.consumed.has(key)) continue;
    const src = byName.get(key);
    if (!src) continue;
    simpleFields.push(toSimpleField(src, override));
  }

  for (const src of credentialFields) {
    if (oneOfResult.consumed.has(src.name) || orderedKeys.includes(src.name)) continue;
    simpleFields.push(toSimpleField(src, override));
  }

  return {
    agentSlug,
    fields: [...simpleFields, ...oneOfResult.fields],
  };
}
