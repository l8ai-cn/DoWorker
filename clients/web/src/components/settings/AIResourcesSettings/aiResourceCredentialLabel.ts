import type { ProviderDefinition } from "./types";

const credentialLabelKeys: Record<string, string> = {
  api_key: "settings.aiResources.connection.credentials.apiKey",
  api_token: "settings.aiResources.connection.credentials.apiToken",
  access_key: "settings.aiResources.connection.credentials.accessKey",
  secret_key: "settings.aiResources.connection.credentials.secretKey",
  subscription_key: "settings.aiResources.connection.credentials.subscriptionKey",
  region: "settings.aiResources.connection.credentials.region",
};

export function getAIResourceCredentialLabel(
  field: ProviderDefinition["credentialFields"][number],
  translate: (key: string) => string,
) {
  const translationKey = credentialLabelKeys[field.key];
  return translationKey ? translate(translationKey) : field.label;
}
