export interface CredentialOneOfOverride {
  group: string;
  label: string;
  description?: string;
  envKeys: string[];
}

export interface CredentialUxOverride {
  fieldOrder?: string[];
  oneofGroups?: CredentialOneOfOverride[];
  labels?: Record<string, string>;
  descriptions?: Record<string, string>;
  placeholders?: Record<string, string>;
  securityHints?: Record<string, string>;
}

const OVERRIDES: Record<string, CredentialUxOverride> = {
  "claude-code": {
    fieldOrder: ["ANTHROPIC_BASE_URL", "ANTHROPIC_API_KEY"],
    labels: {
      ANTHROPIC_BASE_URL: "settings.credentialForm.anthropic.baseUrl",
      ANTHROPIC_API_KEY: "settings.credentialForm.anthropic.apiKey",
    },
    placeholders: {
      ANTHROPIC_BASE_URL: "https://api.anthropic.com",
      ANTHROPIC_API_KEY: "sk-ant-...",
    },
    securityHints: {
      ANTHROPIC_BASE_URL: "settings.agentCredentials.baseUrlSecurityHint",
    },
  },
  "codex-cli": {
    labels: { OPENAI_API_KEY: "settings.credentialForm.openai.apiKey" },
    placeholders: { OPENAI_API_KEY: "sk-..." },
  },
  loopal: {
    labels: {
      ANTHROPIC_API_KEY: "settings.credentialForm.anthropic.apiKey",
      OPENAI_API_KEY: "settings.credentialForm.openai.apiKey",
      GOOGLE_API_KEY: "settings.credentialForm.google.apiKey",
    },
    placeholders: {
      ANTHROPIC_API_KEY: "sk-ant-...",
      OPENAI_API_KEY: "sk-...",
    },
  },
  "gemini-cli": {
    labels: { GEMINI_API_KEY: "settings.credentialForm.google.geminiApiKey" },
    placeholders: { GEMINI_API_KEY: "AIza..." },
  },
  aider: {
    labels: {
      OPENAI_API_KEY: "settings.credentialForm.openai.apiKey",
      ANTHROPIC_API_KEY: "settings.credentialForm.anthropic.apiKey",
    },
  },
  opencode: {},
  "cursor-cli": {
    labels: { CURSOR_API_KEY: "settings.credentialForm.cursor.apiKey" },
    descriptions: { CURSOR_API_KEY: "settings.credentialForm.cursor.apiKeyHint" },
  },
  "do-agent": {
    labels: {
      OPENAI_API_KEY: "settings.credentialForm.openai.apiKey",
      ANTHROPIC_API_KEY: "settings.credentialForm.anthropic.apiKey",
    },
    placeholders: {
      OPENAI_API_KEY: "sk-...",
      ANTHROPIC_API_KEY: "sk-ant-...",
    },
  },
  "grok-build": {
    labels: { XAI_API_KEY: "settings.credentialForm.xai.apiKey" },
    placeholders: { XAI_API_KEY: "xai-..." },
  },
  openclaw: {
    labels: { OPENAI_API_KEY: "settings.credentialForm.openai.apiKey" },
    placeholders: { OPENAI_API_KEY: "sk-..." },
  },
  hermes: {
    labels: { OPENAI_API_KEY: "settings.credentialForm.openai.apiKey" },
    placeholders: { OPENAI_API_KEY: "sk-..." },
  },
  "e2e-echo": {
    labels: { E2E_TEST_CRED_KEY: "E2E Test Credential Key" },
    descriptions: {
      E2E_TEST_CRED_KEY: "Internal stub — used only by EnvBundle end-to-end tests.",
    },
    placeholders: { E2E_TEST_CRED_KEY: "Test value (not a real credential)" },
  },
};

export function getCredentialUxOverride(agentSlug: string): CredentialUxOverride | undefined {
  return OVERRIDES[agentSlug];
}
