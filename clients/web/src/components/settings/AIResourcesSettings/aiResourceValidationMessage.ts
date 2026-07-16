type Translate = (key: string) => string;

const knownValidationErrors: Record<string, string> = {
  "credentials rejected": "settings.aiResources.validation.credentialsRejected",
  "invalid AI resource credentials": "settings.aiResources.validation.credentialsRejected",
};

export function aiResourceValidationMessage(error: string, t: Translate): string {
  return knownValidationErrors[error] ? t(knownValidationErrors[error]) : error;
}
